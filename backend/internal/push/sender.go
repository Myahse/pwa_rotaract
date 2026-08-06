package push

import (
	"encoding/json"
	"fmt"

	webpush "github.com/SherClockHolmes/webpush-go"
)

type VAPIDConfig struct {
	PublicKey  string
	PrivateKey string
	Subject    string
}

type Subscription struct {
	Endpoint string
	P256dh   string
	Auth     string
}

type NotificationPayload struct {
	Title  string         `json:"title"`
	Body   string         `json:"body"`
	Icon   string         `json:"icon,omitempty"`
	Badge  string         `json:"badge,omitempty"`
	Tag    string         `json:"tag,omitempty"`
	Data   map[string]any `json:"data,omitempty"`
	Silent bool           `json:"silent,omitempty"`
}

type Sender struct {
	cfg VAPIDConfig
}

func NewSender(cfg VAPIDConfig) *Sender {
	return &Sender{cfg: cfg}
}

func (s *Sender) Enabled() bool {
	return s.cfg.PublicKey != "" && s.cfg.PrivateKey != "" && s.cfg.Subject != ""
}

func (s *Sender) PublicKey() string {
	return s.cfg.PublicKey
}

func (s *Sender) Send(sub Subscription, payload NotificationPayload) error {
	if !s.Enabled() {
		return fmt.Errorf("web push is not configured")
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal push payload: %w", err)
	}

	resp, err := webpush.SendNotification(body, &webpush.Subscription{
		Endpoint: sub.Endpoint,
		Keys: webpush.Keys{
			P256dh: sub.P256dh,
			Auth:   sub.Auth,
		},
	}, &webpush.Options{
		Subscriber:      s.cfg.Subject,
		VAPIDPublicKey:  s.cfg.PublicKey,
		VAPIDPrivateKey: s.cfg.PrivateKey,
		TTL:             86400,
	})
	if err != nil {
		return fmt.Errorf("send push notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("push provider returned status %d", resp.StatusCode)
	}

	return nil
}
