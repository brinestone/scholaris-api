package users

import (
	"context"
	"database/sql"
	"encoding/json"
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
	"github.com/brinestone/scholaris/helpers"
	"github.com/brinestone/scholaris/models"
	"github.com/brinestone/scholaris/util"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// Creates a new user with an externally provided account
//
//encore:api private method=POST path=/users/external
func NewExternalUser(ctx context.Context, req dto.NewExternalUserRequest) (ans *dto.NewUserResponse, err error) {
	tx, err := userDb.Begin(ctx)
	if err != nil {
		return
	}

	uid, err := createExternalUser(ctx, tx, req)
	if err != nil {
		rlog.Debug("here1")
		tx.Rollback()
		return
	}
	tx.Commit()

	ans = &dto.NewUserResponse{
		UserId: uid,
	}

	user, err := findUserByIdFromDb(ctx, uid)
	if err != nil {
		rlog.Error(util.MsgDbAccessError, "err", err)
		return
	}

	NewUsers.Publish(ctx, UserAccountCreated{
		UserId:    uid,
		AccountId: user.ProvidedAccounts[0].Id,
		Timestamp: time.Now(),
		NewUser:   true,
	})

	return
}

// Deletes a user's account (internal API)
//
//encore:api private method=DELETE path=/users/:id
func DeleteInternal(ctx context.Context, id uint64) (err error) {
	tx, err := userDb.Begin(ctx)
	if err != nil {
		rlog.Error(util.MsgDbAccessError, "msg", err.Error())
		err = &util.ErrUnknown
		return
	}
	tx.Commit()

	if err = deleteUserAccount(ctx, tx, id); err != nil {
		return
	}

	DeletedUsers.Publish(ctx, UserDeleted{
		UserId:    id,
		Timestamp: time.Now(),
	})
	return
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
func FetchUsers(ctx context.Context, req dto.CursorBasedPaginationParams) (*dto.FetchUsersResponse, error) {
	ans, err := findAllUsers(ctx, req.After, req.Size)
	if err != nil {
		return nil, err
	}

	return &dto.FetchUsersResponse{
		Users: usersToDto(ans...),
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

	hash, err := findUserPasswordHashById(ctx, user.Id)
	if errors.Is(err, sqldb.ErrNoRows) || hash == nil {
		return nil, &errs.Error{
			Code:    errs.FailedPrecondition,
			Message: "You have not configured a password using your provider",
		}
	} else if err != nil {
		rlog.Error(util.MsgDbAccessError, "msg", err.Error())
		return nil, &util.ErrUnknown
	}

	if err = bcrypt.CompareHashAndPassword([]byte(*hash), []byte(req.Password)); err != nil {
		return nil, &errs.Error{
			Code:    errs.Unauthenticated,
			Message: "Invalid email or password",
		}
	}

	return user, nil
}

// Create a new user account
//
//encore:api private method=POST path=/users/internal
func NewInternalUser(ctx context.Context, req dto.NewInternalUserRequest) (ans *dto.NewUserResponse, err error) {
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
		return
	}

	uid, err := createInternalUser(ctx, req, tx)
	if err != nil {
		_ = tx.Rollback()
		rlog.Error(err.Error())
		return
	}

	tx.Commit()

	ans = &dto.NewUserResponse{
		UserId: uid,
	}

	user, _ := findUserByIdFromDb(ctx, uid)

	NewUsers.Publish(ctx, UserAccountCreated{
		UserId:    uid,
		AccountId: user.ProvidedAccounts[0].Id,
		Timestamp: time.Now(),
		NewUser:   true,
	})
	return
}

// Find a user by their ID
//
//encore:api public method=GET path=/users/:id
func FindUserByIdPublic(ctx context.Context, id uint64) (ans *dto.User, err error) {
	user, err := FindUserById(ctx, id)
	if err != nil {
		if errs.Convert(err) == nil {
			return
		}

		rlog.Error(util.MsgCallError, "msg", err.Error())
		err = &util.ErrUnknown
		return
	} else if user == nil {
		err = &util.ErrNotFound
		return
	}

	ans = &usersToDto(user)[0]

	return
}

// Find a user by their internal account's ID (Private API)
//
//encore:api private method=GET path=/users/:id/internal-by-internal-id
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

type FindUserByExternalIdResponse struct {
	AccountIndex int
	User         models.User
}

// Find a user by their external account's ID (internal API)
//
//encore:api private method=GET path=/users/:id/internal-by-external-id
func FindUserByExternalId(ctx context.Context, id string) (ans *FindUserByExternalIdResponse, err error) {

	user, err := findUserByExternalIdFromCache(ctx, id)
	if errors.Is(err, cache.Miss) {
		rlog.Warn("cache miss", "userId", id)
		var t *models.User
		t, err = findUserByExternalIdFromdb(ctx, id)
		if t != nil {
			user = *t
		}
	}

	if err != nil {
		return
	}

	accountIndex, _ := helpers.FindIndex(user.ProvidedAccounts, func(a models.UserAccount) bool {
		return a.ExternalId == id
	})

	ans = &FindUserByExternalIdResponse{
		AccountIndex: accountIndex,
		User:         user,
	}
	return
}

func findUserByIdFromCache(ctx context.Context, id uint64) (*models.User, error) {

	u, err := idCache.Get(ctx, cacheKey(id))
	if err != nil {
		return nil, err
	}
	return &u, err
}

func findUserByExternalIdFromCache(ctx context.Context, id string) (u models.User, err error) {
	u, err = idCache.Get(ctx, cacheKey(id))
	if err != nil {
		return
	}

	return
}

func cacheKey[T string | uint64](id T) string {
	var identifier string
	switch v := any(id).(type) {
	case uint64:
		identifier = fmt.Sprintf("%d", v)
	case string:
		identifier = v
	}
	return identifier
}

func findUserByExternalIdFromdb(ctx context.Context, id string) (ans *models.User, err error) {
	query := `SELECT * FROM vw_AllUsers WHERE id=(SELECT "user" FROM provider_accounts WHERE external_id=$1);`
	ans, err = parseUserRow(userDb.QueryRow(ctx, query, id))
	if ans != nil {
		idCache.Set(ctx, cacheKey(id), *ans)
	}
	return
}

func findUserByIdFromDb(ctx context.Context, id uint64) (ans *models.User, err error) {
	query := "SELECT * FROM vw_AllUsers WHERE id=$1"
	ans, err = parseUserRow(userDb.QueryRow(ctx, query, id))
	if ans != nil {
		idCache.Set(ctx, cacheKey(id), *ans)
	}
	return
}

type rowScanner interface {
	Scan(dest ...any) error
}

func parseUserRow(scanner rowScanner) (ans *models.User, err error) {
	ans = new(models.User)
	var accountsJson, emailsJson, phonesJson string
	if err = scanner.Scan(&ans.Id, &ans.Banned, &ans.CreatedAt, &ans.UpdatedAt, &ans.Locked, &accountsJson, &ans.PrimaryEmail, &ans.PrimaryPhone, &emailsJson, &phonesJson); err != nil {
		ans = nil
		return
	}

	if err = json.Unmarshal([]byte(accountsJson), &ans.ProvidedAccounts); err != nil {
		ans = nil
		return
	}

	if err = json.Unmarshal([]byte(emailsJson), &ans.Emails); err != nil {
		ans = nil
		return
	}

	if err = json.Unmarshal([]byte(phonesJson), &ans.PhoneNumbers); err != nil {
		ans = nil
		return
	}
	return
}

func userEmailExists(ctx context.Context, email string) (ans bool, err error) {
	query := "SELECT COUNT(id) FROM account_emails WHERE email=$1;"

	var cnt = 0
	if err = userDb.QueryRow(ctx, query, email).Scan(&cnt); err != nil {
		return
	}
	ans = cnt > 0

	return
}

func findUserByEmailFromDb(ctx context.Context, email string) (ans *models.User, err error) {
	query := `
		SELECT 
			* 
		FROM 
			vw_AllUsers 
		WHERE 
			pa.id=(
				SELECT 
					"user" 
				FROM 
					provider_accounts pa 
				WHERE id=(
					SELECT
						ae.account
					FROM
						account_emails ae
					WHERE
						ae.email=$1
				)
			)
		;
	`
	ans, err = parseUserRow(userDb.QueryRow(ctx, query, email))
	if err != nil {
		return
	}

	emailCache.Set(ctx, email, *ans)
	return
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
		user, err := parseUserRow(rows)
		if err != nil {
			return nil, err
		}
		ans = append(ans, user)
		idCache.Set(ctx, cacheKey(user.Id), *user)
	}

	return ans, nil
}

func createInternalUser(ctx context.Context, req dto.NewInternalUserRequest, tx *sqldb.Tx) (ans uint64, err error) {

	dob, _ := time.Parse("2006/2/1", req.Dob)
	ph, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	avatar := fmt.Sprintf("https://api.dicebear.com/9.x/adventurer/svg?seed?=%s&scale=80", url.QueryEscape(strings.Trim(fmt.Sprintf("%s %s", req.FirstName, req.LastName), " \n\t")))

	query := `SELECT func_create_internal_user($1,$2,$3,$4,$5,$6,$7,$8,$9,$10);`

	if err = tx.QueryRow(ctx, query, req.Email, string(ph), avatar, req.FirstName, req.LastName, req.Gender, dob, req.Phone, true, true).Scan(&ans); err != nil {
		rlog.Error(err.Error())
		err = &util.ErrUnknown
	}

	return
}

func deleteUserAccount(ctx context.Context, tx *sqldb.Tx, user uint64) (err error) {
	if _, err = tx.Exec(ctx, "CALL proc_delete_user($1);", user); err != nil {
		return
	}
	return
}

func usersToDto(u ...*models.User) (ans []dto.User) {
	ans = make([]dto.User, len(u))

	for i, v := range u {
		var u = dto.User{
			Id:               v.Id,
			Banned:           v.Banned,
			Locked:           v.Locked,
			CreatedAt:        v.CreatedAt,
			UpdatedAt:        v.UpdatedAt,
			ProvidedAccounts: make([]dto.UserAccount, len(v.ProvidedAccounts)),
			EmailsAddresses:  make([]dto.UserEmailAddress, len(v.Emails)),
			PhoneNumbers:     make([]dto.UserPhoneNumber, len(v.PhoneNumbers)),
		}

		if v.PrimaryPhone.Valid {
			tmp := uint64(v.PrimaryPhone.Int64)
			u.PrimaryPhone = &tmp
		}

		if v.PrimaryEmail.Valid {
			tmp := uint64(v.PrimaryEmail.Int64)
			u.PrimaryEmail = &tmp
		}

		for j, vv := range v.ProvidedAccounts {
			u.ProvidedAccounts[j] = dto.UserAccount{
				Id:                  vv.Id,
				ExternalId:          vv.ExternalId,
				ImageUrl:            vv.ImageUrl,
				User:                vv.User,
				FirstName:           vv.FirstName,
				LastName:            vv.LastName,
				Provider:            vv.Provider,
				ProviderProfileData: vv.ProviderProfileData,
				Gender:              vv.Gender,
			}
			if vv.Dob.Valid {
				u.ProvidedAccounts[j].Dob = &vv.Dob.Time
			}
		}

		for j, vv := range v.Emails {
			u.EmailsAddresses[j] = dto.UserEmailAddress{
				Id:         vv.Id,
				Email:      vv.Email,
				Account:    vv.Account,
				ExternalId: vv.ExternalId,
				IsPrimary:  vv.IsPrimary,
				Verified:   vv.Verified,
			}
		}

		for j, vv := range v.PhoneNumbers {
			u.PhoneNumbers[j] = dto.UserPhoneNumber{
				Id:         vv.Id,
				Phone:      vv.Phone,
				Account:    vv.Account,
				ExternalId: vv.ExternalId,
				IsPrimary:  vv.IsPrimary,
				Verified:   vv.Verified,
			}
		}

		ans[i] = u
	}
	return
}

func findUserPasswordHashById(ctx context.Context, id uint64) (ans *string, err error) {
	query := `SELECT password_hash FROM provider_accounts WHERE "user" = $1 AND provider='internal';`
	var passwordHash sql.NullString
	if err = userDb.QueryRow(ctx, query, id).Scan(&passwordHash); err != nil {
		return
	}
	ans = &passwordHash.String
	return
}

func createExternalUser(ctx context.Context, tx *sqldb.Tx, req dto.NewExternalUserRequest) (ans uint64, err error) {
	query := "SELECT func_create_external_user($1,$2,$3,$4,$5,$6,$7,$8,$9,$10);"

	emailsJson := helpers.Map(req.Emails, func(e dto.ExternalUserEmailAddressData) string {
		j, _ := json.Marshal(e)
		return string(j)
	})

	phonesJson := helpers.Map(req.Phones, func(a dto.ExternalUserPhoneData) string {
		j, _ := json.Marshal(a)
		return string(j)
	})

	err = tx.QueryRow(ctx, query, req.FirstName, req.LastName, req.ExternalId, req.ProviderData, pq.Array(emailsJson), pq.Array(phonesJson), req.Provider, req.Gender, req.Dob, req.Avatar).Scan(&ans)

	return
}
