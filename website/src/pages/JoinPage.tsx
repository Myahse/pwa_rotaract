import { useState, type FormEvent } from 'react'
import { Link } from 'react-router-dom'
import { apiRequest, ApiClientError } from '../api/client'
import type { AccessRequest } from '../api/types'
import { BrandLogo } from '../components/BrandLogo'
import { CLUB_NAME } from '../lib/club'
import { useI18n } from '../i18n'

const CLUB_APP = import.meta.env.VITE_CLUB_APP_URL ?? 'http://localhost:5173'

export function JoinPage() {
  const { t } = useI18n()
  const [form, setForm] = useState({
    first_name: '',
    last_name: '',
    email: '',
    password: '',
    profession: '',
  })
  const [error, setError] = useState('')
  const [done, setDone] = useState<AccessRequest | null>(null)
  const [loading, setLoading] = useState(false)

  function mapError(err: unknown) {
    if (err instanceof ApiClientError) {
      if (err.message.includes('already a member')) return t.joinAlreadyMember
      if (err.message.includes('pending access request')) return t.joinPending
      if (err.message.includes('password is required')) return t.joinPasswordRequired
    }
    return err instanceof Error ? err.message : t.joinError
  }

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    if (form.password.length > 0 && form.password.length < 8) {
      setError(t.joinPasswordRequired)
      return
    }
    setLoading(true)
    setError('')
    try {
      const req = await apiRequest<AccessRequest>('/access-requests', {
        method: 'POST',
        body: JSON.stringify({
          first_name: form.first_name,
          last_name: form.last_name,
          email: form.email,
          password: form.password || undefined,
          club_name: CLUB_NAME,
          profession: form.profession || undefined,
        }),
      })
      setDone(req)
    } catch (err) {
      setError(mapError(err))
    } finally {
      setLoading(false)
    }
  }

  if (done) {
    return (
      <>
        <section className="vitrine-hero">
          <BrandLogo hero />
          <p className="eyebrow">{t.joinKicker}</p>
          <h1>{t.joinSentTitle}</h1>
          <p className="lede">
            {done.known_member ? t.joinSentKnown : t.joinSentBody}
          </p>
          <div className="cta-row center">
            <a className="btn-primary" href={`${CLUB_APP}/`}>
              {t.joinLoginCta}
            </a>
            <Link to="/" className="btn-outline">
              {t.navHome}
            </Link>
          </div>
        </section>
      </>
    )
  }

  return (
    <>
      <section className="vitrine-hero">
        <BrandLogo hero />
        <p className="eyebrow">{t.joinKicker}</p>
        <h1>{t.joinTitle}</h1>
        <p className="lede">{t.joinLede}</p>
      </section>

      <section className="section" style={{ borderBottom: 0 }}>
        <form className="donate-form" onSubmit={onSubmit}>
          <label>
            {t.joinFirstName}
            <input
              value={form.first_name}
              onChange={(e) => setForm({ ...form, first_name: e.target.value })}
              autoComplete="given-name"
              required
            />
          </label>
          <label>
            {t.joinLastName}
            <input
              value={form.last_name}
              onChange={(e) => setForm({ ...form, last_name: e.target.value })}
              autoComplete="family-name"
              required
            />
          </label>
          <label>
            {t.joinEmail}
            <input
              type="email"
              value={form.email}
              onChange={(e) => setForm({ ...form, email: e.target.value })}
              autoComplete="email"
              required
            />
          </label>
          <p className="field-hint">
            <strong>{t.joinClub}</strong> {t.joinClubFixed}
          </p>
          <label>
            {t.joinProfession}
            <input
              value={form.profession}
              onChange={(e) => setForm({ ...form, profession: e.target.value })}
            />
          </label>
          <label>
            {t.joinPassword}
            <input
              type="password"
              value={form.password}
              onChange={(e) => setForm({ ...form, password: e.target.value })}
              autoComplete="new-password"
            />
          </label>
          <p className="field-hint">{t.joinPasswordHint}</p>
          {error ? <p className="form-error">{error}</p> : null}
          <button type="submit" className="btn-primary btn-block" disabled={loading}>
            {loading ? t.joinSending : t.joinSubmit}
          </button>
        </form>
      </section>
    </>
  )
}
