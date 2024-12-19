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
	"strconv"
	"strings"
	"time"

	"encore.dev"
	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/rlog"
	"github.com/brinestone/scholaris/core/users"
	"github.com/brinestone/scholaris/dto"
	"github.com/brinestone/scholaris/helpers"
	"github.com/brinestone/scholaris/models"
	"github.com/brinestone/scholaris/util"
	"github.com/clerkinc/clerk-sdk-go/clerk"
	"github.com/golang-jwt/jwt/v4"
)

var (
	jwtSigningMethod = jwt.SigningMethodHS256
	ValidProviders   = helpers.SliceOf(dto.ProvClerk, dto.ProvInternal)
)

// Service definition
//
//encore:service
type Service struct {
	clerkClient clerk.Client
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

	var accountIndex, _ = helpers.FindIndex(user.ProvidedAccounts, func(a models.UserAccount) bool {
		emailIndex, ok := helpers.FindIndex(user.Emails, func(e models.UserEmailAddress) bool {
			return e.Email == req.Email
		})
		return ok && a.Id == user.Emails[emailIndex].Id
	})
	var account = user.ProvidedAccounts[accountIndex]

	claims := jwt.MapClaims{
		"sub":         user.Id,
		"iss":         encore.Meta().APIBaseURL,
		"avatar":      account.ImageUrl,
		"email":       req.Email,
		"exp":         time.Now().Add(time.Hour * 24).Unix(),
		"displayName": account.FullName(),
		"provider":    account.Provider,
		"mode":        encore.Meta().Environment,
	}

	phone, phoneAvailable := helpers.Find(user.PhoneNumbers, func(p models.UserPhoneNumber) bool {
		return p.IsPrimary
	})

	if phoneAvailable {
		claims["phone"] = phone.Phone
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
	ans = &Service{clerkClient: client}
	return
}

func getValidIssuerDomain(issuerDomain string) (result string, valid bool) {
	u, err := url.Parse(issuerDomain)
	if err != nil {
		return
	}

	result, valid = helpers.Find(dto.ValidIssuerDomains, func(validDomain string) bool {
		return strings.HasSuffix(u.Hostname(), validDomain)
	})
	return
}

//encore:authhandler
func (s *Service) AuthHandler(ctx context.Context, token string) (ans auth.UID, claims *dto.AuthClaims, err error) {
	t, _, err := new(jwt.Parser).ParseUnverified(token, jwt.MapClaims{})
	if err != nil {
		rlog.Error("jwt parse error", "err", err)
		err = &util.ErrUnknown
		return
	}

	sentClaims, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		rlog.Warn("invalid claims structure", "actualClaims", t.Claims)
		err = &util.ErrUnauthorized
		return
	}

	issuerClaim, ok := sentClaims["iss"]
	if !ok {
		rlog.Warn("no issuer specified", "actualClaims", t.Claims)
		err = &util.ErrUnauthorized
		return
	}

	issuerDomain, valid := getValidIssuerDomain(issuerClaim.(string))
	if !valid {
		rlog.Warn("invalid issuer", "actualClaims", t.Claims, "issuerClaim", issuerClaim)
		err = &util.ErrUnauthorized
		return
	}

	var user *models.User
	switch issuerDomain {
	case dto.ScholarisIssuerDomain:
		claims, user, err = doInternalJwtValidation(ctx, token)
	case dto.ClerkIssuerDomain:
		claims, user, err = s.doClerkJwtValidation(ctx, token)
	default:
		rlog.Warn("unsupported issuer", "actualClaims", t.Claims, "issuerDomain", issuerDomain)
		err = &util.ErrUnauthorized
		return
	}

	if err != nil && errs.Convert(err) == nil {
		return
	} else if err != nil {
		rlog.Error("jwt parse error", "err", err.Error())
		err = &util.ErrUnauthorized
		return
	}

	ans = auth.UID(fmt.Sprint(user.Id))

	return
}

// Performs JWT validation and parsing using Clerk client
func (s *Service) doClerkJwtValidation(ctx context.Context, token string) (ans *dto.AuthClaims, user *models.User, err error) {
	sessionClaims, err := s.clerkClient.VerifyToken(token)
	if err != nil {
		return
	}

	res, err := users.FindUserByExternalId(ctx, sessionClaims.Subject)
	if err != nil {
		return
	}

	account := res.User.ProvidedAccounts[res.AccountIndex]
	email, _ := helpers.Find(res.User.Emails, func(a models.UserEmailAddress) bool { return a.IsPrimary })
	phone, phoneAvailable := helpers.Find(res.User.PhoneNumbers, func(p models.UserPhoneNumber) bool { return p.IsPrimary })
	user = &res.User

	ans = &dto.AuthClaims{
		Email:      email.Email,
		Avatar:     account.ImageUrl,
		Provider:   "clerk",
		ExternalId: sessionClaims.Subject,
		FullName:   account.FullName(),
		Sub:        res.User.Id,
		Account:    account.Id,
		Phone:      nil,
	}

	if phoneAvailable {
		ans.Phone = &phone.Phone
	}

	return
}

// Performs JWT validation and parsing
func doInternalJwtValidation(ctx context.Context, token string) (ans *dto.AuthClaims, user *models.User, err error) {
	var claims jwt.MapClaims = make(jwt.MapClaims)

	t, err := jwt.ParseWithClaims(token, claims, findJwtKey, jwt.WithValidMethods(helpers.SliceOf(jwtSigningMethod.Alg())))
	if errors.Is(err, jwt.ErrTokenExpired) {
		err = &util.ErrUnauthorized
		return
	} else if err != nil {
		return
	}

	if !t.Valid {
		err = &util.ErrUnauthorized
		return
	}

	ans = new(dto.AuthClaims)
	ans.Provider = "internal"
	var userId uint64
	if temp, ok := claims["phone"].(string); ok {
		ans.Phone = &temp
	}
	if temp, ok := claims["avatar"].(string); ok {
		ans.Avatar = &temp
	}
	if temp, ok := claims["displayName"].(string); ok {
		ans.FullName = temp
	}

	if temp, ok := claims["email"].(string); ok {
		ans.Email = temp
	}

	if temp, ok := claims["sub"].(float64); ok {
		userId = uint64(temp)
		ans.Sub = userId
	}

	if temp, ok := claims["account"].(float64); ok {
		ans.Account = uint64(temp)
	}

	user, err = users.FindUserById(ctx, userId)
	if err != nil || user == nil || user.Id != userId {
		err = &util.ErrUnauthorized
		return
	}

	return
}

func deleteUserAccount(ctx context.Context, user uint64) (err error) {
	err = users.DeleteInternal(ctx, user)
	return
}

func findJwtKey(t *jwt.Token) (ans any, err error) {
	ans = []byte(secrets.JwtKey)
	return
}
