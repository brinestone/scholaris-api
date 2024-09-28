package auth

import (
	"context"
	"fmt"
	"strings"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/rlog"
	"github.com/brinestone/scholaris/core/users"
	"github.com/brinestone/scholaris/dto"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

type LoginResponse struct {
	AccessToken string `json:"access_token"`
}

// Login user
//
//encore:api public method=POST tag:login
func LoginUser(ctx context.Context, req dto.LoginRequest) (*LoginResponse, error) {
	var ans = new(LoginResponse)
	user, err := users.FindUserByEmail(ctx, dto.UserLookupByEmailRequest{
		Email: req.Email,
	})
	if err != nil {
		rlog.Error(err.Error())
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

	claims := jwt.MapClaims{
		"sub":         user.Id,
		"iss":         "scholaris",
		"avatar":      user.Avatar,
		"displayName": strings.Trim(fmt.Sprintf("%s %s", user.FirstName, user.LastName), "\t\n"),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	serializedToken, err := token.SignedString([]byte(secrets.JwtKey))
	if err != nil {
		return nil, err
	}

	ans.AccessToken = serializedToken

	return ans, nil
}

// Signs a user
//
//encore:api public method=POST tag:new
func SignUp(ctx context.Context, req dto.NewUserRequest) error {

	_, err := users.NewUser(ctx, req)
	if err != nil {
		return err
	}

	SignUps.Publish(ctx, &UserSignedUp{
		Email:  req.Email,
		UserId: 0,
	})
	return nil
}

// ----

type AuthClaims struct {
	Email       string `json:"email"`
	Avatar      string `json:"avatar,omitempty"`
	DisplayName string `json:"displayName"`
	Sub         string `json:"sub"`
}

var secrets struct {
	JwtKey string
}

//encore:authhandler
func AuthHandler(ctx context.Context, token string) (auth.UID, *AuthClaims, error) {
	var claims jwt.MapClaims
	t, err := jwt.ParseWithClaims(token, claims, findJwtToken, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return "", nil, err
	}

	if !t.Valid {
		return "", nil, &errs.Error{
			Code: errs.Unauthenticated,
		}
	}

	id := claims["sub"].(int64)

	user, err := users.FindUserById(ctx, string(id))
	if err != nil {
		return "", nil, err
	}

	authClaims := new(AuthClaims)
	authClaims.Avatar = user.Avatar
	authClaims.DisplayName = fmt.Sprintf("%s %s", user.FirstName, user.LastName)

	return auth.UID(id), authClaims, nil
}

func findJwtToken(t *jwt.Token) (any, error) {
	return []byte(secrets.JwtKey), nil
}
