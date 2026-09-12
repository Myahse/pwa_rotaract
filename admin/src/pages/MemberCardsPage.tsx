import { useEffect, useMemo, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { apiRequest } from '../api/client'
import type { Club, MemberCard, MemberCardListResponse } from '../api/types'
import { useAuth } from '../auth/AuthContext'

export function MemberCardsPage() {
  const { token } = useAuth()
  const [searchParams] = useSearchParams()
  const [clubs, setClubs] = useState<Club[]>([])
  const [items, setItems] = useState<MemberCard[]>([])
  const [pendingCount, setPendingCount] = useState(0)
  const [clubFilter, setClubFilter] = useState(searchParams.get('club_id') ?? '')
  const [pendingOnly, setPendingOnly] = useState(true)
  const [error, setError] = useState('')
  const [message, setMessage] = useState('')
  const [busy, setBusy] = useState(false)
  const [busyId, setBusyId] = useState<string | null>(null)

  useEffect(() => {
    if (!token) return
    apiRequest<Club[]>('/admin/clubs', {}, token)
      .then(setClubs)
      .catch(() => setClubs([]))
  }, [token])

  async function load() {
    if (!token) return
    const params = new URLSearchParams()
    if (clubFilter) params.set('club_id', clubFilter)
    if (pendingOnly) params.set('pending_only', 'true')
    const query = params.toString()
    const data = await apiRequest<MemberCardListResponse>(
      `/admin/member-cards${query ? `?${query}` : ''}`,
      {},
      token,
    )
    setItems(data.items)
    setPendingCount(data.pending_count)
  }

  useEffect(() => {
    load().catch((err) => setError(err instanceof Error ? err.message : 'Erreur'))
  }, [token, clubFilter, pendingOnly])

  const clubNameById = useMemo(
    () => Object.fromEntries(clubs.map((club) => [club.id, club.name])),
    [clubs],
  )

  async function issueMissing() {
    if (!token) return
    setBusy(true)
    setError('')
    setMessage('')
    try {
      const params = clubFilter ? `?club_id=${clubFilter}` : ''
      const res = await apiRequest<{ created: number }>(
        `/admin/member-cards/issue-missing${params}`,
        { method: 'POST', body: '{}' },
        token,
      )
      setMessage(`${res.created} carte(s) générée(s) pour les membres sans carte.`)
      await load()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Action impossible')
    } finally {
      setBusy(false)
    }
  }

  async function sendPending() {
    if (!token) return
    setBusy(true)
    setError('')
    setMessage('')
    try {
      const params = clubFilter ? `?club_id=${clubFilter}` : ''
      const res = await apiRequest<{ sent: number; failed: number; errors?: string[] }>(
        `/admin/member-cards/send-pending${params}`,
        { method: 'POST', body: '{}' },
        token,
      )
      setMessage(`${res.sent} carte(s) envoyée(s)${res.failed ? ` · ${res.failed} échec(s)` : ''}.`)
      if (res.errors?.length) setError(res.errors.slice(0, 3).join(' · '))
      await load()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Envoi impossible')
    } finally {
      setBusy(false)
    }
  }

  async function sendOne(id: string) {
    if (!token) return
    setBusyId(id)
    setError('')
    setMessage('')
    try {
      await apiRequest(`/admin/member-cards/${id}/send`, { method: 'POST', body: '{}' }, token)
      setMessage('Carte envoyée par email.')
      await load()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Envoi impossible')
    } finally {
      setBusyId(null)
    }
  }

  return (
    <div className="stack gap-lg">
      <header className="page-header">
        <h1>Cartes de membre</h1>
        <p>
          PDF envoyé par email à la création du membre. {pendingCount} en attente d&apos;envoi.
        </p>
      </header>

      <section className="panel stack">
        <div className="actions wrap">
          <label>
            Club
            <select value={clubFilter} onChange={(e) => setClubFilter(e.target.value)}>
              <option value="">Tous les clubs</option>
              {clubs.map((club) => (
                <option key={club.id} value={club.id}>{club.name}</option>
              ))}
            </select>
          </label>
          <label className="checkbox-row">
            <input
              type="checkbox"
              checked={pendingOnly}
              onChange={(e) => setPendingOnly(e.target.checked)}
            />
            Non envoyées seulement
          </label>
        </div>
        <div className="actions wrap">
          <button type="button" className="btn-outline" disabled={busy} onClick={issueMissing}>
            Générer les cartes manquantes
          </button>
          <button type="button" className="btn-primary" disabled={busy || pendingCount === 0} onClick={sendPending}>
            Envoyer toutes les cartes en attente
          </button>
        </div>
      </section>

      {error && <p className="error" role="alert">{error}</p>}
      {message && <p className="success">{message}</p>}

      {items.length === 0 ? (
        <div className="panel empty-state">
          <strong>Aucune carte</strong>
          <p>Générez les cartes manquantes pour les membres existants, puis envoyez-les par email.</p>
        </div>
      ) : (
        <div className="request-list">
          {items.map((item) => (
            <article key={item.id} className="request-card">
              <header>
                <div>
                  <h3>{item.user?.first_name} {item.user?.last_name}</h3>
                  <p className="muted small">
                    {item.user?.email}
                    {' · '}
                    {item.club?.name ?? clubNameById[item.club_id] ?? 'Club'}
                  </p>
                </div>
                <span className={item.sent_at ? 'badge approved' : 'badge pending'}>
                  {item.sent_at ? 'envoyée' : 'en attente'}
                </span>
              </header>
              <p className="small">
                <strong>{item.card_number}</strong>
                <span className="muted">
                  {' · émise le '}
                  {new Date(item.issued_at).toLocaleDateString('fr-FR')}
                  {item.sent_at ? ` · envoyée le ${new Date(item.sent_at).toLocaleDateString('fr-FR')}` : ''}
                </span>
              </p>
              <div className="actions">
                <button
                  type="button"
                  className="btn-small"
                  disabled={busyId === item.id}
                  onClick={() => sendOne(item.id)}
                >
                  {item.sent_at ? 'Renvoyer' : 'Envoyer'}
                </button>
              </div>
            </article>
          ))}
        </div>
      )}
    </div>
  )
}
