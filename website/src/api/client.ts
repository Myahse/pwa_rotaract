import type { ApiError } from './types'

const PRODUCTION_API = 'https://rotaract-civ-api.fly.dev/api/v1'

function normalizeApiBase(raw: string): string {
  const base = raw.trim().replace(/\/$/, '')
  if (!base) return PRODUCTION_API
  if (base.endsWith('/api/v1')) return base
  if (base.endsWith('/api')) return `${base}/v1`
  return `${base}/api/v1`
}

function isStaleRenderApi(base: string): boolean {
  return /onrender\.com/i.test(base)
}

function resolveApiBase(): string {
  const fromEnv = import.meta.env.VITE_API_BASE as string | undefined
  if (fromEnv?.trim()) {
    const normalized = normalizeApiBase(fromEnv)
    if (import.meta.env.PROD && isStaleRenderApi(normalized)) return '/api/v1'
    return normalized
  }
  if (import.meta.env.PROD) return '/api/v1'
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

async function parseError(res: Response): Promise<string> {
  try {
    const body = (await res.json()) as ApiError
    return body.message || body.error || res.statusText
  } catch {
    return res.statusText
  }
}

function networkErrorMessage(): string {
  return import.meta.env.PROD
    ? 'Le serveur API est indisponible (maintenance ou hébergement suspendu). Réessayez dans quelques minutes.'
    : `Impossible de joindre l’API (${API}). Démarrez le backend local ou vérifiez VITE_API_BASE.`
}

export async function apiRequest<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const headers = new Headers(options.headers)
  if (options.body && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }

  let res: Response
  try {
    res = await fetch(`${API}${path}`, { ...options, headers })
  } catch {
    throw new ApiClientError(0, networkErrorMessage())
  }
  if (!res.ok) {
    throw new ApiClientError(res.status, await parseError(res))
  }
  if (res.status === 204) {
    return undefined as T
  }
  return (await res.json()) as T
}

export async function apiUpload<T>(path: string, body: FormData): Promise<T> {
  let res: Response
  try {
    res = await fetch(`${API}${path}`, { method: 'POST', body })
  } catch {
    throw new ApiClientError(0, networkErrorMessage())
  }
  if (!res.ok) {
    throw new ApiClientError(res.status, await parseError(res))
  }
  return (await res.json()) as T
}
