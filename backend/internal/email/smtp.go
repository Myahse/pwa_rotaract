package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/smtp"
	"strings"
)

type Config struct {
	Enabled          bool
	Host             string
	Port             string
	Username         string
	Password         string
	From             string
	FromName         string
	BrevoAPIKey      string
	BrevoSenderEmail string
	BrevoSenderName  string
}

type Client struct {
	cfg    Config
	logger *slog.Logger
}

func NewClient(cfg Config, logger *slog.Logger) *Client {
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
		c.logger.Info("email invite (smtp disabled, logged only)",
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
		c.logger.Info("club access email (smtp disabled, logged only)",
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
	if c.cfg.BrevoAPIKey != "" {
		return c.sendBrevo(to, subject, htmlBody)
	}

	from := c.cfg.From
	if c.cfg.FromName != "" {
		from = fmt.Sprintf("%s <%s>", c.cfg.FromName, c.cfg.From)
	}

	var msg bytes.Buffer
	msg.WriteString(fmt.Sprintf("From: %s\r\n", from))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", to))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(htmlBody)

	addr := fmt.Sprintf("%s:%s", c.cfg.Host, c.cfg.Port)
	auth := smtp.PlainAuth("", c.cfg.Username, c.cfg.Password, c.cfg.Host)

	if err := smtp.SendMail(addr, auth, c.cfg.From, []string{to}, msg.Bytes()); err != nil {
		return fmt.Errorf("send email: %w", err)
	}

	c.logger.Info("email sent", "to", to, "subject", subject)
	return nil
}

func (c *Client) sendBrevo(to, subject, htmlBody string) error {
	senderEmail := c.cfg.BrevoSenderEmail
	if senderEmail == "" {
		senderEmail = c.cfg.From
	}
	senderName := c.cfg.BrevoSenderName
	if senderName == "" {
		senderName = c.cfg.FromName
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
	payload.Sender.Email = senderEmail
	payload.Sender.Name = senderName
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
