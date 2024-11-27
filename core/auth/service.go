// All endpoints for authentication
package auth

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"encore.dev"
	"encore.dev/beta/auth"
	"encore.dev/rlog"
	"github.com/brinestone/scholaris/core/users"
	"github.com/brinestone/scholaris/dto"
	"github.com/brinestone/scholaris/helpers"
	"github.com/brinestone/scholaris/models"
	"github.com/brinestone/scholaris/util"
	"github.com/clerkinc/clerk-sdk-go/clerk"
	"github.com/golang-jwt/jwt/v4"
)

var jwtSigningMethod = jwt.SigningMethodHS256

// Service definition
//
//encore:service
type Service struct {
	client clerk.Client
}

type VerifyCaptchaRequest struct {
	// The request's reCaptcha token from the client.
	Token string `json:"token"`
}

// Deletes a user's account.
//
//encore:api auth method=POST path=/auth/delete tag:needs_captcha_ver
func (s *Service) DeleteAccount(ctx context.Context, req dto.DeleteAccountRequest) (err error) {
	uid, _ := auth.UserID()
	userId, _ := strconv.ParseUint(string(uid), 10, 64)

	if err = deleteUserAccount(ctx, userId); err != nil {
		rlog.Error("user account deletion error", "msg", err.Error())
		err = &util.ErrUnknown
	}

	return
}

// Verifies reCaptcha tokens
//
//encore:api private method=POST path=/auth/recaptcha-verify
func (s *Service) VerifyCaptchaToken(ctx context.Context, req VerifyCaptchaRequest) error {
	return verifyCaptcha(req.Token)
}

type LoginResponse struct {
	// The user's access token
	AccessToken string `json:"accessToken"`
}

// Signs in an existing user using their email and password
//
//encore:api public method=POST path=/auth/sign-in tag:sign-in
func (s *Service) SignIn(ctx context.Context, req dto.LoginRequest) (*LoginResponse, error) {
	if err := verifyCaptcha(req.CaptchaToken); err != nil {
		rlog.Error(err.Error())
		return nil, &util.ErrCaptchaError
	}

	var ans = new(LoginResponse)
	user, err := users.VerifyCredentials(ctx, req)
	if err != nil {
		return nil, err
	}

	var accountIndex, _ = helpers.Find(user.ProvidedAccounts, func(a models.UserAccount) bool {
		emailIndex, ok := helpers.Find(user.Emails, func(e models.UserEmailAddress) bool {
			return e.Email == req.Email
		})
		return ok && a.Id == user.Emails[emailIndex].Id
	})
	var account = user.ProvidedAccounts[accountIndex]

	claims := jwt.MapClaims{
		"sub":      user.Id,
		"iss":      "scholaris",
		"avatar":   account.ImageUrl,
		"email":    req.Email,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
		"fullName": account.FullName(),
		"provider": account.Provider,
		"mode":     encore.Meta().Environment,
	}

	token := jwt.NewWithClaims(jwtSigningMethod, claims)
	serializedToken, err := token.SignedString([]byte(secrets.JwtKey))
	if err != nil {
		return nil, &util.ErrUnauthorized
	}

	ans.AccessToken = serializedToken

	SignIns.Publish(ctx, UserSignedIn{
		Email:     req.Email,
		UserId:    user.Id,
		Timestamp: time.Now(),
	})

	return ans, nil
}

type CaptchaCheckResponse struct {
	Success            bool      `json:"success"`
	ChallengeTimestamp time.Time `json:"challenge_ts"`
	HostName           string    `json:"hostname"`
	ErrorCodes         []string  `json:"error-codes,omitempty" encore:"optional"`
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
//encore:api public method=POST path=/auth/sign-up
func (s *Service) SignUp(ctx context.Context, req dto.NewInternalUserRequest) error {
	if err := verifyCaptcha(req.CaptchaToken); err != nil {
		rlog.Error(err.Error())
		return &util.ErrCaptchaError
	}

	user, err := users.NewInternalUser(ctx, req)
	if err != nil {
		return err
	}

	SignUps.Publish(ctx, UserSignedUp{
		Email:  req.Email,
		UserId: user.UserId,
	})
	return nil
}

// ----

type AuthClaims struct {
	Email      string `json:"email"`
	Avatar     string `json:"avatar"`
	Provider   string `json:"provider"`
	ExternalId string `json:"externalId"`
	FullName   string `json:"displayName"`
	Mode       string `json:"mode"`
	Sub        uint64
}

var secrets struct {
	JwtKey           string
	CaptchaSecretKey string
	ClerkSecret      string
}

func initService() (ans *Service, err error) {
	client, err := clerk.NewClient(secrets.ClerkSecret)
	if err != nil {
		return
	}
	ans = &Service{client: client}
	return
}

//encore:authhandler
func (s *Service) AuthHandler(ctx context.Context, token string) (ans auth.UID, claims *AuthClaims, err error) {
	claims = &AuthClaims{}
	sessionClaims, err := s.client.VerifyToken(token, clerk.WithCustomClaims(claims))
	if err != nil {
		rlog.Error("clerk error", "msg", err)
		err = &util.ErrUnauthorized
		return
	}

	res, err := users.FindUserByExternalId(ctx, sessionClaims.Subject)
	if err != nil {
		rlog.Error(util.MsgCallError, "msg", err)
		err = &util.ErrUnknown
		return
	}

	claims.Sub = res.User.Id
	rlog.Debug("handled jwt", "token", token, "claims", claims, "sessionClaims", sessionClaims)

	return
}

func deleteUserAccount(ctx context.Context, user uint64) (err error) {
	err = users.DeleteInternal(ctx, user)
	return
}
