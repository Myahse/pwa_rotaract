# Rotaract CIV — Backend API

Go HTTP API for the Rotaract CIV PWA.

## Club registration (request → admin → Google finish)

```text
1. Club submits public request (contact email + club details)
2. Admin reviews and approves
3. Contact receives email with /register-club?token=...
4. Contact finishes with Google Sign-In (email must match contact_email)
5. Club + president account are created; JWT session is returned
```

| Method | Path | Auth |
|--------|------|------|
| POST | `/api/v1/club-registration-requests` | Public |
| GET | `/api/v1/admin/club-registration-requests` | Admin (`?status=` optional, default `pending`) |
| GET | `/api/v1/admin/club-registration-requests/{id}` | Admin |
| POST | `/api/v1/admin/club-registration-requests/{id}/approve` | Admin |
| POST | `/api/v1/admin/club-registration-requests/{id}/reject` | Admin |
| GET | `/api/v1/club-registration/access/{token}` | Public |
| POST | `/api/v1/club-registration/complete` | Public (`token` + `google_id_token`) |
| GET | `/api/v1/auth/google-client-id` | Public |

Google finish is implemented but **commented out for now** (routes + `GOOGLE_CLIENT_ID`).  
When enabling: uncomment the Google routes in `server.go`, `/register-club` in the PWA, and set `GOOGLE_CLIENT_ID`.  
Authorized JavaScript origins must include the PWA origin (e.g. `http://localhost:5173`).  
Without `BREVO_API_KEY`, transactional emails are logged to the API console instead of being sent.

## Public events (website)

Published events drive the public site (`website/` `/events` and `/events/:id`). Flyers and gallery images are uploaded from Admin.

| Method | Path | Auth |
|--------|------|------|
| GET | `/api/v1/events` | Public (published only) |
| GET | `/api/v1/events/{eventID}` | Public (published only) |
| GET | `/api/v1/admin/events` | Admin |
| POST | `/api/v1/admin/events` | Admin |
| GET | `/api/v1/admin/events/{eventID}` | Admin |
| PATCH | `/api/v1/admin/events/{eventID}` | Admin |
| DELETE | `/api/v1/admin/events/{eventID}` | Admin |
| POST | `/api/v1/admin/events/{eventID}/flyer` | Admin (`multipart`, field `file`) |
| POST | `/api/v1/admin/events/{eventID}/images` | Admin (`multipart`, field `file`) |
| DELETE | `/api/v1/admin/events/{eventID}/images/{imageID}` | Admin |

## Member onboarding (email invite — primary flow)

```text
1. Admin creates club + assigns head (or approves a club registration request)
2. Head enters member email (+ optional role)
3. Member receives email with personal link: /register?token=...
4. Member completes profile → joined instantly under their club
```

Fallback without email invite: `POST /access-requests` (club name + profile, head/admin approves).

## Setup

### 1. Local PostgreSQL

