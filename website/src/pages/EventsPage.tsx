import { useEffect, useState } from 'react'
import { EventCard } from '../components/EventCard'
import { RevealSection } from '../components/RevealSection'
import { apiRequest } from '../api/client'
import type { PublicEvent } from '../api/types'
import { pastEvents, upcomingEvents } from '../lib/publicEvent'
import { useI18n } from '../i18n'

export function EventsPage() {
  const { t, lang } = useI18n()
  const [items, setItems] = useState<PublicEvent[] | null>(null)
  const [error, setError] = useState('')

  useEffect(() => {
    apiRequest<PublicEvent[]>('/events')
      .then((data) => setItems(data ?? []))
      .catch((err) => setError(err instanceof Error ? err.message : t.eventsLoadError))
  }, [t.eventsLoadError])

  const upcoming = upcomingEvents(items ?? [])
  const past = pastEvents(items ?? [])
  const featured = upcoming[0]
  const moreUpcoming = upcoming.slice(1)

  return (
    <div className="screen-scroll">
      <RevealSection className="screen-section screen-section--events" immediate>
        <div className="screen-section-inner screen-section-inner--wide screen-section-inner--fill screen-section-inner--stack">
          <p className="screen-eyebrow">{t.eventsKicker}</p>
          <h1>{t.eventsTitle}</h1>
          <p className="lede">{t.eventsLede}</p>
          {error ? <p className="form-error" role="alert">{error}</p> : null}
          {items === null && !error ? <p className="muted">{t.eventsLoading}</p> : null}

          {featured ? (
            <>
              <h2>{t.eventsFeatured}</h2>
              <div className="screen-section--fill-grid">
                <EventCard
                  event={featured}
                  lang={lang}
                  upcomingLabel={t.eventUpcoming}
                  pastLabel={t.eventPast}
                />
              </div>
            </>
          ) : items && !error ? (
            <>
              <h2>{t.eventsUpcoming}</h2>
              <p className="home-empty">{t.eventsEmpty}</p>
            </>
          ) : null}
        </div>
      </RevealSection>

      {moreUpcoming.length > 0 ? (
        <RevealSection className="screen-section screen-section--alt screen-section--events">
          <div className="screen-section-inner screen-section-inner--wide screen-section-inner--fill screen-section-inner--stack">
            <h2>{t.eventsUpcoming}</h2>
            <div className="event-grid screen-section--fill-grid">
              {moreUpcoming.map((event) => (
                <EventCard
                  key={event.id}
                  event={event}
                  lang={lang}
                  upcomingLabel={t.eventUpcoming}
                  pastLabel={t.eventPast}
                />
              ))}
            </div>
          </div>
        </RevealSection>
      ) : null}

      {past.length > 0 ? (
        <RevealSection className="screen-section screen-section--events">
          <div className="screen-section-inner screen-section-inner--wide screen-section-inner--fill screen-section-inner--stack">
            <h2>{t.eventsPast}</h2>
            <div className="event-grid screen-section--fill-grid">
              {past.map((event) => (
                <EventCard
                  key={event.id}
                  event={event}
                  lang={lang}
                  upcomingLabel={t.eventUpcoming}
                  pastLabel={t.eventPast}
                />
              ))}
            </div>
          </div>
        </RevealSection>
      ) : null}
    </div>
  )
}
