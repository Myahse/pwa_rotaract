# AGENTS.md — Rotaract CIV

Instructions for AI agents (and humans) building this monorepo **from beginning to end**.

Read this file first on every non-trivial task. Product vision (non-technical, FR): [`PRESENTATION.md`](./PRESENTATION.md). Short map: [`README.md`](./README.md).

---

## 1. Mission

Replace WhatsApp-as-management for Rotaract Côte d’Ivoire with **three separate products**:

| Product | Folder | Job |
|---------|--------|-----|
| **Admin** | `admin/` | National governance of clubs & requests |
| **Club management** | `frontend/` | Run a club day-to-day (ops, members, memory) |
| **Social** | `social/` *(create)* | Facebook-like community: posts, discussions, relay info |
| Public vitrine | `website/` | Marketing site FR/EN — **not** social |
| API | `backend/` | Single Go API + PostgreSQL for all authenticated apps |

Never merge social feed UX into admin or club-management shells. Never turn management into a social network.

---

## 2. Master agent prompt (copy into a new chat)

Use this prompt to start a focused build session:

```text
You are the lead engineer for Rotaract CIV (monorepo).

PRODUCT
- Rotaract is Rotary’s youth program. In Côte d’Ivoire, clubs currently manage almost everything on WhatsApp.
- We build THREE platforms + one public site:
  1) admin/ — national admin
  2) frontend/ — club management PWA (login → club facilities)
  3) social/ — community network (posts, discussions) — DISTINCT from management
  4) website/ — public FR/EN vitrine only
- Shared backend/: Go API + PostgreSQL.

RULES
- Follow AGENTS.md, PRESENTATION.md (product), README.md (map).
- Prefer extending the existing stack over rewrites.
- Mobile-first UX; French UI for admin + management; website bilingual.
- Small, focused diffs. No drive-by refactors. No new markdown docs unless asked.
- Separate JWT localStorage keys per app. UUIDs everywhere.
- Admin APIs under /api/v1/admin/* with RequireAdmin.
- Club ops under /api/v1/clubs/{id}/… with membership/permissions.
- Social domain gets its own tables + /api/v1/social/* (or /feed/*) — do not overload club chat as the social product.

STACK (do not change without explicit approval)
- Backend: Go 1.22+, chi, pgx, goose migrations, JWT, bcrypt, optional SMTP + Web Push
- Frontends: React 19 + TypeScript + Vite 6 + react-router-dom 7 + plain CSS variables (no Tailwind unless already introduced)
- Member app: vite-plugin-pwa
- DB: PostgreSQL

CURRENT TASK
- [paste one phase or checklist item from AGENTS.md §5]
- Deliver working code, run builds for touched apps, summarize what shipped and what’s next.
```

---

## 3. Most efficient stack (canonical — stick to it)

Chosen for **speed of delivery**, **one deployable API**, and **what already works in this repo**. Do not introduce a second backend language, Next.js, or a heavy UI kit unless the user explicitly asks.

### Backend
| Piece | Choice | Why |
|-------|--------|-----|
| Language | **Go** | Already in repo; fast, simple deploy |
| HTTP | **chi** | Lightweight, middleware-friendly |
| DB | **PostgreSQL + pgx** | Relational model fits clubs/roles/mandates |
| Migrations | **goose** SQL files | Explicit, reviewable |
| Auth | **JWT + bcrypt** | Separate tokens per SPA |
| Realtime | **WebSocket** (existing hub) | Club ops chat; social can reuse later for notifications |
| Files | Local upload dir (evolve to S3 later) | Enough for MVP |

### Frontends (4 SPAs)
| App | Port (dev) | Stack |
|-----|------------|--------|
| `frontend` | 5173 | React 19 + Vite + PWA + CSS vars |
| `admin` | 5174 | React 19 + Vite + CSS vars |
| `website` | 5175 | React 19 + Vite + i18n FR/EN |
| `social` | 5176 | React 19 + Vite + CSS vars (create when starting Phase D) |

**Shared efficiency rules**
- One API (`backend`) — no BFF per app unless forced by deploy constraints.
- Duplicate thin `api/client.ts` + `auth/storage.ts` per app (different storage keys) — avoid premature monorepo packages until 2+ apps share large types.
- Brand: magenta `#be034d`, fonts Fraunces + Sora where already used.
- Proxy `/api` → `http://localhost:8088` in each Vite app that talks to the API.

### Explicitly defer (efficient “no”)
- No GraphQL until REST hurts
- No microservices
- No Tailwind/MUI/Chakra unless requested
- No rewriting Go → Node/Python
- No merging the three products into one SPA with role tabs