Install [PostgreSQL 16+](https://www.postgresql.org/download/windows/).

Set the **full connection string** in `.env` (the app only reads `DATABASE_URL`, not separate username/password lines):

```
DATABASE_URL=postgresql://postgres:YOUR_PASSWORD@localhost:5432/postgres?sslmode=disable
```

Replace `YOUR_PASSWORD` and the database name if you use a dedicated database (e.g. `rotaract`).

### 2. Configure and run the API

```bash
cp .env.example .env
go mod tidy
go run ./cmd/server
```

Migrations run automatically on startup. The first admin is created from `BOOTSTRAP_ADMIN_*` when no admin exists yet.

Configure Brevo for production email delivery (`BREVO_API_KEY`, `BREVO_SENDER_EMAIL`). In dev, if `BREVO_API_KEY` is empty, invite links are **logged to the console** instead.

### Database: local vs Neon

| Environment | `DATABASE_URL` |
|-------------|----------------|
| **Local (default)** | `postgresql://postgres:pass@localhost:5432/postgres?sslmode=disable` |
| **Neon (later)** | `postgresql://user:pass@ep-....neon.tech/neondb?sslmode=require` |

Only change `DATABASE_URL` in `.env` when switching — no code changes needed.

## Profile (authenticated user)

## Import Tombola members

The Tombola database is separate. Set `TOMBOLA_DATABASE_URL` temporarily, then run:

```bash
go run ./cmd/import-tombola -dry-run
go run ./cmd/import-tombola
```

The import matches existing clubs by name and is safe to run again. Existing emails are not overwritten. Tombola passwords use a different hash format, so imported members must use the password-reset email before logging in.

## Remove smoke / mock users

After running `cmd/smoke` against a shared database, remove `@smoke.test` users while keeping admins and Tombola imports:

```bash
go run ./cmd/cleanup-mock -dry-run
go run ./cmd/cleanup-mock
```

Optional: set `TOMBOLA_DATABASE_URL` so imported member emails are always kept.

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/auth/me` | Current profile |
| PATCH | `/api/v1/users/me` | Update personal info |
| POST | `/api/v1/users/me/avatar` | Upload profile picture (`multipart/form-data`, field `avatar`) |
| DELETE | `/api/v1/users/me/avatar` | Remove profile picture |

Updatable fields: `first_name`, `last_name`, `phone`, `birth_date`, `profession`, `member_since`.  
Email cannot be changed yet (will need verification flow).

Avatar formats: JPEG, PNG, WebP — max 5 MB.

## Birthday notifications & widget

Uses each member's `birth_date` to:

1. **Push notification** at 8:00 AM (Africa/Abidjan by default) — personalized message
2. **Widget data** for the PWA home screen / lock screen experience

### Setup Web Push

```bash
go run ./cmd/vapid
```

Copy the keys into `.env`, then members register their device:

```http
POST /api/v1/users/me/push-subscription
Authorization: Bearer <token>

{
  "endpoint": "...",
  "p256dh": "...",
  "auth": "..."
}
```

### Widget endpoint (PWA polls this)

```http
GET /api/v1/widgets/birthday
Authorization: Bearer <token>
```

Example when it's your birthday:

```json
{
  "is_my_birthday": true,
  "title": "Joyeux anniversaire Aya !",
  "message": "Aya, toute la communauté Rotaract CIV te souhaite...",
  "emoji": "🎂",
  "theme": { "primary": "#2563eb", "accent": "#f59e0b" }
}
```

### Club birthdays (bureau / member directory)

```http
GET /api/v1/clubs/{clubID}/birthdays/today
```

### Manual trigger (cron / testing)

```http
POST /api/v1/internal/birthdays/run
X-Cron-Secret: <CRON_SECRET>
```

### PWA frontend (next step)

The React PWA will need:

- Service worker with **Web Push** + `push` event handler
- **Periodic Background Sync** or daily poll of `/widgets/birthday`
- Birthday celebration screen on open when `is_my_birthday: true`
- Note: true native lock screen widgets (like Duolingo's app) require a native wrapper; PWAs use push notifications + home screen widget patterns (Android WebAPK, iOS limited)

## Email invite endpoints

### Head sends invite

```http
POST /api/v1/clubs/{clubID}/email-invites
Authorization: Bearer <head-token>
Content-Type: application/json

{
  "email": "membre@example.com",
  "role_id": "optional-uuid-of-role"
}
```

Member receives an email with a link like:
`https://your-app/register?token=abc123...`

### Member opens link (public)

```http
GET /api/v1/invite/token/{token}
```

Returns club name, pre-filled email, available roles, expiry.

### Member completes registration (public)

```http
POST /api/v1/register
Content-Type: application/json

{
  "token": "from-email-link",
  "email": "membre@example.com",
  "password": "secret123",
  "first_name": "Aya",
  "last_name": "Koné",
  "birth_date": "1998-05-12",
  "profession": "Ingénieure",
  "member_since": "2024-09-01"
}
```

Email must match the invitation. If the head preset a role, `role_id` is optional.

### Head lists sent invites

```http
GET /api/v1/clubs/{clubID}/email-invites
```

## Environment variables

| Variable | Description |
|----------|-------------|
| `DATABASE_URL` | PostgreSQL connection string (local or Neon) |
| `APP_PUBLIC_URL` | Frontend URL used in email links |
| `INVITE_TTL` | Invite link validity (default `168h` = 7 days) |
| `BREVO_API_KEY` | Brevo API key (empty = dev mode, logs only) |
| `BREVO_SENDER_EMAIL` | Verified sender address in Brevo |
| `BREVO_SENDER_NAME` | Sender display name |

## Chat messages

### List messages

```http
GET /api/v1/chat/groups/{groupID}/messages?limit=50&before=2026-01-01T00:00:00Z
Authorization: Bearer <token>
```

### Send message

```http
POST /api/v1/chat/groups/{groupID}/messages
Authorization: Bearer <token>
Content-Type: application/json

{ "content": "Hello club!" }
```

New messages are broadcast to WebSocket subscribers for that group.

### WebSocket

```http
GET /api/v1/ws/chat?token=<jwt>&group_id=<uuid>
```

Pass the JWT as `token` query param (or `Authorization: Bearer`). Events look like:

```json
{ "type": "message", "group_id": "...", "message": { ... } }
```

## Smoke test

Run the API server first, then:

```bash
go run ./cmd/smoke
```

Environment (from `.env` or shell):

| Variable | Default |
|----------|---------|
| `SMOKE_BASE_URL` | `http://localhost:8088` |
| `BOOTSTRAP_ADMIN_EMAIL` | `admin@rotaract-civ.local` |
| `BOOTSTRAP_ADMIN_PASSWORD` | `change-me` |
| `CRON_SECRET` | optional — skips internal birthday job if unset |

The script exercises health, auth, clubs, registration, commissions, chat groups/messages, WebSocket, access requests, club registration requests (submit/list/approve/reject), push subscription, and birthdays.

## Other flows

- **Generic club code** (no email): `GET /invite/{code}` + `POST /register` with `invite_code`
- **Access request**: `POST /access-requests` when member has no invite
- **Admin/head approve**: `/access-requests/{id}/approve`

## Next steps

- [ ] Resend / revoke email invites
- [ ] React PWA frontend
- [ ] Email template branding (club logo)
