package mail

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"
)

//go:embed templates/*
var templateFS embed.FS

// TemplateData represents the data structure passed to email templates
type TemplateData struct {
	Name            string
	VerificationURL string
	ExpiresIn       string
	FromName        string
	FromEmail       string
	Subject         string
	// Add other common template fields as needed
}

// Mailer handles email template rendering and sending
type Mailer struct {
	client    *Client
	templates *template.Template
	fromEmail string
	fromName  string
}

// NewMailer creates a new Mailer instance with template support
func NewMailer(client *Client, fromEmail string, fromName string) (*Mailer, error) {
	// Parse all email templates
	templates, err := template.ParseFS(templateFS, "templates/*.html", "templates/*.txt")
	if err != nil {
		return nil, fmt.Errorf("parsing email templates: %w", err)
	}

	return &Mailer{
		client:    client,
		templates: templates,
		fromEmail: fromEmail,
		fromName:  fromName,
	}, nil
}

// SendTemplate renders and sends an email using the specified template
func (m *Mailer) SendTemplate(ctx context.Context, templateName string, to string, data interface{}) error {
	// Render HTML version
	htmlTemplate := templateName + ".html"
	var htmlBody bytes.Buffer
	if err := m.templates.ExecuteTemplate(&htmlBody, htmlTemplate, data); err != nil {
		return fmt.Errorf("rendering HTML template %s: %w", htmlTemplate, err)
	}

	// Render text version (if exists)
	textTemplate := templateName + ".txt"
	var textBody bytes.Buffer
	if err := m.templates.ExecuteTemplate(&textBody, textTemplate, data); err != nil {
		// Text template is optional, only log error if template exists but fails
		if !isTemplateNotFound(err) {
			return fmt.Errorf("rendering text template %s: %w", textTemplate, err)
		}
	}

	// Create email
	email := Email{
		From:     m.fromEmail,
		To:       to,
		Subject:  getSubjectFromData(data),
		HtmlBody: htmlBody.String(),
	}

	// Add text body if available
	if textBody.Len() > 0 {
		email.TextBody = textBody.String()
	}

	// Send email
	_, err := m.client.SendEmail(email)
	if err != nil {
		return fmt.Errorf("sending template email: %w", err)
	}

	return nil
}

// Helper function to check if error is template not found
func isTemplateNotFound(err error) bool {
	return err != nil && err.Error() == "template not found"
}

// Helper function to extract subject from template data
func getSubjectFromData(data interface{}) string {
	// Try to get subject from TemplateData
	if td, ok := data.(TemplateData); ok && td.Subject != "" {
		return td.Subject
	}

	// Try to get subject from map
	if m, ok := data.(map[string]interface{}); ok {
		if subject, ok := m["Subject"].(string); ok {
			return subject
		}
	}

	// Default subject if none provided
	return "Important Notification"
}
