export type SmoothScrollMetrics = {
  scrollY: number
  maxScrollY: number
  progress: number
}

export type SmoothScrollApi = {
  scrollTo: (y: number) => void
  setTarget: (y: number) => void
  getMetrics: () => SmoothScrollMetrics
  subscribe: (listener: () => void) => () => void
}

let activeApi: SmoothScrollApi | null = null

export function setSmoothScrollApi(api: SmoothScrollApi | null) {
  activeApi = api
}

export function getSmoothScrollApi(): SmoothScrollApi | null {
  return activeApi
}

export function readScrollMetrics(): SmoothScrollMetrics {
  const maxScrollY = Math.max(0, document.documentElement.scrollHeight - window.innerHeight)
  const scrollY = window.scrollY
  return {
    scrollY,
    maxScrollY,
    progress: maxScrollY > 0 ? scrollY / maxScrollY : 0,
  }
}

function getHeaderOffset(): number {
  const raw = getComputedStyle(document.documentElement).getPropertyValue('--header-h').trim()
  const parsed = parseFloat(raw)
  return Number.isFinite(parsed) ? parsed : 76
}

export function scrollToSectionById(id: string) {
  const el = document.getElementById(id)
  if (!el) return

  const y = Math.max(0, el.getBoundingClientRect().top + window.scrollY - getHeaderOffset())
  const api = getSmoothScrollApi()
  if (api) api.setTarget(y)
  else window.scrollTo(0, y)
}

export function scrollToTop(instant = false) {
  const api = getSmoothScrollApi()
  if (api) {
    if (instant) api.scrollTo(0)
    else api.setTarget(0)
  } else {
    window.scrollTo(0, 0)
  }
}
