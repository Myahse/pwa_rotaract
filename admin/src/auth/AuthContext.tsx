import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react'
import { apiRequest } from '../api/client'
import type { LoginResponse, User } from '../api/types'
import { clearToken, getToken, setToken } from './storage'

type AuthState = {
  user: User | null
  loading: boolean
  login: (email: string, password: string) => Promise<void>
  logout: () => void
  token: string | null
}

const AuthContext = createContext<AuthState | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [token, setTokenState] = useState<string | null>(getToken())
  const [loading, setLoading] = useState(true)

  const loadSession = useCallback(async (accessToken: string) => {
    const profile = await apiRequest<User>('/auth/me', {}, accessToken)
    if (!profile.is_admin) {
      throw new Error('Ce compte n\'est pas administrateur')
    }
    setUser(profile)
  }, [])

  useEffect(() => {
    const existing = getToken()
    if (!existing) {
      setLoading(false)
      return
    }
    loadSession(existing)
      .catch(() => {
        clearToken()
        setTokenState(null)
      })
      .finally(() => setLoading(false))
  }, [loadSession])

  const login = useCallback(async (email: string, password: string) => {
    const res = await apiRequest<LoginResponse>('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    })
    setToken(res.access_token)
    setTokenState(res.access_token)
    await loadSession(res.access_token)
  }, [loadSession])

  const logout = useCallback(() => {
    clearToken()
    setTokenState(null)
    setUser(null)
  }, [])

  const value = useMemo(
    () => ({ user, loading, login, logout, token }),
    [user, loading, login, logout, token],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth(): AuthState {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}
