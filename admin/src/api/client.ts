import type { ApiError } from './types'

const PRODUCTION_API = 'https://pwa-rotaract.onrender.com/api/v1'

function normalizeApiBase(raw: string): string {
  const base = raw.trim().replace(/\/$/, '')
  if (!base) return PRODUCTION_API
  if (base.endsWith('/api/v1')) return base
  if (base.endsWith('/api')) return `${base}/v1`
  return `${base}/api/v1`
}

function resolveApiBase(): string {
  // Production builds always call Render directly (Vercel VITE_API_BASE overrides caused 404s).
  if (import.meta.env.PROD) return PRODUCTION_API
  const fromEnv = import.meta.env.VITE_API_BASE as string | undefined
  if (fromEnv?.trim()) return normalizeApiBase(fromEnv)
  return '/api/v1'
}

const API = resolveApiBase()

export class ApiClientError extends Error {
  status: number

  constructor(status: number, message: string) {
    super(message)
    this.status = status
  }
}

function isHtmlBody(text: string): boolean {
  const trimmed = text.trimStart().toLowerCase()
  return trimmed.startsWith('<!doctype') || trimmed.startsWith('<html')
}

async function readBody(res: Response): Promise<string> {
  return res.text()
}

async function parseJsonBody<T>(res: Response, body: string): Promise<T> {
  if (isHtmlBody(body)) {
    throw new ApiClientError(
      res.status,
      'Réponse HTML inattendue : l’API n’est pas joignable sur cette URL. Utilisez le déploiement Vercel (rewrites /api) ou définissez VITE_API_BASE vers le backend Render.',
    )
  }
  if (!body.trim()) {
    return undefined as T
  }
  try {
    return JSON.parse(body) as T
  } catch {
    throw new ApiClientError(res.status, body.slice(0, 200) || 'Réponse API invalide')
  }
}

async function parseError(res: Response, body: string): Promise<string> {
  if (isHtmlBody(body)) {
    return 'Réponse HTML inattendue — vérifiez le proxy API (Vercel /api ou VITE_API_BASE).'
  }
  if (body.trim() === '404 page not found') {
    return `Route API introuvable (${res.url}). Redéployez l’admin ou contactez le support.`
  }
  try {
    const parsed = JSON.parse(body) as ApiError
    return parsed.message || parsed.error || res.statusText
  } catch {
    return body.slice(0, 200) || res.statusText
  }
}

export async function apiRequest<T>(
  path: string,
  options: RequestInit = {},
  token?: string | null,
): Promise<T> {
  const headers = new Headers(options.headers)
  if (options.body && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }
  if (token) {
    headers.set('Authorization', `Bearer ${token}`)
  }

  const res = await fetch(`${API}${path}`, { ...options, headers })
  const body = await readBody(res)
  if (!res.ok) {
    throw new ApiClientError(res.status, await parseError(res, body))
  }
  if (res.status === 204) {
    return undefined as T
  }
  return parseJsonBody<T>(res, body)
}

export async function apiUpload<T>(
  path: string,
  body: FormData,
  token?: string | null,
): Promise<T> {
  const headers = new Headers()
  if (token) {
    headers.set('Authorization', `Bearer ${token}`)
  }

  const res = await fetch(`${API}${path}`, { method: 'POST', body, headers })
  const responseBody = await readBody(res)
  if (!res.ok) {
    throw new ApiClientError(res.status, await parseError(res, responseBody))
  }
  return parseJsonBody<T>(res, responseBody)
}
