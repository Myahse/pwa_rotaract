package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Env         string
	Host        string
	Port        string
	CORSOrigins []string
	LogLevel    slog.Level

	DatabaseURL      string
	JWTSecret        string
	JWTAccessTTL     time.Duration
	PasswordResetTTL time.Duration

	BootstrapAdminEmail    string
	BootstrapAdminPassword string

	AppPublicURL  string
	APIPublicURL  string
	InviteTTL     time.Duration
	UploadDir     string
	MaxAvatarSize int64

	BirthdayTimezone   string
	BirthdayNotifyHour int
	CronSecret         string

	VAPID VAPIDConfig

	SMTP  SMTPConfig
	Brevo BrevoConfig

	GoogleClientID string
}

type VAPIDConfig struct {
	PublicKey  string
	PrivateKey string
	Subject    string
}

type SMTPConfig struct {
	Enabled  bool
	Host     string
	Port     string
	Username string
	Password string
	From     string
	FromName string
}

type BrevoConfig struct {
	APIKey      string
	SenderEmail string
	SenderName  string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Env:                    getEnv("APP_ENV", "development"),
		Host:                   getEnv("HOST", "0.0.0.0"),
		Port:                   getEnv("PORT", "8088"),
		CORSOrigins:            splitCSV(getEnv("CORS_ORIGINS", "http://localhost:5173")),
		DatabaseURL:            os.Getenv("DATABASE_URL"),
		JWTSecret:              getEnv("JWT_SECRET", "dev-change-me"),
		BootstrapAdminEmail:    os.Getenv("BOOTSTRAP_ADMIN_EMAIL"),
		BootstrapAdminPassword: os.Getenv("BOOTSTRAP_ADMIN_PASSWORD"),
		AppPublicURL:           getEnv("APP_PUBLIC_URL", "http://localhost:5173"),
		APIPublicURL:           getEnv("API_PUBLIC_URL", "http://localhost:8088"),
		UploadDir:              getEnv("UPLOAD_DIR", "./uploads"),
		MaxAvatarSize:          parseInt64(getEnv("MAX_AVATAR_SIZE", "5242880")),
		BirthdayTimezone:       getEnv("BIRTHDAY_TIMEZONE", "Africa/Abidjan"),
		BirthdayNotifyHour:     parseInt(getEnv("BIRTHDAY_NOTIFY_HOUR", "8")),
		CronSecret:             os.Getenv("CRON_SECRET"),
		VAPID: VAPIDConfig{
			PublicKey:  os.Getenv("VAPID_PUBLIC_KEY"),
			PrivateKey: os.Getenv("VAPID_PRIVATE_KEY"),
			Subject:    getEnv("VAPID_SUBJECT", "mailto:admin@rotaract-civ.local"),
		},
		SMTP: SMTPConfig{
			Enabled:  os.Getenv("SMTP_HOST") != "",
			Host:     os.Getenv("SMTP_HOST"),
			Port:     getEnv("SMTP_PORT", "587"),
			Username: os.Getenv("SMTP_USER"),
			Password: os.Getenv("SMTP_PASSWORD"),
			From:     getEnv("SMTP_FROM", "noreply@rotaract-civ.local"),
			FromName: getEnv("SMTP_FROM_NAME", "Rotaract CIV"),
		},
		Brevo: BrevoConfig{
			APIKey:      os.Getenv("BREVO_API_KEY"),
			SenderEmail: os.Getenv("BREVO_SENDER_EMAIL"),
			SenderName:  getEnv("BREVO_SENDER_NAME", "Rotaract CIV"),
		},
		// Deferred: uncomment GOOGLE_CLIENT_ID in .env when enabling Google Sign-In.
		GoogleClientID: os.Getenv("GOOGLE_CLIENT_ID"),
	}

	ttl, err := time.ParseDuration(getEnv("JWT_ACCESS_TTL", "24h"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_ACCESS_TTL: %w", err)
	}
	cfg.JWTAccessTTL = ttl

	inviteTTL, err := time.ParseDuration(getEnv("INVITE_TTL", "168h"))
	if err != nil {
		return nil, fmt.Errorf("invalid INVITE_TTL: %w", err)
	}
	cfg.InviteTTL = inviteTTL

	resetTTL, err := time.ParseDuration(getEnv("PASSWORD_RESET_TTL", "1h"))
	if err != nil {
		return nil, fmt.Errorf("invalid PASSWORD_RESET_TTL: %w", err)
	}
	cfg.PasswordResetTTL = resetTTL

	level, err := parseLogLevel(getEnv("LOG_LEVEL", "info"))
	if err != nil {
		return nil, err
	}
	cfg.LogLevel = level

	return cfg, nil
}

func (c *Config) Addr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

func getEnv(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	return origins
}

func parseInt64(value string) int64 {
	var parsed int64
	fmt.Sscanf(strings.TrimSpace(value), "%d", &parsed)
	return parsed
}

func parseInt(value string) int {
	var parsed int
	fmt.Sscanf(strings.TrimSpace(value), "%d", &parsed)
	return parsed
}

func parseLogLevel(value string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("invalid LOG_LEVEL %q", value)
	}
}
