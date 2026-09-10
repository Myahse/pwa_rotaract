import { Link } from 'react-router-dom'
import type { PublicEvent } from '../api/types'
import type { Lang } from '../i18n'
import { loc } from '../lib/publicEvent'

function formatDate(iso: string, lang: Lang) {
  return new Intl.DateTimeFormat(lang === 'fr' ? 'fr-FR' : 'en-GB', {
    weekday: 'short',
    day: 'numeric',
    month: 'long',
    year: 'numeric',
  }).format(new Date(iso))
}

export function EventCard({
  event,
  lang,
  upcomingLabel,
  pastLabel,
}: {
  event: PublicEvent
  lang: Lang
  upcomingLabel: string
  pastLabel: string
}) {
  const upcoming = event.status === 'upcoming'
  return (
    <Link to={`/events/${event.id}`} className="event-card">
      {event.flyer_url ? (
        <img className="event-card-flyer" src={event.flyer_url} alt="" />
      ) : null}
      <header className="event-card-head">
        <p className="event-date">{formatDate(event.starts_at, lang)}</p>
        <span className={upcoming ? 'badge ok' : 'badge wait'}>{upcoming ? upcomingLabel : pastLabel}</span>
      </header>
      <h3>{loc(event, lang, 'title')}</h3>
      <p>{loc(event, lang, 'summary')}</p>
      <p className="event-meta">
        {[event.city, loc(event, lang, 'venue')].filter(Boolean).join(' · ')}
      </p>
    </Link>
  )
}
