package mail

import (
    "bytes"
    "context"
    "embed"
    "fmt"
    "text/template"
)

//go:embed templates/*.txt
var templateFS embed.FS

// Mailer handles email template rendering and sending
type Mailer struct {
    client    *Client
    templates *template.Template
    fromEmail string
    fromName  string
}

// NewMailer creates a new Mailer instance
func NewMailer(client *Client, fromEmail, fromName string) (*Mailer, error) {
    // Parse all text email templates
    templates, err := template.ParseFS(templateFS, "templates/*.txt")
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

// SendTemplate renders and sends a text email using the specified template
func (m *Mailer) SendTemplate(ctx context.Context, templateName string, to string, data interface{}) error {
    // Render text template
    var body bytes.Buffer
    if err := m.templates.ExecuteTemplate(&body, templateName+".txt", data); err != nil {
        return fmt.Errorf("rendering template %s: %w", templateName, err)
    }

    // Create email
    email := Email{
        From:     fmt.Sprintf("%s <%s>", m.fromName, m.fromEmail),
        To:       to,
        Subject:  getSubjectFromData(data),
        TextBody: body.String(),
    }

    // Send email
    _, err := m.client.SendEmail(email)
    if err != nil {
        return fmt.Errorf("sending template email: %w", err)
    }

    return nil
}

// Helper function to extract subject from template data
func getSubjectFromData(data interface{}) string {
    // Try to get subject from map
    if m, ok := data.(map[string]interface{}); ok {
        if subject, ok := m["Subject"].(string); ok {
            return subject
        }
    }
    
    // Default subject if none provided
    return "Important Notification"
}