import { useEffect, useMemo, useState } from 'react'
import { apiRequest } from '../api/client'
import type { AccessRequest, Club } from '../api/types'
import { useAuth } from '../auth/AuthContext'

type Tab = 'pending' | 'all'

function normalizeClubKey(name: string): string {
  let s = name.toLowerCase().trim().replace(/uugb/g, 'iugb')
  for (const word of ['rotaract', 'club', 'de', 'du', 'la', 'le', 'les', 'rotary']) {
    s = s.replaceAll(word, '')
  }
  return s.replace(/[^a-z0-9]/g, '')
}

function suggestClubId(clubs: Club[], clubName: string): string {
  const key = normalizeClubKey(clubName)
  for (const club of clubs) {
    if (normalizeClubKey(club.name) === key) return club.id
  }
  const lower = clubName.toLowerCase()
  for (const club of clubs) {
    const name = club.name.toLowerCase()
    if (name.includes(lower) || lower.includes(name)) return club.id
  }
  return clubs[0]?.id ?? ''
}

export function AccessRequestsPage() {
  const { token } = useAuth()
  const [requests, setRequests] = useState<AccessRequest[]>([])
  const [clubs, setClubs] = useState<Club[]>([])
  const [clubChoices, setClubChoices] = useState<Record<string, string>>({})
  const [tab, setTab] = useState<Tab>('pending')
  const [error, setError] = useState('')
  const [busyId, setBusyId] = useState<string | null>(null)

  const activeClubs = useMemo(() => clubs.filter((c) => c.is_active), [clubs])

  async function load() {
    if (!token) return
    const [requestData, clubData] = await Promise.all([
      apiRequest<AccessRequest[]>('/admin/access-requests', {}, token),
      apiRequest<Club[]>('/admin/clubs', {}, token),
    ])
    setRequests(requestData)
    setClubs(clubData)
    setClubChoices((prev) => {
      const next = { ...prev }
      for (const req of requestData) {
        if (!req.club_id && !next[req.id]) {
          next[req.id] = suggestClubId(clubData.filter((c) => c.is_active), req.club_name)
        }
      }
      return next
    })
  }

  useEffect(() => {
    load().catch((err) => setError(err instanceof Error ? err.message : 'Erreur'))
  }, [token])

  async function approve(request: AccessRequest) {
    if (!token) return
    const body: { club_id?: string } = {}
    if (!request.club_id) {
      const clubId = clubChoices[request.id]
      if (!clubId) {
        setError('Choisissez le club à associer avant d’approuver.')
        return
      }
      body.club_id = clubId
    }

    setBusyId(request.id)
    setError('')
    try {
      await apiRequest(`/admin/access-requests/${request.id}/approve`, {
        method: 'POST',
        body: JSON.stringify(body),
      }, token)
      await load()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Action impossible')
    } finally {
      setBusyId(null)
    }
  }

  async function reject(id: string) {
    if (!token) return
    setBusyId(id)
    setError('')
    try {
      await apiRequest(`/admin/access-requests/${id}/reject`, { method: 'POST', body: '{}' }, token)
      await load()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Action impossible')
    } finally {
      setBusyId(null)
    }
  }

  const pending = requests.filter((r) => r.status === 'pending')
  const visible = tab === 'pending' ? pending : requests

  function statusBadge(status: string) {
    if (status === 'pending') return 'badge pending'
    if (status === 'approved') return 'badge approved'
    if (status === 'rejected') return 'badge rejected'
    return 'badge'
  }

  return (
    <div className="stack gap-lg">
      <header className="page-header">
        <h1>Demandes d&apos;accès</h1>
        <p>{pending.length} en attente de validation</p>
      </header>

      <div className="tabs" role="tablist">
        <button type="button" className={`tab ${tab === 'pending' ? 'active' : ''}`} onClick={() => setTab('pending')}>
          En attente ({pending.length})
        </button>
        <button type="button" className={`tab ${tab === 'all' ? 'active' : ''}`} onClick={() => setTab('all')}>
          Toutes ({requests.length})
        </button>
      </div>

      {error && <p className="error" role="alert">{error}</p>}

      {visible.length === 0 ? (
        <div className="panel empty-state">
          <strong>Rien à traiter</strong>
          <p>{tab === 'pending' ? 'Aucune demande en attente pour le moment.' : 'Aucune demande enregistrée.'}</p>
        </div>
      ) : (
        <div className="request-list">
          {visible.map((r) => (
            <article key={r.id} className="request-card">
              <header>
                <div>
                  <h3>{r.first_name} {r.last_name}</h3>
                  <p className="muted small">{r.email}</p>
                  {r.known_member ? <p className="muted small">Compte déjà en base (tombola / club)</p> : null}
                </div>
                <span className={statusBadge(r.status)}>{r.status}</span>
              </header>
              <p className="small">
                Club demandé : <strong>{r.club_name}</strong>
                <span className="muted"> · {new Date(r.created_at).toLocaleDateString('fr-FR')}</span>
              </p>
              {r.status === 'pending' && !r.club_id && (
                <label className="gap-sm" style={{ marginTop: '0.75rem' }}>
                  Club à associer
                  <select
                    value={clubChoices[r.id] ?? ''}
                    onChange={(e) => setClubChoices((prev) => ({ ...prev, [r.id]: e.target.value }))}
                    required
                  >
                    <option value="">— Choisir un club —</option>
                    {activeClubs.map((club) => (
                      <option key={club.id} value={club.id}>{club.name}</option>
                    ))}
                  </select>
                  <span className="muted small">
                    Le nom saisi ne correspond pas exactement à un club enregistré. Choisissez le bon club avant validation.
                  </span>
                </label>
              )}
              {r.status === 'pending' && (
                <div className="actions">
                  <button
                    type="button"
                    className="btn-small"
                    disabled={busyId === r.id || (!r.club_id && !clubChoices[r.id])}
                    onClick={() => approve(r)}
                  >
                    Approuver
                  </button>
                  <button type="button" className="btn-small ghost" disabled={busyId === r.id} onClick={() => reject(r.id)}>
                    Rejeter
                  </button>
                </div>
              )}
            </article>
          ))}
        </div>
      )}
    </div>
  )
}
