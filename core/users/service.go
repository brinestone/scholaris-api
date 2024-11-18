package users

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/rlog"
	"encore.dev/storage/cache"
	"encore.dev/storage/sqldb"
	"github.com/brinestone/scholaris/dto"
	"github.com/brinestone/scholaris/models"
	"github.com/brinestone/scholaris/util"
	"golang.org/x/crypto/bcrypt"
)

type FetchUsersResponse struct {
	Users []*models.User `json:"users"`
}

// Uploads a user's profile photo
//
//encore:api raw auth path=/avatars/:id
func UploadProfilePhoto(w http.ResponseWriter, req *http.Request) {
	userId, _ := auth.UserID()
	key := util.HashThese(string(userId), time.Now().String())

	upload := profilePhotoUploads.Upload(req.Context(), key)
	_, err := io.Copy(upload, req.Body)
	if err != nil {
		upload.Abort(err)
		rlog.Error(util.MsgUploadError, "msg", err.Error())
		errs.HTTPError(w, &util.ErrUnknown)
		return
	}

	if err := upload.Close(); err != nil {
		rlog.Error(util.MsgUploadError, "msg", err.Error())
		errs.HTTPError(w, &util.ErrUnknown)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Fetches a paginated set of Users
//
//encore:api auth method=GET path=/users
func FetchUsers(ctx context.Context, req dto.CursorBasedPaginationParams) (*FetchUsersResponse, error) {
	ans, err := findAllUsers(ctx, req.After, req.Size)
	if err != nil {
		return nil, err
	}

	return &FetchUsersResponse{
		Users: ans,
	}, nil
}

// Verifies a whether a user's credentials are valid and returns relevant fields
//
//encore:api private method=POST path=/users/verify
func VerifyCredentials(ctx context.Context, req dto.LoginRequest) (*models.User, error) {
	var user *models.User
	var err error

	// user, err = findUserByEmailFromCache(ctx, req.Email)
	user, err = findUserByEmailFromDb(ctx, req.Email)
	if err != nil {
		if errors.Is(err, cache.Miss) {
			rlog.Warn("cache miss", "email", req.Email)
		} else {
			return nil, err
		}
	}

	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, &errs.Error{
			Code:    errs.NotFound,
			Message: "Account not found",
		}
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, &errs.Error{
			Code:    errs.Unauthenticated,
			Message: "Invalid email or password",
		}
	}

	return user, nil
}

// Create a new user account
//
//encore:api private method=POST path=/users
func NewUser(ctx context.Context, req dto.NewUserRequest) (ans *dto.NewUserResponse, err error) {
	emailExists, err := userEmailExists(ctx, req.Email)
	if err != nil {
		return
	}

	if emailExists {
		err = &errs.Error{
			Code:    errs.AlreadyExists,
			Message: "email is already in use",
		}
	}

	tx, err := userDb.Begin(ctx)
	if err != nil {
		rlog.Error(err.Error())
		err = &util.ErrUnknown
	}

	uid, err := createUser(ctx, req, tx)
	if err != nil {
		_ = tx.Rollback()
		rlog.Error(err.Error())
		err = &util.ErrUnknown
	}

	tx.Commit()

	ans = &dto.NewUserResponse{
		UserId: uid,
	}
	return
}

// Find a user by their ID
//
//encore:api private method=GET path=/users/:id
func FindUserById(ctx context.Context, id uint64) (*models.User, error) {
	var user *models.User
	var err error

	if user, err = findUserByIdFromCache(ctx, id); user == nil {
		if errors.Is(err, cache.Miss) {
			rlog.Warn("cache miss", "id", id)
			user, err = findUserByIdFromDb(ctx, id)
		}
	}

	return user, err
}

func findUserByIdFromCache(ctx context.Context, id uint64) (*models.User, error) {
	u, err := idCache.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return &u, err
}

func findUserByIdFromDb(ctx context.Context, id uint64) (*models.User, error) {
	query := fmt.Sprintf("SELECT %s FROM users where id = $1;", allUserFields)
	var ans = new(models.User)

	row := userDb.QueryRow(ctx, query, id)
	if err := row.Scan(&ans.Id, &ans.FirstName, &ans.LastName, &ans.Email, &ans.Dob, &ans.PasswordHash, &ans.Phone, &ans.CreatedAt, &ans.UpdatedAt, &ans.Gender, &ans.Avatar); err != nil {
		if errors.Is(err, sqldb.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	idCache.Set(ctx, id, *ans)
	return ans, nil
}

func userEmailExists(ctx context.Context, email string) (ans bool, err error) {
	query := `
		SELECT 
			COUNT(id) 
		FROM 
			users
		WHERE 
			email = $1
		;
	`

	var cnt = 0
	if err = userDb.QueryRow(ctx, query, email).Scan(&cnt); err != nil {
		return
	}
	ans = cnt > 0

	return
}

func findUserByEmailFromDb(ctx context.Context, email string) (*models.User, error) {
	query := fmt.Sprintf("SELECT %s FROM users WHERE email = $1;", allUserFields)
	var ans *models.User = new(models.User)

	row := userDb.QueryRow(ctx, query, email)

	if err := row.Scan(&ans.Id, &ans.FirstName, &ans.LastName, &ans.Email, &ans.Dob, &ans.PasswordHash, &ans.Phone, &ans.CreatedAt, &ans.UpdatedAt, &ans.Gender, &ans.Avatar); err != nil {
		if errors.Is(err, sqldb.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	emailCache.Set(ctx, email, *ans)
	return ans, nil
}

const allUserFields = "id,first_name,last_name,email,dob,password_hash,phone,created_at,updated_at,gender,avatar"

func findAllUsers(ctx context.Context, offset uint64, size uint) ([]*models.User, error) {
	query := fmt.Sprintf("SELECT %s FROM users WHERE id > $1 ORDER BY id DESC OFFSET 0 LIMIT $2;", allUserFields)
	ans := make([]*models.User, 0)

	rows, err := userDb.Query(ctx, query, offset, size)
	if err != nil {
		return ans, err
	}
	defer rows.Close()

	for rows.Next() {
		user := new(models.User)
		if err := rows.Scan(&user.Id, &user.FirstName, &user.LastName, &user.Email, &user.Dob, &user.PasswordHash, &user.Phone, &user.CreatedAt, &user.UpdatedAt, &user.Gender, &user.Avatar); err != nil {
			if errors.Is(err, sqldb.ErrNoRows) {
				break
			}
			return ans, err
		}
		ans = append(ans, user)
		idCache.Set(ctx, user.Id, *user)
	}

	return ans, nil
}

func createUser(ctx context.Context, req dto.NewUserRequest, tx *sqldb.Tx) (ans uint64, err error) {

	dob, _ := time.Parse("2006/2/1", req.Dob)
	ph, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	avatar := fmt.Sprintf("https://api.dicebear.com/9.x/adventurer/svg?seed?=%s&scale=80", url.QueryEscape(strings.Trim(fmt.Sprintf("%s %s", req.FirstName, req.LastName), " \n\t")))

	query := `
		INSERT INTO 
			users (first_name, last_name, email, dob, password_hash, phone, gender, avatar) 
		VALUES 
			($1,$2,$3,$4,$5,$6,$7,$8) 
		RETURNING
			id;
	`

	if err = tx.QueryRow(ctx, query, req.FirstName, req.LastName, req.Email, dob, string(ph), req.Phone, req.Gender, avatar).Scan(&ans); err != nil {
		rlog.Error(err.Error())
		err = &util.ErrUnknown
	}

	return
}
