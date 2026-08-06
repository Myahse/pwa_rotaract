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
import { clearToken, getActiveClubId, getToken, setActiveClubId, setToken } from './storage'

type AuthState = {
  user: User | null
  clubs: ClubMembership[]
  activeClubId: string | null
  loading: boolean
  login: (email: string, password: string) => Promise<void>
  logout: () => void
  refresh: () => Promise<void>
  selectClub: (clubId: string) => void
  token: string | null
}

async function applyToken(
  accessToken: string,
  loadSession: (token: string) => Promise<void>,
  setTokenState: (token: string | null) => void,
) {
  setToken(accessToken)
  setTokenState(accessToken)
  await loadSession(accessToken)
}

const AuthContext = createContext<AuthState | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [clubs, setClubs] = useState<ClubMembership[]>([])
  const [activeClubId, setActiveClubIdState] = useState<string | null>(getActiveClubId())
  const [token, setTokenState] = useState<string | null>(getToken())
  const [loading, setLoading] = useState(true)

  const loadSession = useCallback(async (accessToken: string) => {
    const [profile, memberships] = await Promise.all([
      apiRequest<User>('/auth/me', {}, accessToken),
      apiRequest<ClubMembership[]>('/users/me/clubs', {}, accessToken),
    ])
    setUser(profile)
    setClubs(memberships)
    const stored = getActiveClubId()
    const valid = memberships.find((m) => m.club_id === stored)
    const nextClub = valid?.club_id ?? memberships[0]?.club_id ?? null
    setActiveClubIdState(nextClub)
    if (nextClub) setActiveClubId(nextClub)
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
    await applyToken(res.access_token, loadSession, setTokenState)
  }, [loadSession])

  const logout = useCallback(() => {
    clearToken()
    setTokenState(null)
    setUser(null)
    setClubs([])
    setActiveClubIdState(null)
  }, [])

  const refresh = useCallback(async () => {
    const existing = getToken()
    if (!existing) return
    setTokenState(existing)
    await loadSession(existing)
  }, [loadSession])

  const selectClub = useCallback((clubId: string) => {
    setActiveClubId(clubId)
    setActiveClubIdState(clubId)
  }, [])

  const value = useMemo(
    () => ({ user, clubs, activeClubId, loading, login, logout, refresh, selectClub, token }),
    [user, clubs, activeClubId, loading, login, logout, refresh, selectClub, token],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth(): AuthState {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}
