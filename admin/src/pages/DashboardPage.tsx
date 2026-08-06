import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { apiRequest } from '../api/client'
import type { AccessRequest, Club, ClubRegistrationRequest } from '../api/types'
import { useAuth } from '../auth/AuthContext'
import { IconAccess, IconClubs, IconPlus, IconRegister } from '../components/Icons'

export function DashboardPage() {
  const { token, user } = useAuth()
  const [clubs, setClubs] = useState<Club[]>([])
  const [accessPending, setAccessPending] = useState(0)
  const [clubRegsPending, setClubRegsPending] = useState(0)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (!token) return
    let cancelled = false
    ;(async () => {
      setLoading(true)
      setError('')
      try {
        const [clubList, accessList, regList] = await Promise.all([
          apiRequest<Club[]>('/admin/clubs', {}, token),
          apiRequest<AccessRequest[]>('/admin/access-requests', {}, token),
          apiRequest<ClubRegistrationRequest[]>('/admin/club-registration-requests?status=pending', {}, token),
        ])
        if (cancelled) return
        setClubs(clubList)
        setAccessPending(accessList.filter((r) => r.status === 'pending').length)
        setClubRegsPending(regList.length)
      } catch (err) {
        if (!cancelled) setError(err instanceof Error ? err.message : 'Impossible de charger le tableau de bord')
      } finally {
        if (!cancelled) setLoading(false)
      }
    })()
    return () => { cancelled = true }
  }, [token])

  const activeClubs = clubs.filter((c) => c.is_active).length
  const hour = new Date().getHours()
  const greeting = hour < 12 ? 'Bonjour' : hour < 18 ? 'Bon après-midi' : 'Bonsoir'
  const name = user?.first_name || 'Admin'
  const recent = [...clubs].sort((a, b) => +new Date(b.created_at) - +new Date(a.created_at)).slice(0, 4)

  return (
    <div className="stack gap-lg">
      <header className="page-header">
        <h1>{greeting}, {name}</h1>
        <p>Voici l&apos;état de la plateforme Rotaract CIV.</p>
      </header>

      {error && <p className="error" role="alert">{error}</p>}

      <section className="stat-grid" aria-label="Indicateurs">
        <article className="stat-card">
          <div className="stat-label">Clubs actifs</div>
          <div className="stat-value">{loading ? '—' : activeClubs}</div>
          <div className="stat-hint">{clubs.length} au total</div>
        </article>
        <article className="stat-card">
          <div className="stat-label">Accès en attente</div>
          <div className="stat-value">{loading ? '—' : accessPending}</div>
          <div className="stat-hint">Demandes membres</div>
        </article>
        <article className="stat-card">
          <div className="stat-label">Nouveaux clubs</div>
          <div className="stat-value">{loading ? '—' : clubRegsPending}</div>
          <div className="stat-hint">Inscriptions à revoir</div>
        </article>
        <article className="stat-card">
          <div className="stat-label">À traiter</div>
          <div className="stat-value">{loading ? '—' : accessPending + clubRegsPending}</div>
          <div className="stat-hint">Actions prioritaires</div>
        </article>
      </section>

      <section className="stack">
        <h2>Actions rapides</h2>
        <div className="quick-grid">
          <Link to="/clubs" className="quick-link">
            <span className="quick-icon"><IconClubs /></span>
            <span>
              Gérer les clubs
              <small>Liste, création, présidents</small>
            </span>
          </Link>
          <Link to="/access-requests" className="quick-link">
            <span className="quick-icon"><IconAccess /></span>
            <span>
              Demandes d&apos;accès
              <small>{accessPending} en attente</small>
            </span>
          </Link>
          <Link to="/club-registrations" className="quick-link">
            <span className="quick-icon"><IconRegister /></span>
            <span>
              Inscriptions de clubs
              <small>{clubRegsPending} à examiner</small>
            </span>
          </Link>
        </div>
      </section>

      <section className="stack">
        <div className="row-between">
          <h2>Clubs récents</h2>
          <Link to="/clubs" className="btn-ghost btn-small ghost" style={{ display: 'inline-flex' }}>
            Voir tout
          </Link>
        </div>
        {recent.length === 0 && !loading ? (
          <div className="panel empty-state">
            <strong>Aucun club pour l&apos;instant</strong>
            <p>Créez le premier club pour démarrer la plateforme.</p>
            <div className="page-actions" style={{ justifyContent: 'center' }}>
              <Link to="/clubs" className="btn-primary"><IconPlus style={{ width: 18, height: 18 }} /> Créer un club</Link>
            </div>
          </div>
        ) : (
          <div className="club-list">
            {recent.map((c) => (
              <Link key={c.id} to={`/clubs/${c.id}`} className="club-item">
                {c.logo_url ? (
                  <img src={c.logo_url} alt="" className="club-logo-sm" />
                ) : (
                  <span className="brand-mark" style={{ width: 48, height: 48 }}>{c.name[0]}</span>
                )}
                <div className="club-item-body">
                  <strong>{c.name}</strong>
                  <div className="club-item-meta">
                    {[c.commune, c.city].filter(Boolean).join(', ') || '—'}
                    {!c.is_active && <> · <span className="badge inactive">inactif</span></>}
                  </div>
                </div>
                <span className="club-item-arrow" aria-hidden>›</span>
              </Link>
            ))}
          </div>
        )}
        {recent.length > 0 && (
          <div className="panel table-wrap">
            <table className="table">
              <thead>
                <tr>
                  <th>Club</th>
                  <th>Localisation</th>
                  <th>Code</th>
                  <th></th>
                </tr>
              </thead>
              <tbody>
                {recent.map((c) => (
                  <tr key={c.id}>
                    <td>
                      <div className="club-row" style={{ display: 'flex', alignItems: 'center', gap: '0.65rem' }}>
                        {c.logo_url && <img src={c.logo_url} alt="" className="club-logo-sm" />}
                        <span>{c.name}</span>
                      </div>
                    </td>
                    <td>{[c.commune, c.city].filter(Boolean).join(', ') || '—'}</td>
                    <td><code>{c.invite_code}</code></td>
                    <td><Link to={`/clubs/${c.id}`}>Ouvrir</Link></td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>
    </div>
  )
}
