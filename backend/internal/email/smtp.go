package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

type Config struct {
	Enabled          bool
	BrevoAPIKey      string
	BrevoSenderEmail string
	BrevoSenderName  string
}

type Client struct {
	cfg    Config
	logger *slog.Logger
}

func NewClient(cfg Config, logger *slog.Logger) *Client {
	if logger == nil {
		logger = slog.Default()
	}
	return &Client{cfg: cfg, logger: logger}
}

type RegistrationInvite struct {
	To          string
	ClubName    string
	RegisterURL string
	ExpiresAt   string
}

func (c *Client) SendRegistrationInvite(invite RegistrationInvite) error {
	subject := fmt.Sprintf("Invitation Rotaract CIV — %s", invite.ClubName)
	body := buildRegistrationInviteBody(invite)

	if !c.cfg.Enabled {
		c.logger.Info("email invite (delivery disabled, logged only)",
			"to", invite.To,
			"subject", subject,
			"register_url", invite.RegisterURL,
		)
		return nil
	}

	return c.send(invite.To, subject, body)
}

type ClubAccessEmail struct {
	To          string
	ClubName    string
	ContactName string
	RegisterURL string
	ExpiresAt   string
}

type PasswordResetEmail struct {
	To        string
	FirstName string
	ResetURL  string
	ExpiresAt string
}

func (c *Client) SendPasswordReset(reset PasswordResetEmail) error {
	subject := "Réinitialisez votre mot de passe — Rotaract CIV"
	body := renderTemplate("password-reset.html", reset)
	if !c.cfg.Enabled {
		c.logger.Info("password reset email (delivery disabled, logged only)", "to", reset.To, "reset_url", reset.ResetURL)
		return nil
	}
	return c.send(reset.To, subject, body)
}

func (c *Client) SendClubAccess(access ClubAccessEmail) error {
	subject := fmt.Sprintf("Finalisez l'inscription de %s — Rotaract CIV", access.ClubName)
	body := buildClubAccessBody(access)

	if !c.cfg.Enabled {
		c.logger.Info("club access email (delivery disabled, logged only)",
			"to", access.To,
			"subject", subject,
			"register_url", access.RegisterURL,
			"expires_at", access.ExpiresAt,
		)
		return nil
	}

	if err := c.send(access.To, subject, body); err != nil {
		return err
	}
	c.logger.Info("club access email sent", "to", access.To, "club", access.ClubName)
	return nil
}

func (c *Client) send(to, subject, htmlBody string) error {
	if c.cfg.BrevoAPIKey == "" {
		return fmt.Errorf("BREVO_API_KEY is required for email delivery")
	}
	if c.cfg.BrevoSenderEmail == "" {
		return fmt.Errorf("BREVO_SENDER_EMAIL is required for email delivery")
	}

	payload := struct {
		Sender struct {
			Email string `json:"email"`
			Name  string `json:"name,omitempty"`
		} `json:"sender"`
		To []struct {
			Email string `json:"email"`
		} `json:"to"`
		Subject     string `json:"subject"`
		HTMLContent string `json:"htmlContent"`
	}{
		Subject:     subject,
		HTMLContent: htmlBody,
	}
	payload.Sender.Email = c.cfg.BrevoSenderEmail
	payload.Sender.Name = c.cfg.BrevoSenderName
	payload.To = []struct {
		Email string `json:"email"`
	}{{Email: to}}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode Brevo email: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, "https://api.brevo.com/v3/smtp/email", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create Brevo email request: %w", err)
	}
	req.Header.Set("accept", "application/json")
	req.Header.Set("api-key", c.cfg.BrevoAPIKey)
	req.Header.Set("content-type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("send Brevo email: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		responseBody, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
		return fmt.Errorf("Brevo email rejected with status %d: %s", res.StatusCode, strings.TrimSpace(string(responseBody)))
	}

	c.logger.Info("email sent via Brevo", "to", to, "subject", subject)
	return nil
}

func buildRegistrationInviteBody(invite RegistrationInvite) string {
	return renderTemplate("registration-invite.html", invite)
}

func buildClubAccessBody(access ClubAccessEmail) string {
	return renderTemplate("club-access.html", access)
}

func renderTemplate(name string, data any) string {
	var body bytes.Buffer
	if err := emailTemplates.ExecuteTemplate(&body, name, data); err != nil {
		return ""
	}
	return body.String()
}
