export function isValidPostulantPeriod(year: number, month: number) {
  return Number.isFinite(year) && year >= 2000 && year <= 2100 && Number.isFinite(month) && month >= 1 && month <= 12
}

export function formatPostulantPeriod(year: number, month: number) {
  if (!isValidPostulantPeriod(year, month)) return 'Mois non défini'

  const date = new Date(year, month - 1, 1)
  if (Number.isNaN(date.getTime())) return 'Mois non défini'

  const label = new Intl.DateTimeFormat('fr-FR', {
    month: 'long',
    year: 'numeric',
  }).format(date)
  return label.charAt(0).toUpperCase() + label.slice(1)
}

export function previousCalendarMonth(now = new Date()) {
  const date = new Date(now.getFullYear(), now.getMonth() - 1, 1)
  return { year: date.getFullYear(), month: date.getMonth() + 1 }
}
