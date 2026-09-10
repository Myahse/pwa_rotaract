import { useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { apiRequest } from '../api/client'
import type { PublicEvent } from '../api/types'
import { useAuth } from '../auth/AuthContext'

export function EventsPage() {
  const { token } = useAuth()
  const navigate = useNavigate()
  const [items, setItems] = useState<PublicEvent[]>([])
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)

  async function load() {
    if (!token) return
    const data = await apiRequest<PublicEvent[]>('/admin/events', {}, token)
    setItems(data)
  }

  useEffect(() => {
    load().catch((err) => setError(err instanceof Error ? err.message : 'Erreur'))
  }, [token])

  async function create() {
    if (!token) return
    setBusy(true)
    setError('')
    try {
      const created = await apiRequest<PublicEvent>('/admin/events', {
        method: 'POST',
        body: JSON.stringify({
          published: false,
          starts_at: new Date().toISOString(),
          title_fr: 'Nouvel événement',
          title_en: 'New event',
        }),
      }, token)
      navigate(`/events/${created.id}`)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Création impossible')
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="stack gap-lg">
      <header className="page-header row-between">
        <div>
          <h1>Événements</h1>
          <p>Affiches, détails et page publique — visibles sur le site dès publication.</p>
        </div>
        <button type="button" className="btn-primary" disabled={busy} onClick={() => void create()}>
          {busy ? 'Création…' : 'Nouvel événement'}
        </button>
      </header>

      {error ? <p className="error" role="alert">{error}</p> : null}

      {items.length === 0 ? (
        <div className="panel empty-state">
          <strong>Aucun événement</strong>
          <p>Créez une page, ajoutez l’affiche et publiez pour le site public.</p>
        </div>
      ) : (
        <div className="request-list">
          {items.map((item) => (
            <article key={item.id} className="request-card">
              <header>
                <div>
                  <h3>{item.title_fr}</h3>
                  <p className="muted small">
                    {new Date(item.starts_at).toLocaleString('fr-FR')}
                    {item.city ? ` · ${item.city}` : ''}
                  </p>
                </div>
                <span className={item.published ? 'badge approved' : 'badge pending'}>
                  {item.published ? 'publié' : 'brouillon'}
                </span>
              </header>
              {item.flyer_url ? (
                <img src={item.flyer_url} alt="" className="event-thumb" />
              ) : null}
              <div className="actions">
                <Link className="btn-small" to={`/events/${item.id}`}>
                  Concevoir la page
                </Link>
              </div>
            </article>
          ))}
        </div>
      )}
    </div>
  )
}
