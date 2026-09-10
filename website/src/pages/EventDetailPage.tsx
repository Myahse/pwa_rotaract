import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { apiRequest, ApiClientError } from '../api/client'
import type { PublicEvent } from '../api/types'
import { loc } from '../lib/publicEvent'
import { useI18n } from '../i18n'

function formatDateTime(iso: string, lang: 'fr' | 'en') {
  return new Intl.DateTimeFormat(lang === 'fr' ? 'fr-FR' : 'en-GB', {
    weekday: 'long',
    day: 'numeric',
    month: 'long',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(new Date(iso))
}

export function EventDetailPage() {
  const { eventId } = useParams()
  const { t, lang } = useI18n()
  const [event, setEvent] = useState<PublicEvent | null>(null)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!eventId) return
    setEvent(null)
    setError('')
    apiRequest<PublicEvent>(`/events/${eventId}`)
      .then(setEvent)
      .catch((err) => {
        if (err instanceof ApiClientError && err.status === 404) {
          setError(t.eventNotFound)
          return
        }
        setError(err instanceof Error ? err.message : t.eventsLoadError)
      })
  }, [eventId, t.eventNotFound, t.eventsLoadError])

  if (error) {
    return (
      <section className="vitrine-hero">
        <p className="eyebrow">{t.eventsKicker}</p>
        <h1>{t.eventsTitle}</h1>
        <p className="lede">{error}</p>
        <Link to="/events" className="btn-outline">{t.eventBack}</Link>
      </section>
    )
  }

  if (!event) {
    return <p className="muted">{t.eventsLoading}</p>
  }

  const title = loc(event, lang, 'title')
  const body = loc(event, lang, 'body')
  const summary = loc(event, lang, 'summary')
  const venue = loc(event, lang, 'venue')
  const cta = loc(event, lang, 'cta_label')
  const upcoming = event.status === 'upcoming'

  return (
    <article className="event-detail">
      <p>
        <Link to="/events" className="back-link">{t.eventBack}</Link>
      </p>
      <header className="vitrine-hero event-detail-hero">
        {event.flyer_url ? (
          <img className="event-detail-flyer" src={event.flyer_url} alt={title} />
        ) : null}
        <p className="eyebrow">{t.eventsKicker}</p>
        <p className="event-date">{formatDateTime(event.starts_at, lang)}</p>
        <h1>{title}</h1>
        <p className="lede">{summary}</p>
        <p className="event-meta">
          {[event.city, venue].filter(Boolean).join(' · ')}
        </p>
        <span className={upcoming ? 'badge ok' : 'badge wait'}>
          {upcoming ? t.eventUpcoming : t.eventPast}
        </span>
        {cta && event.cta_url ? (
          <div className="cta-row center">
            {event.cta_url.startsWith('/') ? (
              <Link className="btn-primary" to={event.cta_url}>{cta}</Link>
            ) : (
              <a className="btn-primary" href={event.cta_url} target="_blank" rel="noreferrer">
                {cta}
              </a>
            )}
          </div>
        ) : null}
      </header>

      {body ? (
        <section className="section">
          <div className="event-body">{body}</div>
        </section>
      ) : null}

      {event.images?.length ? (
        <section className="section" style={{ borderBottom: 0 }}>
          <h2>{t.eventGallery}</h2>
          <div className="event-gallery">
            {event.images.map((image) => (
              <img key={image.id} src={image.url} alt="" />
            ))}
          </div>
        </section>
      ) : null}
    </article>
  )
}
