# Checkpoint — Rotaract CIV

**Date:** 2026-07-15  
**Status:** Working monorepo baseline (member PWA + admin + Go API)

Use this file as a restore / reference point for what exists and what still needs work.

---

## Stack

| Layer | Tech | Dev URL |
|-------|------|---------|
| Member PWA | React 19 + Vite 6 + `vite-plugin-pwa` | http://localhost:5173 |
| Admin | React 19 + Vite 6 | http://localhost:5174 |
| API | Go + PostgreSQL + WebSocket | http://localhost:8088 |

Separate JWT sessions per app (different `localStorage` keys). UUIDs everywhere.

---

## What’s in place

### Backend (`backend/`)

- Auth (login, JWT), bootstrap admin from env
- Clubs, members, roles, access requests
- Club registration requests (public submit → admin approve/reject; Google finish coded but commented out for now)
- Email invites (SMTP or console log in dev)
- Registration via invite token / invite code / access request
- Profile + avatar upload
- Chat groups + messages + WebSocket hub
- Birthday scheduler + Web Push (VAPID) + widget endpoint
- Club diary (tree: parrain / président / membre) — migrations `008+`
- Club mandates — migration `009`
- Club location + commissions — migration `010`
- Club executive roles — migration `011`
- Smoke test: `go run ./cmd/smoke`
- VAPID key helper: `go run ./cmd/vapid`

Migrations auto-run on server start (`001` → `013`).

### Member PWA (`frontend/`)

Routes:

- Public: `/login`, `/register`, `/access-request`
- Protected: `/` (home), `/profile`, `/clubs/:clubId`, `/clubs/:clubId/manage` (head), `/chat/:groupId`

Features:

- Auth + protected layout
- Club page + diary panel / tree view
- Head manage: email invites, commissions, access requests, diary
- Chat page
- Profile page

### Admin (`admin/`)

Routes:

- `/login`
- Clubs list + club detail (incl. diary panel)
- Access requests review
- Club registration requests review (backend ready; admin UI next)
- Member PWA `/register-club` Google finish page (route commented out)

---

## How to run (from this checkpoint)

```bash
# API
cd backend
cp .env.example .env   # if needed
go run ./cmd/server

# Member app
cd frontend && npm install && npm run dev

# Admin
cd admin && npm install && npm run dev
```

Set `CORS_ORIGINS=http://localhost:5173,http://localhost:5174` in `backend/.env`.

---

## Still open / next

- [ ] Web Push registration in the PWA (service worker + subscribe)
- [ ] Birthday celebration UI when widget says `is_my_birthday`
- [ ] Admin UI for club registration requests
- [ ] Public club registration form in member PWA
- [ ] Enable Google Sign-In (uncomment routes + `GOOGLE_CLIENT_ID`) + SMTP
- [x] Neon `DATABASE_URL` wired (run API from `backend/` so `.env` loads)
- [ ] Resend / revoke email invites
- [ ] Email template branding (club logo)
- [ ] Production deploy (static builds + API)
- [ ] Polish head tools / UX as needed

---

## Key paths

```text
README.md                 # monorepo overview
CHECKPOINT.md             # this file
backend/README.md         # API docs & env
backend/internal/         # handlers, services, repos
backend/internal/database/migrations/
frontend/src/App.tsx
frontend/src/pages/HeadManagePage.tsx
frontend/src/components/ClubDiaryPanel.tsx
admin/src/pages/ClubsPage.tsx
admin/src/pages/ClubDetailPage.tsx
```

---

## Note

Do not commit secrets (`.env`, VAPID private keys, SMTP passwords). Prefer `.env.example` only in git.
