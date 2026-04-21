package email

import (
	"fmt"
	"net/smtp"
	"strings"

	"dermify-api/config"
)

// Client sends emails via SMTP.
type Client struct {
	cfg config.SMTPConfig
}

// InvitationOptions allows sender customization per organization.
type InvitationOptions struct {
	FromEmail string
	FromName  string
}

// NewClient creates an email client from SMTP configuration.
func NewClient(cfg config.SMTPConfig) *Client {
	return &Client{cfg: cfg}
}

// Enabled returns true if SMTP is configured (host is non-empty and not localhost).
func (c *Client) Enabled() bool {
	return c.cfg.Host != "" && c.cfg.Host != "localhost"
}

// SendInvitation sends an organization invitation email.
func (c *Client) SendInvitation(to, orgName, inviterName, token string, opts *InvitationOptions) error {
	acceptURL := fmt.Sprintf("%s/invite/%s", c.cfg.FrontendURL, token)

	subject := fmt.Sprintf("You've been invited to join %s on Dermify", orgName)
	body := buildInvitationHTML(orgName, inviterName, acceptURL)

	return c.send(to, subject, body, opts)
}

func (c *Client) send(to, subject, htmlBody string, opts *InvitationOptions) error {
	if !c.Enabled() {
		return nil
	}

	fromEmail := c.cfg.FromEmail
	fromName := c.cfg.FromName
	if opts != nil {
		if opts.FromEmail != "" {
			fromEmail = opts.FromEmail
		}
		if opts.FromName != "" {
			fromName = opts.FromName
		}
	}

	from := fmt.Sprintf("%s <%s>", fromName, fromEmail)

	headers := []string{
		fmt.Sprintf("From: %s", from),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=UTF-8",
	}

	msg := []byte(strings.Join(headers, "\r\n") + "\r\n\r\n" + htmlBody)

	addr := fmt.Sprintf("%s:%d", c.cfg.Host, c.cfg.Port)

	var auth smtp.Auth
	if c.cfg.Username != "" {
		auth = smtp.PlainAuth("", c.cfg.Username, c.cfg.Password, c.cfg.Host)
	}

	return smtp.SendMail(addr, auth, fromEmail, []string{to}, msg)
}

func buildInvitationHTML(orgName, inviterName, acceptURL string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; margin: 0; padding: 0; background-color: #f9fafb;">
  <div style="max-width: 560px; margin: 40px auto; background: #fff; border-radius: 8px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); padding: 40px;">
    <h1 style="color: #2563eb; font-size: 24px; margin: 0 0 8px;">Dermify</h1>
    <h2 style="color: #111827; font-size: 20px; margin: 0 0 24px;">You're invited to join %s</h2>
    <p style="color: #4b5563; font-size: 16px; line-height: 1.5;">
      <strong>%s</strong> has invited you to join their organization on Dermify.
    </p>
    <div style="text-align: center; margin: 32px 0;">
      <a href="%s" style="display: inline-block; background-color: #2563eb; color: #fff; text-decoration: none; padding: 12px 32px; border-radius: 6px; font-size: 16px; font-weight: 600;">
        Accept Invitation
      </a>
    </div>
    <p style="color: #6b7280; font-size: 14px;">
      If you don't have a Dermify account, you'll be asked to create one first.
    </p>
    <hr style="border: none; border-top: 1px solid #e5e7eb; margin: 24px 0;">
    <p style="color: #9ca3af; font-size: 12px;">
      If you didn't expect this invitation, you can safely ignore this email.
    </p>
  </div>
</body>
</html>`, orgName, inviterName, acceptURL)
}
