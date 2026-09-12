import { useEffect, useMemo, useState, type FormEvent } from 'react'
import { Link, useParams } from 'react-router-dom'
import { apiRequest, apiUpload } from '../api/client'
import type { ClubDuePayment } from '../api/types'
import { useAuth } from '../auth/AuthContext'
import { prepareReceiptFile, RECEIPT_ACCEPT } from '../lib/receiptUpload'

const WAVE_PAY_URL = 'https://pay.wave.com/m/M_ci_pHlyZFYyH1Su/c/ci/'

function currentMonthValue() {
  const now = new Date()
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`
}

export function CotisationsPage() {
  const { clubId } = useParams<{ clubId: string }>()
  const { token, uiCaps } = useAuth()
  const [items, setItems] = useState<ClubDuePayment[]>([])
  const [dueMonth, setDueMonth] = useState(currentMonthValue())
  const [amount, setAmount] = useState('')
  const [file, setFile] = useState<File | null>(null)
  const [error, setError] = useState('')
  const [message, setMessage] = useState('')
  const [busy, setBusy] = useState(false)
  const [busyId, setBusyId] = useState<string | null>(null)

  const pending = useMemo(() => items.filter((item) => item.status === 'pending'), [items])

  async function load() {
    if (!token || !clubId) return
    const data = await apiRequest<ClubDuePayment[]>(`/clubs/${clubId}/dues`, {}, token)
    setItems(data)
  }

  useEffect(() => {
    load().catch((err) => setError(err instanceof Error ? err.message : 'Erreur'))
  }, [token, clubId])

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    if (!token || !clubId || !file) return
    const gift = Number(amount)
    if (!Number.isFinite(gift) || gift <= 0) {
      setError('Indiquez un montant valide.')
      return
    }
    setBusy(true)
    setError('')
    setMessage('')
    try {
      const prepared = await prepareReceiptFile(file)
      const body = new FormData()
      body.append('due_month', dueMonth)
      body.append('amount_xof', String(Math.round(gift)))
      body.append('file', prepared.blob, prepared.filename)
      await apiUpload<ClubDuePayment>(`/clubs/${clubId}/dues`, body, token)
      setMessage('Cotisation envoyée. Le responsable validera votre reçu Wave.')
      setAmount('')
      setFile(null)
      await load()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Envoi impossible')
    } finally {
      setBusy(false)
    }
  }

  async function markReceived(id: string) {
    if (!token || !clubId) return
    setBusyId(id)
    try {
      await apiRequest(`/clubs/${clubId}/dues/${id}/received`, { method: 'POST', body: '{}' }, token)
      await load()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Action impossible')
    } finally {
      setBusyId(null)
    }
  }

  return (
    <div className="stack gap-lg">
      <Link to={uiCaps.nav.club ? `/clubs/${clubId}` : '/home'} className="back">
        ← {uiCaps.nav.club ? 'Retour au club' : 'Retour à l\'accueil'}
      </Link>
      <header className="section-head">
        <div>
          <h2>Cotisations mensuelles</h2>
          <p className="muted small">{pending.length} en attente de validation</p>
        </div>
      </header>

      {error && <p className="error">{error}</p>}
      {message && <p className="success">{message}</p>}

      <section className="card stack">
        <h3>Déclarer un paiement</h3>
        <p className="muted small">
          Payez via Wave puis envoyez le reçu comme pour un don.
          {' '}
          <a href={WAVE_PAY_URL} target="_blank" rel="noopener noreferrer">Ouvrir Wave</a>
        </p>
        <form onSubmit={onSubmit} className="stack">
          <label>
            Mois
            <input type="month" value={dueMonth} onChange={(e) => setDueMonth(e.target.value)} required />
          </label>
          <label>
            Montant (F CFA)
            <input type="number" min={1} value={amount} onChange={(e) => setAmount(e.target.value)} required />
          </label>
          <label>
            Reçu Wave (photo ou PDF)
            <input
              type="file"
              accept={RECEIPT_ACCEPT}
              onChange={(e) => setFile(e.target.files?.[0] ?? null)}
              required
            />
          </label>
          <button type="submit" className="btn-primary" disabled={busy}>{busy ? 'Envoi…' : 'Envoyer la cotisation'}</button>
        </form>
      </section>

      <section className="stack">
        <h3>Historique</h3>
        {items.length === 0 ? (
          <p className="muted">Aucune cotisation enregistrée.</p>
        ) : (
          <div className="record-list">
            {items.map((item) => (
              <article key={item.id} className="record-card card">
                <header>
                  <div>
                    <h3>{item.user?.first_name} {item.user?.last_name}</h3>
                    <p className="muted small">{item.due_month}</p>
                  </div>
                  <span className={`badge ${item.status === 'received' ? 'approved' : 'pending'}`}>
                    {item.status === 'received' ? 'reçu' : 'en attente'}
                  </span>
                </header>
                <p className="small">
                  <strong>{item.amount_xof.toLocaleString('fr-FR')} F CFA</strong>
                  <span className="muted"> · {new Date(item.created_at).toLocaleDateString('fr-FR')}</span>
                </p>
                <div className="actions">
                  {item.receipt_url ? (
                    <a className="btn-small btn-outline" href={item.receipt_url} target="_blank" rel="noopener noreferrer">
                      Voir le reçu
                    </a>
                  ) : (
                    <span className="muted small">Reçu en attente</span>
                  )}
                  {uiCaps.canManageDues && item.status === 'pending' ? (
                    <button type="button" className="btn-small" disabled={busyId === item.id} onClick={() => markReceived(item.id)}>
                      Marquer reçu
                    </button>
                  ) : null}
                </div>
              </article>
            ))}
          </div>
        )}
      </section>
    </div>
  )
}
