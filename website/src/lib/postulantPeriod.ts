export function isValidPostulantPeriod(year: number, month: number) {
  return Number.isFinite(year) && year >= 2000 && year <= 2100 && Number.isFinite(month) && month >= 1 && month <= 12
}

export function resolvePostulantPeriod(year: number, month: number) {
  if (isValidPostulantPeriod(year, month)) {
    return { year, month }
  }
  return previousCalendarMonth()
}

export function formatPostulantPeriod(year: number, month: number, lang: 'fr' | 'en') {
  const period = resolvePostulantPeriod(year, month)
  const date = new Date(period.year, period.month - 1, 1)
  if (Number.isNaN(date.getTime())) return ''

  return new Intl.DateTimeFormat(lang === 'fr' ? 'fr-FR' : 'en-US', {
    month: 'long',
    year: 'numeric',
  }).format(date)
}

export function previousCalendarMonth(now = new Date()) {
  const date = new Date(now.getFullYear(), now.getMonth() - 1, 1)
  return { year: date.getFullYear(), month: date.getMonth() + 1 }
}
