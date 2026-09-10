import { useEffect, useState, type FormEvent } from 'react'
import { apiRequest, apiUpload } from '../api/client'
import type { FeaturedPostulant, SiteGalleryImage } from '../api/types'
import { useAuth } from '../auth/AuthContext'
import { formatPostulantPeriod, previousCalendarMonth } from '../lib/postulantPeriod'

const websiteUrl = import.meta.env.VITE_WEBSITE_URL || 'http://localhost:5175'

type PostulantForm = Omit<FeaturedPostulant, 'id' | 'flyer_url' | 'updated_at'>

function emptyPostulantForm(): PostulantForm {
  const prev = previousCalendarMonth()
  return {
    period_year: prev.year,
    period_month: prev.month,
    published: false,
    first_name: '',
    last_name: '',
    home_club: '',
    quote_fr: '',
    quote_en: '',
    visit_count: 0,
    clubs_visited: [],
  }
}

function postulantToForm(item: FeaturedPostulant): PostulantForm {
  return {
    period_year: item.period_year,
    period_month: item.period_month,
    published: item.published,
    first_name: item.first_name,
    last_name: item.last_name,
    home_club: item.home_club,
    quote_fr: item.quote_fr,
    quote_en: item.quote_en,
    visit_count: item.visit_count,
    clubs_visited: item.clubs_visited ?? [],
  }
}

