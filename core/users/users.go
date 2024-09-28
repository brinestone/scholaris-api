package users

import (
	"context"
	"errors"
	"time"

	"encore.dev/beta/errs"
	"encore.dev/rlog"
	"encore.dev/storage/sqldb"
	"github.com/brinestone/scholaris/dto"
	"github.com/brinestone/scholaris/models"
	"golang.org/x/crypto/bcrypt"
)

// Finds a user using their email address
//
//encore:api private method=GET path=/users
func FindUserByEmail(ctx context.Context, req dto.UserLookupByEmailRequest) (*models.User, error) {
	return findUserByEmail(ctx, req.Email)
}

// Create a new user account
//
//encore:api private method=POST path=/users
func NewUser(ctx context.Context, req dto.NewUserRequest) (*models.User, error) {

	existingUser, err := findUserByEmail(ctx, req.Email)
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
		return nil, errors.New("unknown error occured")
	}
	return nil, nil
}

// Find a user by their ID
//
//encore:api private method=GET path=/users/:id
func FindUserById(ctx context.Context, id string) (*models.User, error) {
	return nil, nil
}

func findUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := "SELECT id,first_name,last_name,email,dob,password_hash,phone,created_at,updated_at FROM users WHERE email = $1;"
	var ans *models.User = new(models.User)

	row := userDb.QueryRow(ctx, query, email)

	if err := row.Scan(&ans.Id, &ans.FirstName, &ans.LastName, &ans.Email, &ans.Dob, &ans.PasswordHash, &ans.Phone, &ans.CreatedAt, &ans.UpdatedAt); err != nil {
		if errors.Is(err, sqldb.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return ans, nil
}
