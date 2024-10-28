// All endpoints for authentication
package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"encore.dev/beta/auth"
	"encore.dev/rlog"
	"github.com/brinestone/scholaris/core/users"
	"github.com/brinestone/scholaris/dto"
	"github.com/brinestone/scholaris/util"
	"github.com/golang-jwt/jwt/v4"
)

var jwtSigningMethod = jwt.SigningMethodHS256

type VerifyCaptchaRequest struct {
	// The request's reCaptcha token from the client.
	Token string `json:"token"`
}

// Verifies reCaptcha tokens
//
//encore:api private method=POST path=/auth/recaptcha-verify
func VerifyCaptchaToken(ctx context.Context, req VerifyCaptchaRequest) error {
	return verifyCaptcha(req.Token)
}

type LoginResponse struct {
	// The user's access token
	AccessToken string `json:"accessToken"`
}

// Signs in an existing user using their email and password
//
//encore:api public method=POST path=/auth/sign-in tag:sign-in
func SignIn(ctx context.Context, req dto.LoginRequest) (*LoginResponse, error) {
	if err := verifyCaptcha(req.CaptchaToken); err != nil {
		rlog.Error(err.Error())
		return nil, &util.ErrCaptchaError
	}

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
		return nil, &util.ErrUnauthorized
	}

	ans.AccessToken = serializedToken

	return ans, nil
}

type CaptchaCheckResponse struct {
	Success            bool      `json:"success"`
	ChallengeTimestamp time.Time `json:"challenge_ts"`
	HostName           string    `json:"hostname"`
	ErrorCodes         []string  `json:"error-codes,omitempty"`
}

func verifyCaptcha(token string) error {
	req := make(url.Values)
	req.Add("secret", secrets.CaptchaSecretKey)
	req.Add("response", token)

	response, err := http.PostForm("https://www.google.com/recaptcha/api/siteverify", req)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	// var bodyReader = new(bytes.Reader)

	j, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	var captchaResponse CaptchaCheckResponse
	if err = json.Unmarshal(j, &captchaResponse); err != nil {
		return err
	}

	if !captchaResponse.Success {
		return errors.New(strings.Trim(strings.Join(captchaResponse.ErrorCodes, "\n"), "\n\t"))
	}

	return nil
}

// Creates a new user account
//
//encore:api public method=POST path=/auth/sign-up tag:new
func SignUp(ctx context.Context, req dto.NewUserRequest) error {
	if err := verifyCaptcha(req.CaptchaToken); err != nil {
		rlog.Error(err.Error())
		return &util.ErrCaptchaError
	}

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
	JwtKey           string
	CaptchaSecretKey string
}

//encore:authhandler
func JwtAuthHandler(ctx context.Context, token string) (auth.UID, *AuthClaims, error) {
	var claims jwt.MapClaims = make(jwt.MapClaims)

	t, err := jwt.ParseWithClaims(token, claims, findJwtToken, jwt.WithValidMethods([]string{jwtSigningMethod.Alg()}))
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return "", nil, &util.ErrUnauthorized
		}
		return "", nil, err
	}

	if !t.Valid {
		return "", nil, &util.ErrUnauthorized
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
		return "", nil, &util.ErrUnauthorized
	}

	return auth.UID(fmt.Sprint(id)), authClaims, nil
}

func findJwtToken(t *jwt.Token) (any, error) {
	return []byte(secrets.JwtKey), nil
}
