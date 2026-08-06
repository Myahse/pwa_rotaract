import { useCallback, useEffect, useState, type FormEvent } from 'react'
import { apiRequest } from '../api/client'
import type {
  ApproveClubRegistrationResult,
  ClubRegistrationRequest,
  ClubRegistrationStatus,
} from '../api/types'
import { useAuth } from '../auth/AuthContext'

type Tab = ClubRegistrationStatus | 'all'

export function ClubRegistrationsPage() {
  const { token } = useAuth()
  const [requests, setRequests] = useState<ClubRegistrationRequest[]>([])
  const [tab, setTab] = useState<Tab>('pending')
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')
  const [selected, setSelected] = useState<ClubRegistrationRequest | null>(null)
  const [busy, setBusy] = useState(false)
  const [approveForm, setApproveForm] = useState({ slug: '', country: '', city: '', commune: '', review_note: '' })
  const [rejectNote, setRejectNote] = useState('')

  const load = useCallback(async (filter: Tab) => {
    if (!token) return
    if (filter === 'all') {
      const [p, a, r] = await Promise.all([
        apiRequest<ClubRegistrationRequest[]>('/admin/club-registration-requests?status=pending', {}, token),
        apiRequest<ClubRegistrationRequest[]>('/admin/club-registration-requests?status=approved', {}, token),
        apiRequest<ClubRegistrationRequest[]>('/admin/club-registration-requests?status=rejected', {}, token),
      ])
      const byId = new Map<string, ClubRegistrationRequest>()
      ;[...p, ...a, ...r].forEach((item) => byId.set(item.id, item))
      setRequests([...byId.values()].sort((x, y) => +new Date(y.created_at) - +new Date(x.created_at)))
      return
    }
    const data = await apiRequest<ClubRegistrationRequest[]>(
      `/admin/club-registration-requests?status=${filter}`,
      {},
      token,
    )
    setRequests(data)
  }, [token])

  useEffect(() => {
    load(tab).catch((err) => setError(err instanceof Error ? err.message : 'Erreur'))
  }, [load, tab])

  function openReview(req: ClubRegistrationRequest) {
    setSelected(req)
    setApproveForm({
      slug: '',
      country: req.country,
      city: req.city,
      commune: req.commune,
      review_note: '',
    })
    setRejectNote('')
    setError('')
    setSuccess('')
  }

  async function onApprove(e: FormEvent) {
    e.preventDefault()
    if (!token || !selected) return
    setBusy(true)
    setError('')
    try {
      const result = await apiRequest<ApproveClubRegistrationResult>(
        `/admin/club-registration-requests/${selected.id}/approve`,
        {
          method: 'POST',
          body: JSON.stringify({
            slug: approveForm.slug || undefined,
            country: approveForm.country || undefined,
            city: approveForm.city || undefined,
            commune: approveForm.commune || undefined,
            review_note: approveForm.review_note || undefined,
          }),
        },
        token,
      )
      setSuccess(result.message || "Demande approuvée — un lien d'accès a été envoyé.")
      setSelected(null)
      await load(tab)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Approbation impossible')
    } finally {
      setBusy(false)
    }
  }

  async function onReject(e: FormEvent) {
    e.preventDefault()
    if (!token || !selected) return
    setBusy(true)
    setError('')
    try {
      await apiRequest(`/admin/club-registration-requests/${selected.id}/reject`, {
        method: 'POST',
        body: JSON.stringify({ note: rejectNote || undefined }),
      }, token)
      setSuccess('Demande rejetée.')
      setSelected(null)
      await load(tab)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Rejet impossible')
    } finally {
      setBusy(false)
    }
  }

  function statusBadge(status: string) {
    if (status === 'pending') return 'badge pending'
    if (status === 'approved') return 'badge approved'
    if (status === 'rejected') return 'badge rejected'
    return 'badge'
  }

  return (
    <div className="stack gap-lg">
      <header className="page-header">
        <h1>Inscriptions de clubs</h1>
        <p>Demandes publiques pour créer un nouveau club Rotaract</p>
      </header>

      <div className="tabs" role="tablist">
        {([
          ['pending', 'En attente'],
          ['approved', 'Approuvées'],
          ['rejected', 'Rejetées'],
          ['all', 'Toutes'],
        ] as const).map(([key, label]) => (
          <button
            key={key}
            type="button"
            className={`tab ${tab === key ? 'active' : ''}`}
            onClick={() => setTab(key)}
          >
            {label}
            {key === 'pending' && tab === 'pending' ? ` (${requests.length})` : ''}
          </button>
        ))}
      </div>

      {success && <p className="success" role="status">{success}</p>}
      {error && !selected && <p className="error" role="alert">{error}</p>}

      {requests.length === 0 ? (
        <div className="panel empty-state">
          <strong>Aucune demande</strong>
          <p>
            {tab === 'pending'
              ? 'Aucune inscription de club en attente.'
              : 'Rien à afficher pour ce filtre.'}
          </p>
        </div>
      ) : (
        <div className="request-list">
          {requests.map((r) => (
            <article key={r.id} className="request-card">
              <header>
                <div>
                  <h3>{r.club_name}</h3>
                  <p className="muted small">
                    {r.contact_first_name} {r.contact_last_name} · {r.contact_email}
                  </p>
                </div>
                <span className={statusBadge(r.status)}>{r.status}</span>
              </header>
              <p className="small">
                {[r.commune, r.city, r.country].filter(Boolean).join(', ')}
                <span className="muted"> · {new Date(r.created_at).toLocaleDateString('fr-FR')}</span>
              </p>
              {r.message && <p className="muted small">{r.message}</p>}
              {r.status === 'pending' && (
                <div className="actions">
                  <button type="button" className="btn-small" onClick={() => openReview(r)}>
                    Examiner
                  </button>
                </div>
              )}
            </article>
          ))}
        </div>
      )}

      {selected && (
        <>
          <div className="drawer-backdrop" onClick={() => setSelected(null)} aria-hidden />
          <div className="drawer" role="dialog" aria-modal="true" aria-labelledby="review-title">
            <div className="drawer-handle" />
            <div className="stack gap-lg">
              <div className="row-between">
                <div>
                  <h2 id="review-title">{selected.club_name}</h2>
                  <p className="muted small">
                    {selected.contact_first_name} {selected.contact_last_name} · {selected.contact_email}
                  </p>
                </div>
                <button type="button" className="btn-ghost btn-small ghost" onClick={() => setSelected(null)}>Fermer</button>
              </div>

              <div className="panel stack" style={{ background: 'var(--bg)', boxShadow: 'none' }}>
                <p className="small"><strong>Localisation</strong> — {[selected.commune, selected.city, selected.country].join(', ')}</p>
                {selected.phone && <p className="small"><strong>Tél.</strong> — {selected.phone}</p>}
                {selected.description && <p className="small"><strong>Description</strong> — {selected.description}</p>}
                {selected.message && <p className="small"><strong>Message</strong> — {selected.message}</p>}
              </div>

              <form onSubmit={onApprove} className="stack">
                <h3>Approuver</h3>
                <p className="muted small">Un e-mail avec lien d&apos;accès sera envoyé au contact.</p>
                <div className="grid-2">
                  <label>
                    Slug (optionnel)
                    <input value={approveForm.slug} onChange={(e) => setApproveForm({ ...approveForm, slug: e.target.value })} placeholder="auto" />
                  </label>
                  <label>
                    Pays
                    <input value={approveForm.country} onChange={(e) => setApproveForm({ ...approveForm, country: e.target.value })} />
                  </label>
                  <label>
                    Ville
                    <input value={approveForm.city} onChange={(e) => setApproveForm({ ...approveForm, city: e.target.value })} />
                  </label>
                  <label>
                    Commune
                    <input value={approveForm.commune} onChange={(e) => setApproveForm({ ...approveForm, commune: e.target.value })} />
                  </label>
                  <label className="span-2">
                    Note interne
                    <input value={approveForm.review_note} onChange={(e) => setApproveForm({ ...approveForm, review_note: e.target.value })} />
                  </label>
                </div>
                {error && <p className="error" role="alert">{error}</p>}
                <button type="submit" className="btn-primary" disabled={busy}>
                  {busy ? 'Traitement…' : 'Approuver et envoyer le lien'}
                </button>
              </form>

              <form onSubmit={onReject} className="stack">
                <h3>Rejeter</h3>
                <label>
                  Motif (optionnel)
                  <textarea value={rejectNote} onChange={(e) => setRejectNote(e.target.value)} rows={3} />
                </label>
                <button type="submit" className="btn-danger" disabled={busy}>
                  Rejeter la demande
                </button>
              </form>
            </div>
          </div>
        </>
      )}
    </div>
  )
}
