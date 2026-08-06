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
import type { ClubMembership, LoginResponse, User } from '../api/types'
import { clearToken, getToken, setToken } from './storage'

type AuthState = {
  user: User | null
  clubs: ClubMembership[]
  loading: boolean
  login: (email: string, password: string) => Promise<void>
  logout: () => void
  token: string | null
  loginOpen: boolean
  openLogin: (reason?: string) => void
  closeLogin: () => void
  loginReason: string
  requireAuth: (action: () => void | Promise<void>, reason?: string) => void
}

const AuthContext = createContext<AuthState | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [clubs, setClubs] = useState<ClubMembership[]>([])
  const [token, setTokenState] = useState<string | null>(getToken())
  const [loading, setLoading] = useState(true)
  const [loginOpen, setLoginOpen] = useState(false)
  const [loginReason, setLoginReason] = useState('')
  const [pendingAction, setPendingAction] = useState<(() => void | Promise<void>) | null>(null)

  const loadSession = useCallback(async (accessToken: string) => {
    const [profile, memberships] = await Promise.all([
      apiRequest<User>('/auth/me', {}, accessToken),
      apiRequest<ClubMembership[]>('/users/me/clubs', {}, accessToken),
    ])
    setUser(profile)
    setClubs(memberships)
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

  const openLogin = useCallback((reason = 'Connectez-vous pour continuer') => {
    setLoginReason(reason)
    setLoginOpen(true)
  }, [])

  const closeLogin = useCallback(() => {
    setLoginOpen(false)
    setPendingAction(null)
  }, [])

  const login = useCallback(async (email: string, password: string) => {
    const res = await apiRequest<LoginResponse>('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    })
    setToken(res.access_token)
    setTokenState(res.access_token)
    await loadSession(res.access_token)
    setLoginOpen(false)
    const action = pendingAction
    setPendingAction(null)
    if (action) await action()
  }, [loadSession, pendingAction])

  const logout = useCallback(() => {
    clearToken()
    setTokenState(null)
    setUser(null)
    setClubs([])
  }, [])

  const requireAuth = useCallback((action: () => void | Promise<void>, reason?: string) => {
    if (getToken() && user) {
      void action()
      return
    }
    setPendingAction(() => action)
    openLogin(reason)
  }, [openLogin, user])

  const value = useMemo(
    () => ({
      user, clubs, loading, login, logout, token,
      loginOpen, openLogin, closeLogin, loginReason, requireAuth,
    }),
    [user, clubs, loading, login, logout, token, loginOpen, openLogin, closeLogin, loginReason, requireAuth],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth(): AuthState {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}
