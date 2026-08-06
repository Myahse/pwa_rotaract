package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

const apiPrefix = "/api/v1"

type client struct {
	base string
	http *http.Client
}

func newClient(base string) *client {
	return &client{
		base: strings.TrimRight(base, "/"),
		http: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *client) request(method, path string, token string, body any) (int, []byte, error) {
	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return 0, nil, err
		}
		reader = bytes.NewReader(payload)
	}

	req, err := http.NewRequest(method, c.base+path, reader)
	if err != nil {
		return 0, nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	return resp.StatusCode, data, err
}

func (c *client) expectOK(step, method, path, token string, body any, allowed ...int) ([]byte, error) {
	status, data, err := c.request(method, path, token, body)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", step, err)
	}
	ok := false
	for _, code := range allowed {
		if status == code {
			ok = true
			break
		}
	}
	if !ok {
		return nil, fmt.Errorf("%s: expected %v got %d body=%s", step, allowed, status, truncate(data, 400))
	}
	return data, nil
}

func truncate(b []byte, n int) string {
	s := string(b)
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func decode(data []byte, out any) error {
	return json.Unmarshal(data, out)
}

type loginResponse struct {
	AccessToken string `json:"access_token"`
	User        struct {
		ID uuid.UUID `json:"id"`
	} `json:"user"`
}

func login(c *client, email, password string) (string, uuid.UUID, error) {
	data, err := c.expectOK("login", http.MethodPost, apiPrefix+"/auth/login", "", map[string]string{
		"email":    email,
		"password": password,
	}, http.StatusOK)
	if err != nil {
		return "", uuid.Nil, err
	}
	var res loginResponse
	if err := decode(data, &res); err != nil {
		return "", uuid.Nil, err
	}
	return res.AccessToken, res.User.ID, nil
}

func main() {
	_ = godotenv.Load()

	base := envOr("SMOKE_BASE_URL", "http://localhost:8088")
	adminEmail := envOr("BOOTSTRAP_ADMIN_EMAIL", "admin@rotaract-civ.local")
	adminPassword := envOr("BOOTSTRAP_ADMIN_PASSWORD", "change-me")
	cronSecret := os.Getenv("CRON_SECRET")

	runID := fmt.Sprintf("%d", time.Now().Unix())
	headEmail := fmt.Sprintf("head-%s@smoke.test", runID)
	memberEmail := fmt.Sprintf("member-%s@smoke.test", runID)
	accessEmail := fmt.Sprintf("access-%s@smoke.test", runID)
	password := "SmokeTest123!"

	c := newClient(base)
	steps := 0
	pass := func(name string) {
		steps++
		fmt.Printf("  ok  %s\n", name)
	}

	fmt.Printf("Smoke test against %s (run %s)\n", base, runID)

	must := func(name string, fn func() error) {
		if err := fn(); err != nil {
			fmt.Printf("FAIL %s: %v\n", name, err)
			os.Exit(1)
		}
		pass(name)
	}

	must("GET /health", func() error {
		_, err := c.expectOK("health", http.MethodGet, "/health", "", nil, http.StatusOK)
		return err
	})
	must("GET /ready", func() error {
		_, err := c.expectOK("ready", http.MethodGet, "/ready", "", nil, http.StatusOK, http.StatusServiceUnavailable)
		return err
	})
	must("GET /api/v1/status", func() error {
		_, err := c.expectOK("status", http.MethodGet, apiPrefix+"/status", "", nil, http.StatusOK)
		return err
	})
	must("GET /api/v1/push/vapid-key", func() error {
		_, err := c.expectOK("vapid", http.MethodGet, apiPrefix+"/push/vapid-key", "", nil, http.StatusOK)
		return err
	})

	var adminToken string
	must("POST /auth/login (admin)", func() error {
		var err error
		adminToken, _, err = login(c, adminEmail, adminPassword)
		return err
	})

	var clubID uuid.UUID
	var inviteCode string
	must("POST /admin/clubs", func() error {
		data, err := c.expectOK("create club", http.MethodPost, apiPrefix+"/admin/clubs", adminToken, map[string]string{
			"name":    "Smoke Club " + runID,
			"slug":    "smoke-" + runID,
			"country": "Côte d'Ivoire",
			"city":    "Abidjan",
			"commune": "Cocody",
		}, http.StatusCreated)
		if err != nil {
			return err
		}
		var club struct {
			ID         uuid.UUID `json:"id"`
			InviteCode string    `json:"invite_code"`
		}
		if err := decode(data, &club); err != nil {
			return err
		}
		clubID = club.ID
		inviteCode = club.InviteCode
		return nil
	})
	must("GET /admin/clubs", func() error {
		_, err := c.expectOK("list clubs", http.MethodGet, apiPrefix+"/admin/clubs", adminToken, nil, http.StatusOK)
		return err
	})
	must("PATCH /admin/clubs/{id}", func() error {
		_, err := c.expectOK("update club", http.MethodPatch, fmt.Sprintf("%s/admin/clubs/%s", apiPrefix, clubID), adminToken, map[string]string{
			"name": "Smoke Club Updated " + runID,
		}, http.StatusOK)
		return err
	})
	must("POST /admin/clubs/{id}/head", func() error {
		_, err := c.expectOK("create head", http.MethodPost, fmt.Sprintf("%s/admin/clubs/%s/head", apiPrefix, clubID), adminToken, map[string]string{
			"email":      headEmail,
			"password":   password,
			"first_name": "Smoke",
			"last_name":  "Head",
		}, http.StatusCreated)
		return err
	})

	var headToken string
	var headID uuid.UUID
	must("POST /auth/login (head)", func() error {
		var err error
		headToken, headID, err = login(c, headEmail, password)
		return err
	})

	must("GET /auth/me", func() error {
		_, err := c.expectOK("auth me", http.MethodGet, apiPrefix+"/auth/me", headToken, nil, http.StatusOK)
		return err
	})
	must("PATCH /users/me", func() error {
		_, err := c.expectOK("update profile", http.MethodPatch, apiPrefix+"/users/me", headToken, map[string]string{
			"profession": "Smoke tester",
		}, http.StatusOK)
		return err
	})

	var memberRoleID uuid.UUID
	must("GET /invite/{code}", func() error {
		data, err := c.expectOK("preview invite", http.MethodGet, apiPrefix+"/invite/"+inviteCode, "", nil, http.StatusOK)
		if err != nil {
			return err
		}
		var preview struct {
			Roles []struct {
				ID   uuid.UUID `json:"id"`
				Name string    `json:"name"`
			} `json:"roles"`
		}
		if err := decode(data, &preview); err != nil {
			return err
		}
		for _, role := range preview.Roles {
			if role.Name == "Membre" {
				memberRoleID = role.ID
				break
			}
		}
		if memberRoleID == uuid.Nil {
			return fmt.Errorf("Membre role not found in invite preview")
		}
		return nil
	})

	var memberID uuid.UUID
	must("POST /register (invite code)", func() error {
		data, err := c.expectOK("register", http.MethodPost, apiPrefix+"/register", "", map[string]any{
			"email":       memberEmail,
			"password":    password,
			"first_name":  "Smoke",
			"last_name":   "Member",
			"invite_code": inviteCode,
			"role_id":     memberRoleID,
		}, http.StatusCreated)
		if err != nil {
			return err
		}
		var user struct {
			ID uuid.UUID `json:"id"`
		}
		if err := decode(data, &user); err != nil {
			return err
		}
		memberID = user.ID
		return nil
	})

	var memberToken string
	must("POST /auth/login (member)", func() error {
		var err error
		memberToken, _, err = login(c, memberEmail, password)
		return err
	})

	must("GET /clubs/{id}", func() error {
		_, err := c.expectOK("get club", http.MethodGet, fmt.Sprintf("%s/clubs/%s/", apiPrefix, clubID), headToken, nil, http.StatusOK)
		return err
	})
	must("GET /clubs/{id}/members", func() error {
		_, err := c.expectOK("list members", http.MethodGet, fmt.Sprintf("%s/clubs/%s/members", apiPrefix, clubID), headToken, nil, http.StatusOK)
		return err
	})
	must("GET /clubs/{id}/roles", func() error {
		_, err := c.expectOK("list roles", http.MethodGet, fmt.Sprintf("%s/clubs/%s/roles", apiPrefix, clubID), headToken, nil, http.StatusOK)
		return err
	})
	must("GET /clubs/{id}/invite", func() error {
		_, err := c.expectOK("club invite", http.MethodGet, fmt.Sprintf("%s/clubs/%s/invite", apiPrefix, clubID), headToken, nil, http.StatusOK)
		return err
	})
	must("POST /clubs/{id}/email-invites", func() error {
		_, err := c.expectOK("send email invite", http.MethodPost, fmt.Sprintf("%s/clubs/%s/email-invites", apiPrefix, clubID), headToken, map[string]string{
			"email": fmt.Sprintf("invited-%s@smoke.test", runID),
		}, http.StatusAccepted)
		return err
	})
	var emailInviteID uuid.UUID
	must("GET /clubs/{id}/email-invites", func() error {
		data, err := c.expectOK("list email invites", http.MethodGet, fmt.Sprintf("%s/clubs/%s/email-invites", apiPrefix, clubID), headToken, nil, http.StatusOK)
		if err != nil {
			return err
		}
		var invites []struct {
			ID    uuid.UUID `json:"id"`
			Email string    `json:"email"`
		}
		if err := decode(data, &invites); err != nil {
			return err
		}
		target := fmt.Sprintf("invited-%s@smoke.test", runID)
		for _, inv := range invites {
			if inv.Email == target {
				emailInviteID = inv.ID
				break
			}
		}
		if emailInviteID == uuid.Nil {
			return fmt.Errorf("email invite not found for %s", target)
		}
		return nil
	})
	must("DELETE /clubs/{id}/email-invites/{iid}", func() error {
		_, err := c.expectOK("revoke email invite", http.MethodDelete,
			fmt.Sprintf("%s/clubs/%s/email-invites/%s", apiPrefix, clubID, emailInviteID),
			headToken, nil, http.StatusOK)
		return err
	})
	must("GET /clubs/{id}/birthdays/today", func() error {
		_, err := c.expectOK("club birthdays", http.MethodGet, fmt.Sprintf("%s/clubs/%s/birthdays/today", apiPrefix, clubID), headToken, nil, http.StatusOK)
		return err
	})

	var commissionID uuid.UUID
	must("POST /clubs/{id}/commissions", func() error {
		data, err := c.expectOK("create commission", http.MethodPost, fmt.Sprintf("%s/clubs/%s/commissions", apiPrefix, clubID), headToken, map[string]string{
			"name": "Commission Smoke " + runID,
		}, http.StatusCreated)
		if err != nil {
			return err
		}
		var commission struct {
			ID uuid.UUID `json:"id"`
		}
		if err := decode(data, &commission); err != nil {
			return err
		}
		commissionID = commission.ID
		return nil
	})
	must("GET /clubs/{id}/commissions", func() error {
		_, err := c.expectOK("list commissions", http.MethodGet, fmt.Sprintf("%s/clubs/%s/commissions", apiPrefix, clubID), headToken, nil, http.StatusOK)
		return err
	})
	must("POST /clubs/{id}/commissions/{cid}/members", func() error {
		_, err := c.expectOK("add commission president", http.MethodPost,
			fmt.Sprintf("%s/clubs/%s/commissions/%s/members", apiPrefix, clubID, commissionID),
			headToken,
			map[string]any{
				"user_id":     memberID,
				"member_role": "president",
			},
			http.StatusCreated)
		return err
	})
	must("GET /clubs/{id}/commissions/{cid}", func() error {
		_, err := c.expectOK("get commission", http.MethodGet,
			fmt.Sprintf("%s/clubs/%s/commissions/%s", apiPrefix, clubID, commissionID),
			headToken, nil, http.StatusOK)
		return err
	})
	must("PATCH /clubs/{id}/commissions/{cid}", func() error {
		_, err := c.expectOK("update commission", http.MethodPatch,
			fmt.Sprintf("%s/clubs/%s/commissions/%s", apiPrefix, clubID, commissionID),
			headToken, map[string]string{
				"name":        "Commission Updated " + runID,
				"description": "Updated by smoke test",
			}, http.StatusOK)
		return err
	})

	var generalGroupID, commissionGroupID uuid.UUID
	must("GET /clubs/{id}/chat/groups", func() error {
		data, err := c.expectOK("list chat groups", http.MethodGet, fmt.Sprintf("%s/clubs/%s/chat/groups", apiPrefix, clubID), headToken, nil, http.StatusOK)
		if err != nil {
			return err
		}
		var groups []struct {
			ID        uuid.UUID `json:"id"`
			GroupType string    `json:"group_type"`
		}
		if err := decode(data, &groups); err != nil {
			return err
		}
		for _, g := range groups {
			if g.GroupType == "club" {
				generalGroupID = g.ID
			}
		}
		if generalGroupID == uuid.Nil {
			return fmt.Errorf("general club chat group not found")
		}
		return nil
	})
	must("POST /clubs/{id}/commissions/{cid}/chat/groups", func() error {
		data, err := c.expectOK("create commission group", http.MethodPost,
			fmt.Sprintf("%s/clubs/%s/commissions/%s/chat/groups", apiPrefix, clubID, commissionID),
			memberToken,
			map[string]string{"name": "Commission chat"},
			http.StatusCreated)
		if err != nil {
			return err
		}
		var group struct {
			ID uuid.UUID `json:"id"`
		}
		if err := decode(data, &group); err != nil {
			return err
		}
		commissionGroupID = group.ID
		return nil
	})
	must("GET /chat/groups/{id}", func() error {
		_, err := c.expectOK("get group", http.MethodGet, fmt.Sprintf("%s/chat/groups/%s", apiPrefix, generalGroupID), memberToken, nil, http.StatusOK)
		return err
	})
	must("POST /chat/groups/{id}/messages", func() error {
		_, err := c.expectOK("send message", http.MethodPost, fmt.Sprintf("%s/chat/groups/%s/messages", apiPrefix, generalGroupID), memberToken, map[string]string{
			"content": "Hello from smoke test " + runID,
		}, http.StatusCreated)
		return err
	})
	var messageID uuid.UUID
	must("GET /chat/groups/{id}/messages", func() error {
		data, err := c.expectOK("list messages", http.MethodGet, fmt.Sprintf("%s/chat/groups/%s/messages", apiPrefix, generalGroupID), headToken, nil, http.StatusOK)
		if err != nil {
			return err
		}
		var res struct {
			Messages []struct {
				ID      uuid.UUID `json:"id"`
				Content string    `json:"content"`
			} `json:"messages"`
		}
		if err := decode(data, &res); err != nil {
			return err
		}
		if len(res.Messages) == 0 {
			return fmt.Errorf("expected at least one message")
		}
		messageID = res.Messages[0].ID
		return nil
	})
	must("PATCH /chat/groups/{id}/messages/{mid}", func() error {
		_, err := c.expectOK("edit message", http.MethodPatch,
			fmt.Sprintf("%s/chat/groups/%s/messages/%s", apiPrefix, generalGroupID, messageID),
			memberToken, map[string]string{
				"content": "Edited smoke message " + runID,
			}, http.StatusOK)
		return err
	})
	must("GET /ws/chat (websocket)", func() error {
		wsURL := strings.Replace(c.base, "http://", "ws://", 1)
		wsURL = strings.Replace(wsURL, "https://", "wss://", 1)
		wsURL += fmt.Sprintf("%s/ws/chat?token=%s&group_id=%s", apiPrefix, memberToken, generalGroupID)

		conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			if resp != nil {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("dial: %v body=%s", err, truncate(body, 200))
			}
			return err
		}
		defer conn.Close()

		_, err = c.expectOK("send ws message", http.MethodPost, fmt.Sprintf("%s/chat/groups/%s/messages", apiPrefix, generalGroupID), headToken, map[string]string{
			"content": "WS broadcast " + runID,
		}, http.StatusCreated)
		if err != nil {
			return err
		}

		_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		_, payload, err := conn.ReadMessage()
		if err != nil {
			return fmt.Errorf("read ws event: %w", err)
		}
		if !strings.Contains(string(payload), "WS broadcast") {
			return fmt.Errorf("unexpected ws payload: %s", truncate(payload, 200))
		}
		return nil
	})

	must("POST /access-requests", func() error {
		_, err := c.expectOK("access request", http.MethodPost, apiPrefix+"/access-requests", "", map[string]any{
			"email":      accessEmail,
			"password":   password,
			"first_name": "Access",
			"last_name":  "Request",
			"club_name":  "Smoke Club Updated " + runID,
			"role_id":    memberRoleID,
		}, http.StatusAccepted)
		return err
	})

	var requestID uuid.UUID
	must("GET /clubs/{id}/access-requests", func() error {
		data, err := c.expectOK("list access requests", http.MethodGet, fmt.Sprintf("%s/clubs/%s/access-requests", apiPrefix, clubID), headToken, nil, http.StatusOK)
		if err != nil {
			return err
		}
		var requests []struct {
			ID    uuid.UUID `json:"id"`
			Email string    `json:"email"`
		}
		if err := decode(data, &requests); err != nil {
			return err
		}
		for _, req := range requests {
			if req.Email == accessEmail {
				requestID = req.ID
				break
			}
		}
		if requestID == uuid.Nil {
			return fmt.Errorf("access request not found for %s", accessEmail)
		}
		return nil
	})
	must("POST /clubs/{id}/access-requests/{rid}/approve", func() error {
		_, err := c.expectOK("approve access request", http.MethodPost,
			fmt.Sprintf("%s/clubs/%s/access-requests/%s/approve", apiPrefix, clubID, requestID),
			headToken,
			map[string]any{"role_id": memberRoleID},
			http.StatusOK)
		return err
	})
	must("GET /admin/access-requests", func() error {
		_, err := c.expectOK("admin access requests", http.MethodGet, apiPrefix+"/admin/access-requests", adminToken, nil, http.StatusOK)
		return err
	})

	clubRegEmail := fmt.Sprintf("club-reg-%s@smoke.test", runID)
	clubRegName := "Smoke New Club " + runID
	var clubRegID uuid.UUID
	must("POST /club-registration-requests", func() error {
		data, err := c.expectOK("submit club registration", http.MethodPost, apiPrefix+"/club-registration-requests", "", map[string]any{
			"club_name":          clubRegName,
			"contact_email":      clubRegEmail,
			"contact_first_name": "Club",
			"contact_last_name":  "President",
			"country":            "Côte d'Ivoire",
			"city":               "Abidjan",
			"commune":            "Plateau",
			"description":        "Smoke club registration",
			"message":            "Please approve this smoke request",
		}, http.StatusAccepted)
		if err != nil {
			return err
		}
		var req struct {
			ID     uuid.UUID `json:"id"`
			Status string    `json:"status"`
		}
		if err := decode(data, &req); err != nil {
			return err
		}
		if req.Status != "pending" {
			return fmt.Errorf("expected pending status, got %s", req.Status)
		}
		clubRegID = req.ID
		return nil
	})
	must("GET /admin/club-registration-requests", func() error {
		data, err := c.expectOK("list club registrations", http.MethodGet,
			apiPrefix+"/admin/club-registration-requests", adminToken, nil, http.StatusOK)
		if err != nil {
			return err
		}
		var requests []struct {
			ID           uuid.UUID `json:"id"`
			ContactEmail string    `json:"contact_email"`
			Status       string    `json:"status"`
		}
		if err := decode(data, &requests); err != nil {
			return err
		}
		found := false
		for _, req := range requests {
			if req.ID == clubRegID && req.ContactEmail == clubRegEmail {
				found = true
				if req.Status != "pending" {
					return fmt.Errorf("expected pending in list, got %s", req.Status)
				}
				break
			}
		}
		if !found {
			return fmt.Errorf("club registration %s not found in admin list", clubRegID)
		}
		return nil
	})
	must("GET /admin/club-registration-requests/{id}", func() error {
		data, err := c.expectOK("get club registration", http.MethodGet,
			fmt.Sprintf("%s/admin/club-registration-requests/%s", apiPrefix, clubRegID),
			adminToken, nil, http.StatusOK)
		if err != nil {
			return err
		}
		var req struct {
			ID       uuid.UUID `json:"id"`
			ClubName string    `json:"club_name"`
		}
		if err := decode(data, &req); err != nil {
			return err
		}
		if req.ID != clubRegID || req.ClubName != clubRegName {
			return fmt.Errorf("unexpected club registration payload")
		}
		return nil
	})
	must("POST /admin/club-registration-requests/{id}/approve", func() error {
		data, err := c.expectOK("approve club registration", http.MethodPost,
			fmt.Sprintf("%s/admin/club-registration-requests/%s/approve", apiPrefix, clubRegID),
			adminToken,
			map[string]string{"review_note": "smoke approved"},
			http.StatusOK)
		if err != nil {
			return err
		}
		var res struct {
			Message string `json:"message"`
			Request struct {
				ID     uuid.UUID `json:"id"`
				Status string    `json:"status"`
			} `json:"request"`
			ExpiresAt time.Time `json:"expires_at"`
		}
		if err := decode(data, &res); err != nil {
			return err
		}
		if res.Request.Status != "approved" {
			return fmt.Errorf("expected approved status, got %s", res.Request.Status)
		}
		if res.ExpiresAt.IsZero() {
			return fmt.Errorf("expected access token expiry")
		}
		if strings.TrimSpace(res.Message) == "" {
			return fmt.Errorf("expected approval message")
		}
		return nil
	})

	clubRegRejectEmail := fmt.Sprintf("club-reg-reject-%s@smoke.test", runID)
	var clubRegRejectID uuid.UUID
	must("POST /club-registration-requests (reject path)", func() error {
		data, err := c.expectOK("submit club registration reject path", http.MethodPost, apiPrefix+"/club-registration-requests", "", map[string]any{
			"club_name":          "Smoke Reject Club " + runID,
			"contact_email":      clubRegRejectEmail,
			"contact_first_name": "Reject",
			"contact_last_name":  "Me",
			"country":            "Côte d'Ivoire",
			"city":               "Abidjan",
			"commune":            "Marcory",
		}, http.StatusAccepted)
		if err != nil {
			return err
		}
		var req struct {
			ID uuid.UUID `json:"id"`
		}
		if err := decode(data, &req); err != nil {
			return err
		}
		clubRegRejectID = req.ID
		return nil
	})
	must("POST /admin/club-registration-requests/{id}/reject", func() error {
		data, err := c.expectOK("reject club registration", http.MethodPost,
			fmt.Sprintf("%s/admin/club-registration-requests/%s/reject", apiPrefix, clubRegRejectID),
			adminToken,
			map[string]string{"note": "smoke rejected"},
			http.StatusOK)
		if err != nil {
			return err
		}
		var res struct {
			Status string `json:"status"`
		}
		if err := decode(data, &res); err != nil {
			return err
		}
		if res.Status != "rejected" {
			return fmt.Errorf("expected rejected, got %s", res.Status)
		}
		return nil
	})
	must("GET /admin/club-registration-requests?status=rejected", func() error {
		data, err := c.expectOK("list rejected club registrations", http.MethodGet,
			apiPrefix+"/admin/club-registration-requests?status=rejected", adminToken, nil, http.StatusOK)
		if err != nil {
			return err
		}
		var requests []struct {
			ID uuid.UUID `json:"id"`
		}
		if err := decode(data, &requests); err != nil {
			return err
		}
		for _, req := range requests {
			if req.ID == clubRegRejectID {
				return nil
			}
		}
		return fmt.Errorf("rejected club registration %s not found", clubRegRejectID)
	})

	pushEndpoint := fmt.Sprintf("https://push.smoke.test/%s", runID)
	must("POST /users/me/push-subscription", func() error {
		_, err := c.expectOK("push subscribe", http.MethodPost, apiPrefix+"/users/me/push-subscription", memberToken, map[string]string{
			"endpoint": pushEndpoint,
			"p256dh":   "BFakeP256dhKeyForSmokeTestOnly000000000000000000000000",
			"auth":     "FakeAuthKey00",
		}, http.StatusOK)
		return err
	})
	must("DELETE /users/me/push-subscription", func() error {
		_, err := c.expectOK("push unsubscribe", http.MethodDelete, apiPrefix+"/users/me/push-subscription", memberToken, map[string]string{
			"endpoint": pushEndpoint,
		}, http.StatusOK)
		return err
	})
	must("GET /widgets/birthday", func() error {
		_, err := c.expectOK("birthday widget", http.MethodGet, apiPrefix+"/widgets/birthday", memberToken, nil, http.StatusOK)
		return err
	})

	if cronSecret != "" {
		must("POST /internal/birthdays/run", func() error {
			req, err := http.NewRequest(http.MethodPost, c.base+apiPrefix+"/internal/birthdays/run", nil)
			if err != nil {
				return err
			}
			req.Header.Set("X-Cron-Secret", cronSecret)
			resp, err := c.http.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("status %d body=%s", resp.StatusCode, truncate(body, 200))
			}
			return nil
		})
	} else {
		fmt.Println("  skip POST /internal/birthdays/run (CRON_SECRET not set)")
	}

	must("DELETE /clubs/{id}/commissions/{cid}/members/{uid}", func() error {
		_, err := c.expectOK("remove commission member", http.MethodDelete,
			fmt.Sprintf("%s/clubs/%s/commissions/%s/members/%s", apiPrefix, clubID, commissionID, memberID),
			headToken, nil, http.StatusOK)
		return err
	})
	must("DELETE /clubs/{id}/commissions/{cid}", func() error {
		_, err := c.expectOK("delete commission", http.MethodDelete,
			fmt.Sprintf("%s/clubs/%s/commissions/%s", apiPrefix, clubID, commissionID),
			headToken, nil, http.StatusOK)
		return err
	})
	must("DELETE /chat/groups/{id}/messages/{mid}", func() error {
		_, err := c.expectOK("delete message", http.MethodDelete,
			fmt.Sprintf("%s/chat/groups/%s/messages/%s", apiPrefix, generalGroupID, messageID),
			memberToken, nil, http.StatusOK)
		return err
	})

	_ = commissionGroupID
	_ = headID

	fmt.Printf("\nSmoke test passed (%d steps)\n", steps)
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