export function WebsiteHomePage() {
  const { token } = useAuth()
  const [gallery, setGallery] = useState<SiteGalleryImage[]>([])
  const [postulants, setPostulants] = useState<FeaturedPostulant[]>([])
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [form, setForm] = useState<PostulantForm>(emptyPostulantForm)
  const [clubsText, setClubsText] = useState('')
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')
  const [busy, setBusy] = useState('')

  async function load() {
    if (!token) return
    const [galleryData, postulantData] = await Promise.all([
      apiRequest<SiteGalleryImage[]>('/admin/gallery', {}, token),
      apiRequest<FeaturedPostulant[]>('/admin/featured-postulants', {}, token),
    ])
    setGallery(galleryData)
    setPostulants(postulantData)
    const prev = previousCalendarMonth()
    const current =
      postulantData.find((item) => item.period_year === prev.year && item.period_month === prev.month) ??
      postulantData[0] ??
      null
    setSelectedId(current?.id ?? null)
    if (current) {
      setForm(postulantToForm(current))
      setClubsText((current.clubs_visited ?? []).join('\n'))
    } else {
      setForm(emptyPostulantForm())
      setClubsText('')
    }
  }

  const selectedPostulant = postulants.find((item) => item.id === selectedId) ?? null

  function selectPostulant(item: FeaturedPostulant) {
    setSelectedId(item.id)
    setForm(postulantToForm(item))
    setClubsText((item.clubs_visited ?? []).join('\n'))
    setSuccess('')
    setError('')
  }

  function startNewPostulant() {
    setSelectedId(null)
    setForm(emptyPostulantForm())
    setClubsText('')
    setSuccess('')
    setError('')
  }

  useEffect(() => {
    load().catch((err) => setError(err instanceof Error ? err.message : 'Erreur'))
  }, [token])

  async function addGalleryItem() {
    if (!token) return
    setBusy('gallery-add')
    setError('')
    try {
      const item = await apiRequest<SiteGalleryImage>('/admin/gallery', {
        method: 'POST',
        body: JSON.stringify({ published: false, caption_fr: '', caption_en: '' }),
      }, token)
      setGallery((current) => [...current, item])
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Création impossible')
    } finally {
      setBusy('')
    }
  }

  async function updateGalleryItem(item: SiteGalleryImage) {
    if (!token) return
    setBusy(`gallery-${item.id}`)
    setError('')
    try {
      const data = await apiRequest<SiteGalleryImage>(`/admin/gallery/${item.id}`, {
        method: 'PATCH',
        body: JSON.stringify({
          published: item.published,
          caption_fr: item.caption_fr,
          caption_en: item.caption_en,
          sort_order: item.sort_order,
        }),
      }, token)
      setGallery((current) => current.map((row) => (row.id === data.id ? data : row)))
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Mise à jour impossible')
    } finally {
      setBusy('')
    }
  }

  async function uploadGalleryImage(id: string, files: FileList | null) {
    if (!token || !files?.length) return
    setBusy(`upload-${id}`)
    setError('')
    try {
      const body = new FormData()
      body.append('file', files[0])
      const data = await apiUpload<SiteGalleryImage>(`/admin/gallery/${id}/image`, body, token)
      setGallery((current) => current.map((row) => (row.id === data.id ? data : row)))
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Upload impossible')
    } finally {
      setBusy('')
    }
  }

  async function removeGalleryItem(id: string) {
    if (!token || !window.confirm('Supprimer cette photo de la galerie ?')) return
    setBusy(`delete-${id}`)
    setError('')
    try {
      await apiRequest(`/admin/gallery/${id}`, { method: 'DELETE' }, token)
      setGallery((current) => current.filter((row) => row.id !== id))
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Suppression impossible')
    } finally {
      setBusy('')
    }
  }

  async function savePostulant(e: FormEvent) {
    e.preventDefault()
    if (!token) return
    setBusy('postulant')
    setError('')
    setSuccess('')
    try {
      const clubs = clubsText
        .split('\n')
        .map((line) => line.trim())
        .filter(Boolean)
      const payload = { ...form, clubs_visited: clubs }
      const data = selectedId
        ? await apiRequest<FeaturedPostulant>(`/admin/featured-postulants/${selectedId}`, {
            method: 'PATCH',
            body: JSON.stringify(payload),
          }, token)
        : await apiRequest<FeaturedPostulant>('/admin/featured-postulants', {
            method: 'POST',
            body: JSON.stringify(payload),
          }, token)
      setPostulants((current) => {
        const next = current.filter((row) => row.id !== data.id)
        next.unshift(data)
        return next.sort((a, b) => b.period_year - a.period_year || b.period_month - a.period_month)
      })
      setSelectedId(data.id)
      setSuccess('Meilleur postulant enregistré.')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Enregistrement impossible')
    } finally {
      setBusy('')
    }
  }

  async function uploadPostulantFlyer(files: FileList | null) {
    if (!token || !files?.length || !selectedId) return
    setBusy('postulant-flyer')
    setError('')
    setSuccess('')
    try {
      const body = new FormData()
      body.append('file', files[0])
      const data = await apiUpload<FeaturedPostulant>(`/admin/featured-postulants/${selectedId}/flyer`, body, token)
      setPostulants((current) => current.map((row) => (row.id === data.id ? data : row)))
      setSuccess('Flyer enregistré.')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Upload impossible')
    } finally {
      setBusy('')
    }
  }

  async function removePostulant() {
    if (!token || !selectedId) return
    if (!window.confirm('Supprimer ce postulant du mois ?')) return
    setBusy('postulant-delete')
    setError('')
    setSuccess('')
    try {
      await apiRequest(`/admin/featured-postulants/${selectedId}`, { method: 'DELETE' }, token)
      const remaining = postulants.filter((row) => row.id !== selectedId)
      setPostulants(remaining)
      if (remaining[0]) selectPostulant(remaining[0])
      else startNewPostulant()
      setSuccess('Postulant supprimé.')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Suppression impossible')
    } finally {
      setBusy('')
    }
  }

  function postulantFlyerSrc(item: FeaturedPostulant | null) {
    if (!item?.flyer_url) return ''
    const stamp = encodeURIComponent(item.updated_at)
    return `${item.flyer_url}${item.flyer_url.includes('?') ? '&' : '?'}v=${stamp}`
  }

  function patchGallery(id: string, patch: Partial<SiteGalleryImage>) {
    setGallery((current) => current.map((row) => (row.id === id ? { ...row, ...patch } : row)))
  }

  return (
    <div className="stack gap-lg">
      <header className="page-header">
        <h1>Accueil site public</h1>
        <p>Galerie et meilleur postulant affichés sur la page d&apos;accueil du site vitrine.</p>
        <a className="btn-outline" href={websiteUrl} target="_blank" rel="noreferrer">
          Ouvrir le site
        </a>
      </header>

      {error && <p className="error" role="alert">{error}</p>}
      {success && <p className="form-ok" role="status">{success}</p>}

      <section className="panel stack gap-md">
        <header className="section-head">
          <div>
            <h2>Galerie</h2>
            <p className="muted">Photos indépendantes des événements. Seules les images publiées avec un fichier apparaissent sur le site.</p>
          </div>
          <button type="button" className="btn-primary" onClick={addGalleryItem} disabled={busy === 'gallery-add'}>
            Ajouter une photo
          </button>
        </header>

        {gallery.length === 0 ? (
          <div className="empty-state">
            <strong>Aucune photo</strong>
            <p>Ajoutez des images pour la section galerie de l&apos;accueil.</p>
          </div>
        ) : (
          <div className="event-gallery">
            {gallery.map((item) => (
              <figure key={item.id} className="site-gallery-card">
                {item.url ? (
                  <img src={item.url} alt={item.caption_fr || item.caption_en || 'Galerie'} />
                ) : (
                  <div className="site-gallery-empty">Photo manquante</div>
                )}
                <figcaption className="stack gap-sm">
                  <label>
                    Légende FR
                    <input
                      value={item.caption_fr}
                      onChange={(e) => patchGallery(item.id, { caption_fr: e.target.value })}
                    />
                  </label>
                  <label>
                    Légende EN
                    <input
                      value={item.caption_en}
                      onChange={(e) => patchGallery(item.id, { caption_en: e.target.value })}
                    />
                  </label>
                  <label className="check-row">
                    <input
                      type="checkbox"
                      checked={item.published}
                      onChange={(e) => patchGallery(item.id, { published: e.target.checked })}
                    />
                    Publiée
                  </label>
                  <label>
                    Ordre
                    <input
                      type="number"
                      min={0}
                      value={item.sort_order}
                      onChange={(e) => patchGallery(item.id, { sort_order: Number(e.target.value) || 0 })}
                    />
                  </label>
                  <div className="cta-row">
                    <label className="btn-outline file-btn">
                      {busy === `upload-${item.id}` ? 'Envoi…' : 'Choisir photo'}
                      <input
                        type="file"
                        accept="image/jpeg,image/png,image/webp"
                        hidden
                        onChange={(e) => uploadGalleryImage(item.id, e.target.files)}
                      />
                    </label>
                    <button
                      type="button"
                      className="btn-primary"
                      disabled={busy === `gallery-${item.id}`}
                      onClick={() => updateGalleryItem(item)}
                    >
                      Enregistrer
                    </button>
                    <button
                      type="button"
                      className="btn-outline danger"
                      disabled={busy === `delete-${item.id}`}
                      onClick={() => removeGalleryItem(item.id)}
                    >
                      Supprimer
                    </button>
                  </div>
                </figcaption>
              </figure>
            ))}
          </div>
        )}
      </section>

      <section className="panel stack gap-md">
        <header className="section-head">
          <div>
            <h2>Meilleur postulant</h2>
            <p className="muted">
              Un profil par mois. Le site affiche automatiquement le postulant du mois précédent
              (ex. en septembre, le profil d&apos;août).
            </p>
          </div>
          <button type="button" className="btn-outline" onClick={startNewPostulant}>
            Nouveau mois
          </button>
        </header>

        {postulants.length > 0 ? (
          <div className="postulant-month-list">
            {postulants.map((item) => (
              <button
                key={item.id}
                type="button"
                className={`postulant-month-chip${item.id === selectedId ? ' active' : ''}`}
                onClick={() => selectPostulant(item)}
              >
                {formatPostulantPeriod(item.period_year, item.period_month)}
                {item.published ? ' · publié' : ''}
              </button>
            ))}
          </div>
        ) : null}

        <form className="stack gap-md" onSubmit={savePostulant}>
          <div className="form-grid-2">
            <label>
              Année
              <input
                type="number"
                min={2000}
                max={2100}
                value={form.period_year}
                onChange={(e) => setForm((current) => ({ ...current, period_year: Number(e.target.value) || current.period_year }))}
                required
              />
            </label>
            <label>
              Mois
              <input
                type="number"
                min={1}
                max={12}
                value={form.period_month}
                onChange={(e) => setForm((current) => ({ ...current, period_month: Number(e.target.value) || current.period_month }))}
                required
              />
            </label>
          </div>

          <div className="stack gap-sm">
            <label>Flyer</label>
            {selectedPostulant?.flyer_url ? (
              <img src={postulantFlyerSrc(selectedPostulant)} alt="" className="postulant-flyer-preview" />
            ) : (
              <div className="postulant-flyer-empty">Aucun flyer — uploadez une affiche portrait ou paysage.</div>
            )}
            <label className={`btn-outline file-btn${selectedId ? '' : ' disabled'}`}>
              {busy === 'postulant-flyer' ? 'Envoi…' : selectedPostulant?.flyer_url ? 'Remplacer le flyer' : 'Uploader le flyer'}
              <input
                type="file"
                accept="image/jpeg,image/png,image/webp"
                hidden
                disabled={!selectedId}
                onChange={(e) => uploadPostulantFlyer(e.target.files)}
              />
            </label>
            {!selectedId ? (
              <p className="muted">Enregistrez d&apos;abord le profil du mois, puis uploadez le flyer.</p>
            ) : null}
          </div>

          <div className="stack gap-sm">
              <label className="check-row">
                <input
                  type="checkbox"
                  checked={form.published}
                  onChange={(e) => setForm((current) => ({ ...current, published: e.target.checked }))}
                />
                Publié sur le site
              </label>
              <div className="form-grid-2">
                <label>
                  Prénom
                  <input
                    value={form.first_name}
                    onChange={(e) => setForm((current) => ({ ...current, first_name: e.target.value }))}
                    required
                  />
                </label>
                <label>
                  Nom
                  <input
                    value={form.last_name}
                    onChange={(e) => setForm((current) => ({ ...current, last_name: e.target.value }))}
                    required
                  />
                </label>
              </div>
              <label>
                Club d&apos;origine
                <input
                  value={form.home_club}
                  onChange={(e) => setForm((current) => ({ ...current, home_club: e.target.value }))}
                />
              </label>
              <label>
                Citation FR
                <textarea
                  rows={3}
                  value={form.quote_fr}
                  onChange={(e) => setForm((current) => ({ ...current, quote_fr: e.target.value }))}
                />
              </label>
              <label>
                Citation EN
                <textarea
                  rows={3}
                  value={form.quote_en}
                  onChange={(e) => setForm((current) => ({ ...current, quote_en: e.target.value }))}
                />
              </label>
              <label>
                Nombre de visites
                <input
                  type="number"
                  min={0}
                  value={form.visit_count}
                  onChange={(e) => setForm((current) => ({ ...current, visit_count: Number(e.target.value) || 0 }))}
                />
              </label>
              <label>
                Clubs visités (un par ligne)
                <textarea
                  rows={4}
                  value={clubsText}
                  onChange={(e) => setClubsText(e.target.value)}
                  placeholder={'Rotaract Abidjan Sud\nRotaract Yamoussoukro'}
                />
              </label>
          </div>

          <div className="cta-row">
            <button type="submit" className="btn-primary" disabled={busy === 'postulant'}>
              {busy === 'postulant' ? 'Enregistrement…' : selectedId ? 'Enregistrer' : 'Créer le postulant'}
            </button>
            {selectedId ? (
              <button
                type="button"
                className="btn-outline danger"
                disabled={busy === 'postulant-delete'}
                onClick={() => void removePostulant()}
              >
                Supprimer
              </button>
            ) : null}
          </div>
        </form>
      </section>
    </div>
  )
}
