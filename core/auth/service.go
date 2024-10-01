package auth

import (
	"context"
	"fmt"
	"time"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"github.com/brinestone/scholaris/core/users"
	"github.com/brinestone/scholaris/dto"
	"github.com/golang-jwt/jwt/v4"
)

var jwtSigningMethod = jwt.SigningMethodHS256

type LoginResponse struct {
	AccessToken string `json:"accessToken"`
}

// Signs in an existing user using their email and password
//
//encore:api public method=POST tag:login
func LoginUser(ctx context.Context, req dto.LoginRequest) (*LoginResponse, error) {
	var ans = new(LoginResponse)
	user, err := users.VerifyCredentials(ctx, req)
	if err != nil {
		return nil, err
	}

	claims := jwt.MapClaims{
		"sub":      user.Id,
		"iss":      "scholaris",
		"avatar":   user.Avatar,
		"email":    user.Email,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
		"fullName": user.FullName(),
	}

	token := jwt.NewWithClaims(jwtSigningMethod, claims)
	serializedToken, err := token.SignedString([]byte(secrets.JwtKey))
	if err != nil {
		return nil, &errs.Error{
			Code: errs.Unauthenticated,
		}
	}

	ans.AccessToken = serializedToken

	return ans, nil
}

// Creates a new user account
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
	Email    string `json:"email"`
	Avatar   string `json:"avatar,omitempty"`
	FullName string `json:"displayName"`
	Sub      uint64 `json:"sub"`
}

var secrets struct {
	JwtKey string
}

//encore:authhandler
func JwtAuthHandler(ctx context.Context, token string) (auth.UID, *AuthClaims, error) {
	var claims jwt.MapClaims = make(jwt.MapClaims)

	t, err := jwt.ParseWithClaims(token, claims, findJwtToken, jwt.WithValidMethods([]string{jwtSigningMethod.Alg()}))
	if err != nil {
		return "", nil, err
	}

	if !t.Valid {
		return "", nil, &errs.Error{
			Code: errs.Unauthenticated,
		}
	}

	var id uint64

	authClaims := new(AuthClaims)
	if temp, ok := claims["avatar"].(string); ok {
		authClaims.Avatar = temp
	}
	if temp, ok := claims["fullName"].(string); ok {
		authClaims.FullName = temp
	}

	if temp, ok := claims["email"].(string); ok {
		authClaims.Email = temp
	}

	if temp, ok := claims["sub"].(float64); ok {
		id = uint64(temp)
		authClaims.Sub = id
	}

	user, err := users.FindUserById(ctx, id)
	if err != nil || user == nil || user.Id != id {
		return "", nil, &errs.Error{
			Code: errs.Unauthenticated,
		}
	}

	return auth.UID(fmt.Sprint(id)), authClaims, nil
}

func findJwtToken(t *jwt.Token) (any, error) {
	return []byte(secrets.JwtKey), nil
}
