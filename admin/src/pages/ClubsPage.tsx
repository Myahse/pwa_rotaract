import { useEffect, useMemo, useState, type FormEvent } from 'react'
import { Link } from 'react-router-dom'
import { apiRequest, apiUpload } from '../api/client'
import type { Club, CreateClubResult } from '../api/types'
import { useAuth } from '../auth/AuthContext'
import { IconPlus } from '../components/Icons'

const emptyPresident = {
  first_name: '',
  last_name: '',
  email: '',
  password: '',
  phone: '',
  birth_date: '',
  profession: '',
  member_since: '',
}

const emptyClub = {
  name: '',
  slug: '',
  country: "Côte d'Ivoire",
  city: '',
  commune: '',
  description: '',
  founded_at: '',
}

export function ClubsPage() {
  const { token } = useAuth()
  const [clubs, setClubs] = useState<Club[]>([])
  const [query, setQuery] = useState('')
  const [drawerOpen, setDrawerOpen] = useState(false)
  const [clubForm, setClubForm] = useState(emptyClub)
  const [presidentForm, setPresidentForm] = useState(emptyPresident)
  const [logoFile, setLogoFile] = useState<File | null>(null)
  const [logoPreview, setLogoPreview] = useState<string | null>(null)
  const [avatarFile, setAvatarFile] = useState<File | null>(null)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')
  const [loading, setLoading] = useState(false)

  async function load() {
    if (!token) return
    const data = await apiRequest<Club[]>('/admin/clubs', {}, token)
    setClubs(data)
  }

  useEffect(() => {
    load().catch((err) => setError(err instanceof Error ? err.message : 'Erreur'))
  }, [token])

  useEffect(() => {
    if (!logoFile) {
      setLogoPreview(null)
      return
    }
    const url = URL.createObjectURL(logoFile)
    setLogoPreview(url)
    return () => URL.revokeObjectURL(url)
  }, [logoFile])

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase()
    if (!q) return clubs
    return clubs.filter((c) =>
      [c.name, c.city, c.commune, c.invite_code, c.slug]
        .filter(Boolean)
        .some((v) => v!.toLowerCase().includes(q)),
    )
  }, [clubs, query])

  function closeDrawer() {
    setDrawerOpen(false)
    setError('')
  }

  async function onCreate(e: FormEvent) {
    e.preventDefault()
    if (!token) return
    if (!logoFile) {
      setError('Le logo du club est obligatoire')
      return
    }
    setLoading(true)
    setError('')
    setSuccess('')
    try {
      const body = new FormData()
      body.append('name', clubForm.name)
      if (clubForm.slug) body.append('slug', clubForm.slug)
      body.append('country', clubForm.country)
      body.append('city', clubForm.city)
      body.append('commune', clubForm.commune)
      if (clubForm.description) body.append('description', clubForm.description)
      if (clubForm.founded_at) body.append('founded_at', clubForm.founded_at)
      body.append('logo', logoFile)
      body.append('president_first_name', presidentForm.first_name)
      body.append('president_last_name', presidentForm.last_name)
      body.append('president_email', presidentForm.email)
      body.append('president_password', presidentForm.password)
      if (presidentForm.phone) body.append('president_phone', presidentForm.phone)
      if (presidentForm.birth_date) body.append('president_birth_date', presidentForm.birth_date)
      if (presidentForm.profession) body.append('president_profession', presidentForm.profession)
      if (presidentForm.member_since) body.append('president_member_since', presidentForm.member_since)
      if (avatarFile) body.append('president_avatar', avatarFile)

      const result = await apiUpload<CreateClubResult>('/admin/clubs', body, token)
      setClubForm(emptyClub)
      setPresidentForm(emptyPresident)
      setLogoFile(null)
      setAvatarFile(null)
      setSuccess(`Club « ${result.club.name} » créé avec ${result.president.first_name} ${result.president.last_name}`)
      setDrawerOpen(false)
      await load()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Création impossible')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="stack gap-lg">
      <header className="page-header">
        <div className="row-between">
          <div>
            <h1>Clubs</h1>
            <p>{clubs.length} club{clubs.length !== 1 ? 's' : ''} sur la plateforme</p>
          </div>
          <button type="button" className="btn-primary" onClick={() => { setDrawerOpen(true); setSuccess(''); setError('') }}>
            <IconPlus style={{ width: 18, height: 18 }} /> Nouveau
          </button>
        </div>
      </header>

      {success && <p className="success" role="status">{success}</p>}
      {error && !drawerOpen && <p className="error" role="alert">{error}</p>}

      <div className="search-bar">
        <input
          type="search"
          placeholder="Rechercher un club, une ville, un code…"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          aria-label="Rechercher"
        />
      </div>

      {filtered.length === 0 ? (
        <div className="panel empty-state">
          <strong>{query ? 'Aucun résultat' : 'Aucun club'}</strong>
          <p>{query ? 'Essayez un autre terme.' : 'Créez le premier club pour démarrer.'}</p>
        </div>
      ) : (
        <>
          <div className="club-list">
            {filtered.map((c) => (
              <Link key={c.id} to={`/clubs/${c.id}`} className="club-item">
                {c.logo_url ? (
                  <img src={c.logo_url} alt="" className="club-logo-sm" />
                ) : (
                  <span className="brand-mark" style={{ width: 48, height: 48 }}>{c.name[0]}</span>
                )}
                <div className="club-item-body">
                  <strong>
                    {c.name}{' '}
                    {!c.is_active && <span className="badge inactive">inactif</span>}
                  </strong>
                  <div className="club-item-meta">
                    {[c.commune, c.city].filter(Boolean).join(', ') || '—'}
                    {' · '}
                    <code>{c.invite_code}</code>
                  </div>
                </div>
                <span className="club-item-arrow" aria-hidden>›</span>
              </Link>
            ))}
          </div>

          <section className="panel table-wrap">
            <table className="table">
              <thead>
                <tr>
                  <th>Club</th>
                  <th>Code invite</th>
                  <th>Localisation</th>
                  <th>Création</th>
                  <th></th>
                </tr>
              </thead>
              <tbody>
                {filtered.map((c) => (
                  <tr key={c.id}>
                    <td>
                      <div style={{ display: 'flex', alignItems: 'center', gap: '0.65rem' }}>
                        {c.logo_url && <img src={c.logo_url} alt="" className="club-logo-sm" />}
                        <span>
                          {c.name} {!c.is_active && <span className="badge inactive">inactif</span>}
                        </span>
                      </div>
                    </td>
                    <td><code>{c.invite_code}</code></td>
                    <td>{[c.commune, c.city, c.country].filter(Boolean).join(', ') || '—'}</td>
                    <td>{c.founded_at ? new Date(c.founded_at).toLocaleDateString('fr-FR') : '—'}</td>
                    <td><Link to={`/clubs/${c.id}`}>Gérer</Link></td>
                  </tr>
                ))}
              </tbody>
            </table>
          </section>
        </>
      )}

      {drawerOpen && (
        <>
          <div className="drawer-backdrop" onClick={closeDrawer} aria-hidden />
          <div className="drawer" role="dialog" aria-modal="true" aria-labelledby="create-club-title">
            <div className="drawer-handle" />
            <form onSubmit={onCreate} className="stack gap-lg">
              <div className="row-between">
                <h2 id="create-club-title">Nouveau club</h2>
                <button type="button" className="btn-ghost btn-small ghost" onClick={closeDrawer}>Fermer</button>
              </div>

              <section className="form-section stack">
                <h3>Informations du club</h3>
                <div className="grid-2">
                  <label>
                    Nom du club *
                    <input value={clubForm.name} onChange={(e) => setClubForm({ ...clubForm, name: e.target.value })} required />
                  </label>
                  <label>
                    Slug
                    <input value={clubForm.slug} onChange={(e) => setClubForm({ ...clubForm, slug: e.target.value })} placeholder="auto" />
                  </label>
                  <label>
                    Pays *
                    <input value={clubForm.country} onChange={(e) => setClubForm({ ...clubForm, country: e.target.value })} required />
                  </label>
                  <label>
                    Ville *
                    <input value={clubForm.city} onChange={(e) => setClubForm({ ...clubForm, city: e.target.value })} required />
                  </label>
                  <label>
                    Commune *
                    <input value={clubForm.commune} onChange={(e) => setClubForm({ ...clubForm, commune: e.target.value })} required />
                  </label>
                  <label>
                    Date de création
                    <input type="date" value={clubForm.founded_at} onChange={(e) => setClubForm({ ...clubForm, founded_at: e.target.value })} />
                  </label>
                  <label className="span-2">
                    Description
                    <input value={clubForm.description} onChange={(e) => setClubForm({ ...clubForm, description: e.target.value })} />
                  </label>
                  <label className="span-2">
                    Logo du club *
                    <input type="file" accept="image/jpeg,image/png,image/webp" onChange={(e) => setLogoFile(e.target.files?.[0] ?? null)} required />
                  </label>
                </div>
                {logoPreview && (
                  <div className="logo-preview">
                    <img src={logoPreview} alt="Aperçu logo" />
                  </div>
                )}
              </section>

              <section className="form-section stack">
                <h3>Président du club</h3>
                <p className="muted small">Compte responsable créé automatiquement avec le club.</p>
                <div className="grid-2">
                  <label>
                    Prénom *
                    <input value={presidentForm.first_name} onChange={(e) => setPresidentForm({ ...presidentForm, first_name: e.target.value })} required />
                  </label>
                  <label>
                    Nom *
                    <input value={presidentForm.last_name} onChange={(e) => setPresidentForm({ ...presidentForm, last_name: e.target.value })} required />
                  </label>
                  <label>
                    Email *
                    <input type="email" value={presidentForm.email} onChange={(e) => setPresidentForm({ ...presidentForm, email: e.target.value })} required />
                  </label>
                  <label>
                    Mot de passe *
                    <input type="password" value={presidentForm.password} onChange={(e) => setPresidentForm({ ...presidentForm, password: e.target.value })} required minLength={8} />
                  </label>
                  <label>
                    Téléphone
                    <input value={presidentForm.phone} onChange={(e) => setPresidentForm({ ...presidentForm, phone: e.target.value })} />
                  </label>
                  <label>
                    Date de naissance
                    <input type="date" value={presidentForm.birth_date} onChange={(e) => setPresidentForm({ ...presidentForm, birth_date: e.target.value })} />
                  </label>
                  <label>
                    Profession
                    <input value={presidentForm.profession} onChange={(e) => setPresidentForm({ ...presidentForm, profession: e.target.value })} />
                  </label>
                  <label>
                    Membre Rotaract depuis
                    <input type="date" value={presidentForm.member_since} onChange={(e) => setPresidentForm({ ...presidentForm, member_since: e.target.value })} />
                  </label>
                  <label className="span-2">
                    Photo du président
                    <input type="file" accept="image/jpeg,image/png,image/webp" onChange={(e) => setAvatarFile(e.target.files?.[0] ?? null)} />
                  </label>
                </div>
              </section>

              {error && <p className="error" role="alert">{error}</p>}
              <button type="submit" className="btn-primary btn-block" disabled={loading}>
                {loading ? 'Création…' : 'Créer le club et le président'}
              </button>
            </form>
          </div>
        </>
      )}
    </div>
  )
}