---

## 4. Architecture boundaries

```text
website  → static/public content only
social   → /api/v1/social/*  (feed, posts, comments, reactions…)
frontend → /api/v1/auth, /clubs/*, chat, profile, invites…
admin    → /api/v1/admin/*, plus read club tools as admin
         ↘
      backend (chi) → PostgreSQL
```

**Roles**
- `users.is_admin` → platform admin
- `club_memberships.member_role` = `head` | `member`
- Club permission keys for finer control; admins bypass club checks

**Data domains (keep separated)**
1. Identity & auth  
2. Clubs, memberships, roles, commissions, mandates, diary  
3. Access / club-registration requests  
4. Ops chat (club groups)  
5. Social feed (new) — posts belong to users/clubs, not “chat messages”

---

## 5. End-to-end task list

Work **in order**. Check items off in PRs/commits. Do not start Social UI before Admin + Club management are usable by a pilot club.

### Phase 0 — Foundations (done / maintain)
- [x] Monorepo layout: `backend`, `frontend`, `admin`, `website`
- [x] PostgreSQL migrations, JWT auth, bootstrap admin
- [x] CORS for 5173/5174; document 5175/5176 when needed
- [x] `PRESENTATION.md` + `README.md` + this `AGENTS.md`
- [ ] Smoke test green on a clean DB (`go run ./cmd/smoke`)
- [ ] `.env.example` accurate for all required vars

### Phase A — Admin platform (national)
- [x] Login (admin-only)
- [x] Dashboard KPIs
- [x] Clubs list/create/detail, assign head, diary panel
- [x] Member access-request review UI
- [x] Club-registration request review UI
- [ ] Club search/filters polish + empty states
- [ ] Soft-delete/deactivate flows confirmed with real data
- [ ] Admin can open a club “ops snapshot” (member count, head, pending requests)
- [ ] Audit-friendly success/error toasts (no silent failures)
- [ ] Mobile shell QA (bottom nav / sidebar)

### Phase B — Club management (`frontend`)
- [x] Login-first entry (no marketing landing)
- [x] Home, profile, club page, head manage, ops chat
- [x] Invites, commissions, access requests (head tools)
- [x] Diary / mandates panels
- [ ] Head UX: invite resend/revoke
- [ ] Member directory clarity (roles visible)
- [ ] Birthday widget polish when API says it’s the user’s birthday
- [ ] Web Push registration in PWA
- [ ] Offline-friendly shell (PWA) for login + last club context
- [ ] Pilot checklist: new member join → appears in club → head can manage

### Phase C — Public website (`website`)
- [x] White FR/EN shell + language switcher
- [ ] Real contact email / social links (when provided)
- [ ] Pages to add over time: clubs list (public), projects, news, join CTA pointing to management login/access-request
- [ ] Do **not** build the social feed here

### Phase D — Social platform (`social/`) — build next major product
#### D0 — Scaffold
- [x] Create `social/` Vite React TS app on port **5176**
- [x] Auth reuse: login against `/api/v1/auth/login`, storage key `rotaract_social_token`
- [x] App shell: feed, compose, profile
- [x] Add CORS origin + README row

#### D1 — API domain
- [x] Migration `014`: `social_posts`, `social_comments`, `social_reactions`
- [x] Migration `015`: `social_post_media`, `social_follows` + guest-readable feed
- [x] Public feed/comments/suggestions (optional auth); writes require login
- [x] Photo/video upload on posts (multipart)
- [x] Follow / unfollow / suggestions / following list
- [x] Pagination via `before` + `limit` on feed
- [x] Migration `016`: comment replies, friendships, DMs, comment share to friends
- [x] User profiles + friend requests + private messages (friends-only)
- [x] Migration `018`: social groups, join requests, group chat
- [ ] Moderation hooks for admin (hide post)

#### D2 — Social UX
- [x] Facebook-like shell (top bar, 3-col desktop, mobile bottom nav)
- [x] Guest banner + login modal for gated actions
- [x] Composer with photo/video, feed tabs Pour vous / Abonnements
- [x] Comments, likes, follow, share link
- [x] Nested comment replies + share comment to a friend (DM)
- [x] Friends / Messages / Profile sections in `social/`
- [x] Groups: create, join/request, admin approve, group chat
- [ ] Report/hide stub

