package email

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/smtp"
	"strings"
)

type Config struct {
	Enabled  bool
	Host     string
	Port     string
	Username string
	Password string
	From     string
	FromName string
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

func buildRegistrationInviteBody(invite RegistrationInvite) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="fr">
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #222;">
  <h2>Bienvenue sur Rotaract CIV</h2>
  <p>Vous avez été invité(e) à rejoindre <strong>%s</strong>.</p>
  <p>Cliquez sur le lien ci-dessous pour créer votre compte et compléter votre inscription&nbsp;:</p>
  <p><a href="%s" style="display:inline-block;padding:12px 20px;background:#be034d;color:#fff;text-decoration:none;border-radius:6px;">Créer mon compte</a></p>
  <p>Ou copiez ce lien dans votre navigateur&nbsp;:<br><a href="%s">%s</a></p>
  <p style="color:#666;font-size:14px;">Ce lien expire le %s.</p>
  <p style="color:#666;font-size:14px;">Si vous n'attendiez pas cet email, vous pouvez l'ignorer.</p>
</body>
</html>`,
		escapeHTML(invite.ClubName),
		invite.RegisterURL,
		invite.RegisterURL,
		invite.RegisterURL,
		escapeHTML(invite.ExpiresAt),
	)
}

func buildClubAccessBody(access ClubAccessEmail) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="fr">
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #222;">
  <h2>Votre club est approuvé</h2>
  <p>Bonjour %s,</p>
  <p>La demande d'inscription du club <strong>%s</strong> a été acceptée sur Rotaract CIV.</p>
  <p>Cliquez sur le lien ci-dessous pour finaliser votre inscription avec Google&nbsp;:</p>
  <p><a href="%s" style="display:inline-block;padding:12px 20px;background:#be034d;color:#fff;text-decoration:none;border-radius:6px;">Finaliser avec Google</a></p>
  <p>Ou copiez ce lien&nbsp;:<br><a href="%s">%s</a></p>
  <p style="color:#666;font-size:14px;">Utilisez le compte Google correspondant à l'email de la demande. Ce lien expire le %s.</p>
</body>
</html>`,
		escapeHTML(access.ContactName),
		escapeHTML(access.ClubName),
		access.RegisterURL,
		access.RegisterURL,
		access.RegisterURL,
		escapeHTML(access.ExpiresAt),
	)
}

func escapeHTML(value string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
	)
	return replacer.Replace(value)
}
