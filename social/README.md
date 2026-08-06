# Rotaract Social

Facebook-like community feed for Rotaract Côte d'Ivoire.

## Features
- **Guest browse** — open the feed without an account
- **Login gate** — like, comment, publish, follow require login (modal)
- **Photo & video posts** — up to 6 media files per post
- **Subscriptions** — follow Rotaractiens, filter “Abonnements”
- Distinct from club management (`../frontend/`) and admin (`../admin/`)

```bash
npm install
npm run dev
```

Dev: http://localhost:5176  
API: restart `backend` so migrations `014` + `015` apply.

Storage key: `rotaract_social_token`
