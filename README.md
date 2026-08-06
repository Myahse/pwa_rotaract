# Rotaract CIV

Digital platform for **Rotaract Côte d'Ivoire**.

> Présentation non technique (FR) : [`PRESENTATION.md`](./PRESENTATION.md)  
> Guide agents / build plan : [`AGENTS.md`](./AGENTS.md)

## Why this project exists

Rotaract is a youth program of **Rotary International** (Rotary Foundation): clubs of young adults focused on service, leadership, and fellowship.

In Côte d'Ivoire, day-to-day club life and management still happen mostly on **WhatsApp** and similar chat tools. That works for quick messages, but it does not scale for:

- structured club administration (members, roles, mandates, invites…)
- national / platform oversight
- lasting discussion, posts, and information sharing across clubs

**Goal:** replace ad‑hoc WhatsApp management with a dedicated Rotaract platform — while keeping a separate social space for community life (closer to a Facebook for Rotaract than to an admin console).

Clubs are organised in a **hierarchy** (platform → clubs → members / heads / commissions / mandates, diary of club history, etc.) as already modeled in the backend.

---

## Three platforms we are building

| # | Platform | Purpose | Repo today |
|---|----------|---------|------------|
| **1** | **Admin** | National / platform operators: create and oversee clubs, review access & club-registration requests, assign heads | `admin/` |
| **2** | **Club management** | Inside a club: members, invites, commissions, diary/mandates, profile, operational chat — the *management* app (not the social feed) | `frontend/` |
| **3** | **Social** | Cross-club community: posts, discussions, sharing news and updates — *Facebook-like for Rotaract*, distinct from management | `social/` |

These are **three products**, not one mega-app:

- Management stays focused on club operations.
- Social stays focused on interaction and content.
- Admin stays focused on platform governance.

A thin public **website** shell also exists under `website/` (FR/EN vitrine). It is **not** the social platform; we can grow it as a public site or fold pieces into social later.

---

## Current monorepo map

| App | Folder | Dev URL | Status |
|-----|--------|---------|--------|
| Club management (PWA) | `frontend/` | http://localhost:5173 | In progress — login → club facilities |
| Admin console | `admin/` | http://localhost:5174 | In progress — dashboard, clubs, requests |
| Public website (vitrine) | `website/` | http://localhost:5175 | Early shell — FR/EN content |
| Social platform | `social/` | http://localhost:5176 | MVP — feed, posts, comments, likes |
| API | `backend/` | http://localhost:8088 | Go REST + WebSocket + PostgreSQL |

```text
website (5175)     public vitrine
social  (5176)     community feed, posts, discussions
admin   (5174)  ──┐
frontend(5173)  ──┼──► backend API (8088) ──► PostgreSQL
                  ┘
```

- Separate JWT sessions per app (different `localStorage` keys).
- UUIDs everywhere.
- French UI for management apps; website is bilingual FR/EN.

---

## Quick start

### 1. Backend

```bash
cd backend
cp .env.example .env
# Edit DATABASE_URL and BOOTSTRAP_ADMIN_*
go run ./cmd/server
```

```
CORS_ORIGINS=http://localhost:5173,http://localhost:5174,http://localhost:5175,http://localhost:5176
```

### 2. Club management (`frontend/`)

```bash
cd frontend
npm install
npm run dev
```

Login → home, club tools, chat, profile. Invite register & access-request as secondary flows.

### 3. Admin (`admin/`)

```bash
cd admin
npm install
npm run dev
```

### 4. Public website (`website/`)

```bash
cd website
npm install
npm run dev
```

### 5. Social (`social/`)

```bash
cd social
npm install
npm run dev
```

Login with a club member account → feed, publish, comment, like.

---

## What “done” looks like (north star)

1. **Admin** — run the national platform without spreadsheets / WhatsApp threads.
2. **Club management** — every club runs members, roles, history, and ops in one place.
3. **Social** — Rotaractors across Côte d'Ivoire discuss, post, and relay information in a dedicated community space — separate from management.

---

## Next steps

- Harden **social** (media uploads, moderation, notifications)
- Keep hardening **admin** + **club management** against real club workflows
- Web Push on the member PWA
- Production deploy
