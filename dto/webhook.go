package dto

type ClerkEvent struct {
	Data            ClerkEventData       `json:"data"`
	EventAttributes ClerkEventAttributes `json:"event_attributes"`
	Object          string               `json:"object"`
	Timestamp       int64                `json:"timestamp"`
	Type            string               `json:"type"`
}

type ClerkEventData struct {
	Birthday              string              `json:"birthday"`
	CreatedAt             int64               `json:"created_at"`
	EmailAddresses        []ClerkEmailAddress `json:"email_addresses"`
	ExternalAccounts      []interface{}       `json:"external_accounts"`
	ExternalID            string              `json:"external_id"`
	FirstName             string              `json:"first_name"`
	Gender                string              `json:"gender"`
	ID                    string              `json:"id"`
	ImageURL              string              `json:"image_url"`
	LastName              string              `json:"last_name"`
	LastSignInAt          int64               `json:"last_sign_in_at"`
	Object                string              `json:"object"`
	PasswordEnabled       bool                `json:"password_enabled"`
	PhoneNumbers          []interface{}       `json:"phone_numbers"`
	PrimaryEmailAddressID string              `json:"primary_email_address_id"`
	PrimaryPhoneNumberID  interface{}         `json:"primary_phone_number_id"`
	PrimaryWeb3WalletID   interface{}         `json:"primary_web3_wallet_id"`
	PrivateMetadata       ClertMetadata       `json:"private_metadata"`
	ProfileImageURL       string              `json:"profile_image_url"`
	PublicMetadata        ClertMetadata       `json:"public_metadata"`
	TwoFactorEnabled      bool                `json:"two_factor_enabled"`
	UnsafeMetadata        ClertMetadata       `json:"unsafe_metadata"`
	UpdatedAt             int64               `json:"updated_at"`
	Username              interface{}         `json:"username"`
	Web3Wallets           []interface{}       `json:"web3_wallets"`
}

type ClerkEmailAddress struct {
	EmailAddress string            `json:"email_address"`
	ID           string            `json:"id"`
	LinkedTo     []interface{}     `json:"linked_to"`
	Object       string            `json:"object"`
	Verification ClerkVerification `json:"verification"`
}

type ClerkVerification struct {
	Status   string `json:"status"`
	Strategy string `json:"strategy"`
}

type ClertMetadata struct {
}

type ClerkEventAttributes struct {
	HTTPRequest ClerkHTTPRequest `json:"http_request"`
}

type ClerkHTTPRequest struct {
	ClientIP  string `json:"client_ip"`
	UserAgent string `json:"user_agent"`
}
