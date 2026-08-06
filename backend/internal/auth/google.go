package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type GoogleIdentity struct {
	Sub           string
	Email         string
	EmailVerified bool
	GivenName     string
	FamilyName    string
}

type googleTokenInfo struct {
	Aud           string `json:"aud"`
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified string `json:"email_verified"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Error         string `json:"error"`
	ErrorDesc     string `json:"error_description"`
}

func VerifyGoogleIDToken(ctx context.Context, idToken, clientID string) (*GoogleIdentity, error) {
	clientID = strings.TrimSpace(clientID)
	if clientID == "" {
		return nil, fmt.Errorf("google client id is not configured")
	}
	idToken = strings.TrimSpace(idToken)
	if idToken == "" {
		return nil, fmt.Errorf("google id token is required")
	}

	endpoint := "https://oauth2.googleapis.com/tokeninfo?id_token=" + url.QueryEscape(idToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("verify google id token: %w", err)
	}
	defer res.Body.Close()

	var info googleTokenInfo
	if err := json.NewDecoder(res.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("decode google token info: %w", err)
	}
	if res.StatusCode != http.StatusOK || info.Error != "" {
		msg := info.ErrorDesc
		if msg == "" {
			msg = info.Error
		}
		if msg == "" {
			msg = res.Status
		}
		return nil, fmt.Errorf("invalid google id token: %s", msg)
	}
	if info.Aud != clientID {
		return nil, fmt.Errorf("google id token audience mismatch")
	}

	email := strings.ToLower(strings.TrimSpace(info.Email))
	if email == "" {
		return nil, fmt.Errorf("google account email is missing")
	}
	if !strings.EqualFold(info.EmailVerified, "true") {
		return nil, fmt.Errorf("google account email is not verified")
	}
	if strings.TrimSpace(info.Sub) == "" {
		return nil, fmt.Errorf("google account subject is missing")
	}

	return &GoogleIdentity{
		Sub:           info.Sub,
		Email:         email,
		EmailVerified: true,
		GivenName:     strings.TrimSpace(info.GivenName),
		FamilyName:    strings.TrimSpace(info.FamilyName),
	}, nil
}
