import { useEffect, useState } from 'react'
import { apiRequest } from '../api/client'
import type { Donation } from '../api/types'
import { useAuth } from '../auth/AuthContext'

export function DonationsPage() {
  const { token } = useAuth()
  const [items, setItems] = useState<Donation[]>([])
  const [error, setError] = useState('')
  const [busyId, setBusyId] = useState<string | null>(null)

  async function load() {
    if (!token) return
    const data = await apiRequest<Donation[]>('/admin/donations', {}, token)
    setItems(data)
  }

  useEffect(() => {
    load().catch((err) => setError(err instanceof Error ? err.message : 'Erreur'))
  }, [token])

  async function markReceived(id: string) {
    if (!token) return
    setBusyId(id)
    try {
      await apiRequest(`/admin/donations/${id}/received`, { method: 'POST', body: '{}' }, token)
      await load()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Action impossible')
    } finally {
      setBusyId(null)
    }
  }

  const pending = items.filter((item) => item.status === 'pending')

  return (
    <div className="stack gap-lg">
      <header className="page-header">
        <h1>Dons</h1>
        <p>
          {pending.length} en attente · comparez le reçu Wave et le montant, puis marquez reçu (comme la tombola).
        </p>
      </header>

      {error && <p className="error" role="alert">{error}</p>}

      {items.length === 0 ? (
        <div className="panel empty-state">
          <strong>Aucun don</strong>
          <p>Les dons envoyés depuis le site public apparaîtront ici.</p>
        </div>
      ) : (
        <div className="request-list">
          {items.map((item) => (
            <article key={item.id} className="request-card">
              <header>
                <div>
                  <h3>{item.name}</h3>
                  <p className="muted small">{item.email}</p>
                </div>
                <span className={item.status === 'received' ? 'badge approved' : 'badge pending'}>
                  {item.status === 'received' ? 'reçu' : 'en attente'}
                </span>
              </header>
              <p className="small">
                <strong>{item.amount_xof.toLocaleString('fr-FR')} F CFA</strong>
                {item.known_member ? <span className="muted"> · compte club / tombola</span> : null}
                <span className="muted"> · {new Date(item.created_at).toLocaleDateString('fr-FR')}</span>
              </p>
              <div className="actions">
                {item.receipt_url ? (
                  <a
                    className="btn-small btn-outline"
                    href={item.receipt_url}
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    Voir le reçu
                  </a>
                ) : (
                  <span className="muted small">Reçu en attente</span>
                )}
                {item.status === 'pending' ? (
                  <button
                    type="button"
                    className="btn-small"
                    disabled={busyId === item.id}
                    onClick={() => markReceived(item.id)}
                  >
                    Marquer reçu
                  </button>
                ) : null}
              </div>
            </article>
          ))}
        </div>
      )}
    </div>
  )
}
