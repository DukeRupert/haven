package mail

import (
	"fmt"
	"time"
)

type Email struct {
	From        string            `json:"From,omitempty"`
	To          string            `json:"To,omitempty"`
	Cc          string            `json:"Cc,omitempty"`
	Bcc         string            `json:"Bcc,omitempty"`
	Subject     string            `json:"Subject,omitempty"`
	Tag         string            `json:"Tag,omitempty"`
	HtmlBody    string            `json:"HtmlBody,omitempty"`
	TextBody    string            `json:"TextBody,omitempty"`
	ReplyTo     string            `json:"ReplyTo,omitempty"`
	Headers     []Header          `json:"Headers,omitempty"`
	TrackOpens  bool              `json:"TrackOpens,omitempty"`
	Attachments []Attachment      `json:"Attachments,omitempty"`
	Metadata    map[string]string `json:"Metadata,omitempty"`
}

type Header struct {
	Name  string `json:"Name"`
	Value string `json:"Value"`
}

type Attachment struct {
	Name        string `json:"Name"`
	Content     string `json:"Content"`
	ContentType string `json:"ContentType"`
	ContentID   string `json:"ContentID,omitempty"`
}

type EmailResponse struct {
	To          string    `json:"To"`
	SubmittedAt time.Time `json:"SubmittedAt"`
	MessageID   string    `json:"MessageID"`
	ErrorCode   int       `json:"ErrorCode"`
	Message     string    `json:"Message"`
}

func (c *Client) SendEmail(email Email) (*EmailResponse, error) {
	if err := validateEmail(email); err != nil {
		return nil, fmt.Errorf("validating email: %w", err)
	}

	var response EmailResponse
	err := c.doRequest(requestParams{
		method:    "POST",
		path:      "email",
		payload:   email,
		tokenType: TokenTypeServer,
	}, &response)
	if err != nil {
		return nil, fmt.Errorf("sending email: %w", err)
	}

	if response.ErrorCode != 0 {
		return &response, &ErrorResponse{
			ErrorCode: response.ErrorCode,
			Message:   response.Message,
		}
	}

	return &response, nil
}

func (c *Client) SendEmailBatch(emails []Email) ([]EmailResponse, error) {
	if len(emails) == 0 {
		return nil, fmt.Errorf("email batch is empty")
	}

	for i, email := range emails {
		if err := validateEmail(email); err != nil {
			return nil, fmt.Errorf("validating email at index %d: %w", i, err)
		}
	}

	var responses []EmailResponse
	err := c.doRequest(requestParams{
		method:    "POST",
		path:      "email/batch",
		payload:   emails,
		tokenType: TokenTypeServer,
	}, &responses)
	if err != nil {
		return nil, fmt.Errorf("sending email batch: %w", err)
	}

	return responses, nil
}

func validateEmail(email Email) error {
	if email.From == "" {
		return fmt.Errorf("From address is required")
	}
	if email.To == "" {
		return fmt.Errorf("To address is required")
	}
	if email.HtmlBody == "" && email.TextBody == "" {
		return fmt.Errorf("either HtmlBody or TextBody is required")
	}
	return nil
}