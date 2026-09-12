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

  const res = await fetch(`${API}${path}`, { ...options, headers })
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
  const res = await fetch(`${API}${path}`, { method: 'POST', body, headers })
  if (!res.ok) throw new ApiClientError(res.status, await parseError(res))
  return (await res.json()) as T
}
