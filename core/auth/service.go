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
//encore:api public method=POST path=/auth/sign-up tag:new
func (s *Service) SignUp(ctx context.Context, req dto.NewUserRequest) error {
	if err := verifyCaptcha(req.CaptchaToken); err != nil {
		rlog.Error(err.Error())
		return &util.ErrCaptchaError
	}

	user, err := users.NewUser(ctx, req)
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
	Sub        uint64 `json:"sub"`
	Mode       string `json:"mode"`
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
	sessionClaims, err := s.client.VerifyToken(token, clerk.WithCustomClaims(AuthClaims{}))
	if err != nil {
		err = &util.ErrUnauthorized
		return
	}

	user, err := s.client.Users().Read(sessionClaims.Subject)
	if err != nil {
		rlog.Error("clerk error", "msg", err.Error())
		err = &util.ErrUnknown
		return
	}

	var email string
	for _, e := range user.EmailAddresses {
		if e.ID != *user.PrimaryEmailAddressID {
			continue
		}
		email = e.EmailAddress
		break
	}

	// TODO: rethink this.
	_, err = users.FindUserByExternalId(ctx, sessionClaims.Subject)
	if err != nil {
		rlog.Error(util.MsgCallError, "msg", err.Error())
		err = &util.ErrUnknown
		return
	}

	// TODO: find user from db
	claims = &AuthClaims{
		Email:  email,
		Avatar: user.ProfileImageURL,
	}
	return
}

//// func (s *Service) JwtAuthHandler(ctx context.Context, token string) (auth.UID, *AuthClaims, error) {
//// 	var claims jwt.MapClaims = make(jwt.MapClaims)
//
//// 	t, err := jwt.ParseWithClaims(token, claims, findJwtToken, jwt.WithValidMethods([]string{jwt.SigningMethodES256.Alg()}))
//// 	if err != nil {
//// 		if errors.Is(err, jwt.ErrTokenExpired) {
//// 			return "", nil, &util.ErrUnauthorized
//// 		}
//// 		return "", nil, err
//// 	}
//
//// 	if !t.Valid {
//// 		return "", nil, &util.ErrUnauthorized
//// 	}
//
//// 	var id uint64
//
//// 	authClaims := new(AuthClaims)
//// 	if temp, ok := claims["avatar"].(string); ok {
//// 		authClaims.Avatar = temp
//// 	}
//// 	if temp, ok := claims["fullName"].(string); ok {
//// 		authClaims.FullName = temp
//// 	}
//
//// 	if temp, ok := claims["email"].(string); ok {
//// 		authClaims.Email = temp
//// 	}
//
//// 	if temp, ok := claims["sub"].(float64); ok {
//// 		id = uint64(temp)
//// 		authClaims.Sub = id
//// 	}
//
//// 	user, err := users.FindUserByIdInternal(ctx, id)
//// 	if err != nil || user == nil || user.Id != id {
//// 		return "", nil, &util.ErrUnauthorized
//// 	}
//
//// 	return auth.UID(fmt.Sprint(id)), authClaims, nil
//// }

// func findJwtToken(t *jwt.Token) (any, error) {
// 	return []byte(secrets.JwtKey), nil
// }

func deleteUserAccount(ctx context.Context, user uint64) (err error) {
	err = users.DeleteInternal(ctx, user)
	return
}
