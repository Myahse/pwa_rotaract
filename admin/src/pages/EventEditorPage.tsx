import { useEffect, useState, type FormEvent } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { apiRequest, apiUpload } from '../api/client'
import type { PublicEvent } from '../api/types'
import { useAuth } from '../auth/AuthContext'

const empty: Omit<PublicEvent, 'id' | 'images' | 'status' | 'created_at' | 'updated_at' | 'flyer_url'> = {
  published: false,
  starts_at: '',
  city: '',
  venue_fr: '',
  venue_en: '',
  title_fr: '',
  title_en: '',
  summary_fr: '',
  summary_en: '',
  body_fr: '',
  body_en: '',
  cta_label_fr: '',
  cta_label_en: '',
  cta_url: '',
}

function toLocal(iso: string) {
  if (!iso) return ''
  const d = new Date(iso)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}

export function EventEditorPage() {
  const { eventId } = useParams()
  const { token } = useAuth()
  const navigate = useNavigate()
  const [form, setForm] = useState(empty)
  const [event, setEvent] = useState<PublicEvent | null>(null)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')
  const [busy, setBusy] = useState('')

  async function load() {
    if (!token || !eventId) return
    const data = await apiRequest<PublicEvent>(`/admin/events/${eventId}`, {}, token)
    setEvent(data)
    setForm({
      published: data.published,
      starts_at: data.starts_at,
      city: data.city,
      venue_fr: data.venue_fr,
      venue_en: data.venue_en,
      title_fr: data.title_fr,
      title_en: data.title_en,
      summary_fr: data.summary_fr,
      summary_en: data.summary_en,
      body_fr: data.body_fr,
      body_en: data.body_en,
      cta_label_fr: data.cta_label_fr,
      cta_label_en: data.cta_label_en,
      cta_url: data.cta_url,
    })
  }

  useEffect(() => {
    load().catch((err) => setError(err instanceof Error ? err.message : 'Erreur'))
  }, [token, eventId])

  function patch<K extends keyof typeof empty>(key: K, value: (typeof empty)[K]) {
    setForm((current) => ({ ...current, [key]: value }))
  }

  async function save(e: FormEvent) {
    e.preventDefault()
    if (!token || !eventId) return
    setBusy('save')
    setError('')
    setSuccess('')
    try {
      const data = await apiRequest<PublicEvent>(`/admin/events/${eventId}`, {
        method: 'PATCH',
        body: JSON.stringify({
          ...form,
          starts_at: new Date(form.starts_at).toISOString(),
        }),
      }, token)
      setEvent(data)
      setSuccess('Page enregistrée. Le site public se met à jour tout de suite si l’événement est publié.')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Enregistrement impossible')
    } finally {
      setBusy('')
    }
  }

  async function upload(kind: 'flyer' | 'images', files: FileList | null) {
    if (!token || !eventId || !files?.length) return
    setBusy(kind)
    setError('')
    try {
      let data = event
      for (const file of [...files]) {
        const body = new FormData()
        body.append('file', file)
        const path = kind === 'flyer' ? `/admin/events/${eventId}/flyer` : `/admin/events/${eventId}/images`
        data = await apiUpload<PublicEvent>(path, body, token)
      }
      if (data) setEvent(data)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Upload impossible')
    } finally {
      setBusy('')
    }
  }

  async function removeImage(imageId: string) {
    if (!token || !eventId) return
    const data = await apiRequest<PublicEvent>(`/admin/events/${eventId}/images/${imageId}`, { method: 'DELETE' }, token)
    setEvent(data)
  }

  async function remove() {
    if (!token || !eventId || !window.confirm('Supprimer cet événement du site public ?')) return
    await apiRequest(`/admin/events/${eventId}`, { method: 'DELETE' }, token)
    navigate('/events')
  }

  if (!event) {
    return <p className="muted">Chargement…</p>
  }

  const siteURL = `${import.meta.env.VITE_WEBSITE_URL || 'http://localhost:5175'}/events/${event.id}`

  return (
    <div className="stack gap-lg">
      <header className="page-header">
        <Link to="/events" className="muted small">← Événements</Link>
        <h1>{form.title_fr || 'Page événement'}</h1>
        <p>Affiche, texte et page détail — comme une campagne, pour le site public.</p>
      </header>

      {error ? <p className="error" role="alert">{error}</p> : null}
      {success ? <p className="success">{success}</p> : null}

      <form className="stack gap-lg panel" onSubmit={(e) => void save(e)}>
        <label className="checkbox">
          <input type="checkbox" checked={form.published} onChange={(e) => patch('published', e.target.checked)} />
          Publier sur le site public
        </label>
        {form.published ? (
          <p className="small">
            Page publique : <a href={siteURL} target="_blank" rel="noreferrer">{siteURL}</a>
          </p>
        ) : null}

        <label>
          Date et heure
          <input
            type="datetime-local"
            value={toLocal(form.starts_at)}
            onChange={(e) => {
              const next = e.target.value
              const iso = next ? new Date(next).toISOString() : ''
              patch('starts_at', iso)
            }}
            required
          />
        </label>
        <label>
          Ville
          <input value={form.city} onChange={(e) => patch('city', e.target.value)} />
        </label>

        <div className="grid-2">
          <label>
            Titre (FR)
            <input value={form.title_fr} onChange={(e) => patch('title_fr', e.target.value)} required />
          </label>
          <label>
            Title (EN)
            <input value={form.title_en} onChange={(e) => patch('title_en', e.target.value)} />
          </label>
          <label>
            Lieu (FR)
            <input value={form.venue_fr} onChange={(e) => patch('venue_fr', e.target.value)} />
          </label>
          <label>
            Venue (EN)
            <input value={form.venue_en} onChange={(e) => patch('venue_en', e.target.value)} />
          </label>
          <label className="span-2">
            Accroche (FR)
            <textarea value={form.summary_fr} onChange={(e) => patch('summary_fr', e.target.value)} rows={3} />
          </label>
          <label className="span-2">
            Summary (EN)
            <textarea value={form.summary_en} onChange={(e) => patch('summary_en', e.target.value)} rows={3} />
          </label>
          <label className="span-2">
            Page détail (FR)
            <textarea value={form.body_fr} onChange={(e) => patch('body_fr', e.target.value)} rows={8} />
          </label>
          <label className="span-2">
            Detail page (EN)
            <textarea value={form.body_en} onChange={(e) => patch('body_en', e.target.value)} rows={8} />
          </label>
          <label>
            Bouton (FR)
            <input value={form.cta_label_fr} onChange={(e) => patch('cta_label_fr', e.target.value)} />
          </label>
          <label>
            Button (EN)
            <input value={form.cta_label_en} onChange={(e) => patch('cta_label_en', e.target.value)} />
          </label>
          <label className="span-2">
            Lien du bouton
            <input value={form.cta_url} onChange={(e) => patch('cta_url', e.target.value)} placeholder="https://" />
          </label>
        </div>

        <div className="actions">
          <button type="submit" className="btn-primary" disabled={busy === 'save'}>
            {busy === 'save' ? 'Enregistrement…' : 'Enregistrer la page'}
          </button>
          <button type="button" className="btn-small ghost" onClick={() => void remove()}>
            Supprimer
          </button>
        </div>
      </form>

      <section className="panel stack">
        <h2>Affiche / flyer</h2>
        {event.flyer_url ? <img src={event.flyer_url} alt="" className="event-flyer" /> : <p className="muted">Aucune affiche pour le moment.</p>}
        <label>
          Ajouter une affiche
          <input type="file" accept="image/jpeg,image/png,image/webp" onChange={(e) => void upload('flyer', e.target.files)} />
        </label>
      </section>

      <section className="panel stack">
        <h2>Photos de la page</h2>
        <div className="event-gallery">
          {(event.images ?? []).map((image) => (
            <figure key={image.id}>
              <img src={image.url} alt="" />
              <button type="button" className="btn-small ghost" onClick={() => void removeImage(image.id)}>
                Retirer
              </button>
            </figure>
          ))}
        </div>
        <label>
          Ajouter des photos
          <input type="file" accept="image/jpeg,image/png,image/webp" multiple onChange={(e) => void upload('images', e.target.files)} />
        </label>
      </section>
    </div>
  )
}
