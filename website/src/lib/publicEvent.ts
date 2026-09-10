import type { Lang } from '../i18n'
import type { PublicEvent } from '../api/types'

export function loc(
  event: PublicEvent,
  lang: Lang,
  field: 'title' | 'summary' | 'venue' | 'body' | 'cta_label',
) {
  const fr = event[`${field}_fr`]
  const en = event[`${field}_en`]
  if (lang === 'en') return en || fr
  return fr || en
}

export function upcomingEvents(items: PublicEvent[]) {
  return items
    .filter((event) => event.status === 'upcoming')
    .sort((a, b) => a.starts_at.localeCompare(b.starts_at))
}

export function pastEvents(items: PublicEvent[]) {
  return items.filter((event) => event.status === 'past')
}