#### D3 — Hardening
- [ ] Rate limits / abuse basics
- [ ] Admin moderation endpoints (hide post, ban user from social)
- [ ] Notifications (in-app first; push later)
- [ ] Realtime DMs via WebSocket (optional upgrade)
- [ ] Load test mental model: feed query indexes

### Phase E — Cross-cutting production
- [ ] Single deploy story documented (API + 4 static fronts)
- [ ] Backups for PostgreSQL
- [ ] Object storage for uploads if disk is not enough
- [ ] Monitoring/logging basics
- [ ] Security review (authz on every club & social route)
- [ ] Pilot with 1–2 clubs, then national rollout

---

## 6. Code review checklist (agent must self-review)

Before marking a task done, verify:

### Correctness
- [ ] Matches the product boundary (admin vs management vs social vs website)
- [ ] AuthZ checked server-side (never trust the client role flag alone)
- [ ] Migrations have `up` + `down` when schema changes
- [ ] Error messages safe for users (FR where UI is FR); no stack traces in API JSON

### API
- [ ] Consistent `/api/v1/...` paths and JSON field names (`snake_case`)
- [ ] Admin routes behind `RequireAdmin`
- [ ] Club routes verify membership/permission (admin bypass OK)
- [ ] No N+1 obvious queries on list endpoints

### Frontend
- [ ] Loading / empty / error states present
- [ ] Mobile-first; no desktop-only critical flows
- [ ] Tokens in the correct `localStorage` key for that app
- [ ] No secrets in client bundles

### Quality bar
- [ ] `npm run build` (touched frontends) and/or `go test` / smoke for backend changes
- [ ] Diff is scoped — no unrelated formatting storms
- [ ] No new dependency without a one-line reason

### Security red flags (reject / fix)
- IDOR on `clubId` / `postId`
- Missing auth on write endpoints
- Password or JWT logged
- Public listing of private member PII

---

## 7. How agents should work day-to-day

1. **Pick one checklist item** from §5 — state it in the reply.  
2. **Read** existing handlers/pages before inventing new patterns.  
3. **Implement** the smallest vertical slice (API + UI if needed).  
4. **Self-review** with §6.  
5. **Build/test** what you touched.  
6. **Update** checklist status in the PR description (or ask user before editing this file’s checkboxes in bulk).  
7. Stop when the slice is demoable — don’t boil the ocean.

### Preferred file touch patterns
- Backend feature: `domain` → `migration` → `repository` → `service` → `handler` → `server.go` route  
- Admin UI: `admin/src/pages` + `api/types.ts` + nav in `AdminLayout`  
- Club UI: `frontend/src/pages` + head tools in `HeadManagePage`  
- Social UI: only under `social/` once scaffolded  

### Language
- Product/UI copy for admin + management: **French**  
- Website: **FR + EN** via i18n  
- Code comments: English OK; avoid noisy comments  

---

## 8. Definition of done (per platform)

### Admin — Done enough for national ops
An admin can create a club + president, fetch invite code, review member access requests and club-registration requests, and see dashboard counts — on phone and desktop.

### Club management — Done enough for a pilot club
A head can invite members, see roster/roles, manage commissions, keep diary/mandates, and members can log in, see club, profile, and ops chat.

### Social — Done enough for a pilot community
A member can open a feed, publish a post, comment, react, and see posts from other clubs — without using club-management screens.

### Website — Done enough for public presence
Visitor can read who we are in FR/EN and find a path to request access / contact.

---

## 9. Prompt snippets for common jobs

**Add an admin page**
```text
Add admin UI + wire existing or new /api/v1/admin/… endpoints. Follow admin design system in admin/src/index.css. Update AdminLayout nav. French copy. Mobile-first.
```

**Add club head capability**
```text
Extend HeadManagePage + backend club permission-checked routes. Do not put this in social/. French UI.
```

**Start social MVP**
```text
Execute Phase D0–D2 from AGENTS.md. Scaffold social/ on 5176, migrations for posts/comments/reactions, feed UI. Reuse auth patterns from frontend. Keep product boundary strict.
```

**Code review only**
```text
Review the current diff against AGENTS.md §6. List blockers, risks, and nits. Do not rewrite unrelated code.
```

---

## 10. Reference links inside the repo

| Doc | Use |
|-----|-----|
| `PRESENTATION.md` | Stakeholder story & functional depth (FR) |
| `README.md` | Dev ports & quick start |
| `CHECKPOINT.md` | Historical snapshot of what existed |
| `backend/README.md` | API surface |
| `AGENTS.md` | **This file** — build order, stack, review |

---

*Agents: optimize for a working pilot of all three platforms, not for perfect abstraction.*
