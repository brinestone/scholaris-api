package dto

type SendEmailRequest struct {
	To            string `json:"to"`
	Body          string `json:"body"`
	Subject       string `json:"subject"`
	IsContentHtml bool   `json:"html"`
}
