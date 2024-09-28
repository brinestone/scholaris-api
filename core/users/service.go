package users

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

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

// Fetches a paginated set of Users
//
//encore:api auth method=GET path=/users
func FetchUsers(ctx context.Context, req dto.PaginationParams) (*FetchUsersResponse, error) {
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

	log.Printf("%+v\n", user)

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
func NewUser(ctx context.Context, req dto.NewUserRequest) (*models.User, error) {
	tx, err := userDb.Begin(ctx)
	if err != nil {
		rlog.Error(err.Error())
		return nil, &util.ErrUnknown
	}

	user, err := createUser(ctx, req)
	if err != nil {
		_ = tx.Rollback()
		rlog.Error(err.Error())
		return nil, &util.ErrUnknown
	}

	defer tx.Commit()
	return user, nil
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

func findUserByEmailFromCache(ctx context.Context, email string) (*models.User, error) {
	u, err := emailCache.Get(ctx, email)
	if err != nil {
		return nil, err
	}
	return &u, err
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
	rlog.Debug(query)
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

func createUser(ctx context.Context, req dto.NewUserRequest) (*models.User, error) {
	existingUser, err := findUserByEmailFromDb(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	if existingUser != nil {
		return nil, &errs.Error{
			Code:    errs.AlreadyExists,
			Message: "email is already in use",
		}
	}

	dob, _ := time.Parse("2006/2/1", req.Dob)
	ph, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	query := "INSERT INTO users (first_name, last_name, email, dob, password_hash, phone, gender) VALUES ($1,$2,$3,$4,$5,$6,$7);"
	_, err = userDb.Exec(ctx, query, req.FirstName, req.LastName, req.Email, dob, string(ph), req.Phone, req.Gender)
	if err != nil {
		rlog.Error(err.Error())
		return nil, &util.ErrUnknown
	}

	return findUserByEmailFromDb(ctx, req.Email)
}
