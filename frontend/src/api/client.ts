import type { ApiError } from './types'

const PRODUCTION_API = 'https://rotaract-civ-api.fly.dev/api/v1'

function normalizeApiBase(raw: string): string {
  const base = raw.trim().replace(/\/$/, '')
  if (!base) return PRODUCTION_API
  if (base.endsWith('/api/v1')) return base
  if (base.endsWith('/api')) return `${base}/v1`
  return `${base}/api/v1`
}

function resolveApiBase(): string {
  const fromEnv = import.meta.env.VITE_API_BASE as string | undefined
  if (fromEnv?.trim()) return normalizeApiBase(fromEnv)
  if (import.meta.env.PROD) return PRODUCTION_API
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

  let res: Response
  try {
    res = await fetch(`${API}${path}`, { ...options, headers })
  } catch {
    const message = import.meta.env.PROD
      ? 'Serveur API indisponible. Réessayez dans quelques minutes.'
      : `Impossible de joindre l’API (${API}).`
    throw new ApiClientError(0, message)
  }
  if (!res.ok) {
    throw new ApiClientError(res.status, await parseError(res))
  }
  if (res.status === 204) {
    return undefined as T
  }
  return (await res.json()) as T
}

export async function apiUpload<T>(
  path: string,
  body: FormData,
  token?: string | null,
): Promise<T> {
  const headers = new Headers()
  if (token) headers.set('Authorization', `Bearer ${token}`)
  let res: Response
  try {
    res = await fetch(`${API}${path}`, { method: 'POST', body, headers })
  } catch {
    throw new ApiClientError(0, 'Serveur API indisponible. Réessayez dans quelques minutes.')
  }
  if (!res.ok) throw new ApiClientError(res.status, await parseError(res))
  return (await res.json()) as T
}

/** WebSocket base when API is on Fly (not the Vercel club domain). */
export function chatWebSocketURL(groupId: string, token: string): string {
  const path = `/api/v1/ws/chat?token=${encodeURIComponent(token)}&group_id=${encodeURIComponent(groupId)}`
  if (API.startsWith('http://') || API.startsWith('https://')) {
    const origin = new URL(API).origin
    const wsOrigin = origin.replace(/^http/, 'ws')
    return `${wsOrigin}${path}`
  }
  const wsProto = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  return `${wsProto}//${window.location.host}${path}`
}
