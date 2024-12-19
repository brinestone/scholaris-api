package dto

type SendEmailRequest struct {
	To            string            `json:"to"`
	Body          string            `json:"body"`
	Subject       string            `json:"subject"`
	ToName        string            `json:"toName"`
	IsContentHtml bool              `json:"html"`
	TemplateId    string            `json:"templateId,omitempty"`
	Data          map[string]string `json:"data,omitempty"`
}
