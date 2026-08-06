import { useEffect, useState } from 'react'
import { apiRequest } from '../api/client'
import type { AccessRequest } from '../api/types'
import { useAuth } from '../auth/AuthContext'

type Tab = 'pending' | 'all'

export function AccessRequestsPage() {
  const { token } = useAuth()
  const [requests, setRequests] = useState<AccessRequest[]>([])
  const [tab, setTab] = useState<Tab>('pending')
  const [error, setError] = useState('')
  const [busyId, setBusyId] = useState<string | null>(null)

  async function load() {
    if (!token) return
    const data = await apiRequest<AccessRequest[]>('/admin/access-requests', {}, token)
    setRequests(data)
  }

  useEffect(() => {
    load().catch((err) => setError(err instanceof Error ? err.message : 'Erreur'))
  }, [token])

  async function approve(id: string) {
    if (!token) return
    setBusyId(id)
    try {
      await apiRequest(`/admin/access-requests/${id}/approve`, { method: 'POST', body: '{}' }, token)
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
                </div>
                <span className={statusBadge(r.status)}>{r.status}</span>
              </header>
              <p className="small">
                Club : <strong>{r.club_name}</strong>
                <span className="muted"> · {new Date(r.created_at).toLocaleDateString('fr-FR')}</span>
              </p>
              {r.status === 'pending' && (
                <div className="actions">
                  <button type="button" className="btn-small" disabled={busyId === r.id} onClick={() => approve(r.id)}>
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
