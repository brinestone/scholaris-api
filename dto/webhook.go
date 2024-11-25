package dto

// Clerk Event keys
const (
	CEUserCreated = "user.created"
	CEUserDeleted = "user.deleted"
	// When a user has used clerk to sign in
	CESessionCreated = "session.created"
)

type ClerkEvent struct {
	Data            ClerkEventData       `json:"data"`
	EventAttributes ClerkEventAttributes `json:"event_attributes"`
	Object          string               `json:"object"`
	Timestamp       int64                `json:"timestamp"`
	Type            string               `json:"type"`
}

type ClerkEventData struct {
	Birthday              string                 `json:"birthday"`
	CreatedAt             int64                  `json:"created_at"`
	EmailAddresses        []ClerkEmailAddress    `json:"email_addresses"`
	ExternalAccounts      []ClerkExternalAccount `json:"external_accounts"`
	ExternalID            string                 `json:"external_id"`
	FirstName             string                 `json:"first_name"`
	Gender                string                 `json:"gender"`
	ID                    string                 `json:"id"`
	ImageURL              string                 `json:"image_url"`
	LastName              string                 `json:"last_name"`
	LastSignInAt          int64                  `json:"last_sign_in_at"`
	Object                string                 `json:"object"`
	PasswordEnabled       bool                   `json:"password_enabled"`
	PhoneNumbers          []ClerkPhoneNumber     `json:"phone_numbers"`
	PrimaryEmailAddressID string                 `json:"primary_email_address_id"`
	PrimaryPhoneNumberID  *string                `json:"primary_phone_number_id,omitempty"`
	PrimaryWeb3WalletID   *string                `json:"primary_web3_wallet_id,omitempty"`
	PrivateMetadata       *ClerkMetadata         `json:"private_metadata,omitempty"`
	ProfileImageURL       string                 `json:"profile_image_url"`
	PublicMetadata        ClerkMetadata          `json:"public_metadata"`
	TwoFactorEnabled      bool                   `json:"two_factor_enabled"`
	UnsafeMetadata        ClerkMetadata          `json:"unsafe_metadata"`
	UpdatedAt             int64                  `json:"updated_at"`
	Username              string                 `json:"username,omitempty"`
	Web3Wallets           []ClerkWeb3Wallet      `json:"web3_wallets,omitempty"`
}

type ClerkExternalAccount struct{}

type ClerkPhoneNumber struct {
	DefaultSecondFactor     bool                   `json:"default_second_factor"`
	Id                      string                 `json:"id"`
	LinkedTo                []string               `json:"linked_to,omitempty"`
	Object                  string                 `json:"object,omitempty"`
	PhoneNumber             string                 `json:"phone_number,omitempty"`
	ReservedForSecondFactor bool                   `json:"reserved_for_second_factor,omitempty"`
	Verification            ClerkPhoneVerification `json:"verification,omitempty"`
}

type ClerkPhoneVerification struct {
	Attepmts int    `json:"attempts"`
	Status   string `json:"status"`
	Type     string `json:"type"`
}

type ClerkWeb3Wallet struct {
	Id           string            `json:"id,omitempty"`
	Object       string            `json:"object,omitempty"`
	Verification ClerkVerification `json:"verification,omitempty"`
	Web3Wallet   string            `json:"web3_wallet,omitempty"`
}

type ClerkEmailAddress struct {
	EmailAddress string            `json:"email_address"`
	ID           string            `json:"id"`
	LinkedTo     []string          `json:"linked_to,omitempty"`
	Object       string            `json:"object"`
	Verification ClerkVerification `json:"verification"`
}

type ClerkVerification struct {
	Status   string `json:"status"`
	Strategy string `json:"strategy"`
	Nonce    string `json:"nonce,omitempty"`
	Attempts int64  `json:"attempts,omitempty"`
	ExpireAt int64  `json:"expire_at"`
}

type ClerkMetadata struct {
}

type ClerkEventAttributes struct {
	HTTPRequest ClerkHTTPRequest `json:"http_request"`
}

type ClerkHTTPRequest struct {
	ClientIP  string `json:"client_ip"`
	UserAgent string `json:"user_agent"`
}
