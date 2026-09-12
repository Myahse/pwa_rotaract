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
import type { ClubAccessSummary, ClubMembership, LoginResponse, User } from '../api/types'
import { resolveClubUiCaps, type ClubUiCaps } from './clubUiAccess'
import { clearToken, getActiveClubId, getToken, setActiveClubId, setToken } from './storage'

type AuthState = {
  user: User | null
  clubs: ClubMembership[]
  activeClubId: string | null
  clubAccess: ClubAccessSummary | null
  loading: boolean
  login: (email: string, password: string) => Promise<void>
  logout: () => void
  refresh: () => Promise<void>
  selectClub: (clubId: string) => void
  token: string | null
  isHead: boolean
  hasPermission: (key: string) => boolean
  uiCaps: ClubUiCaps
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
  const [clubAccess, setClubAccess] = useState<ClubAccessSummary | null>(null)
  const [token, setTokenState] = useState<string | null>(getToken())
  const [loading, setLoading] = useState(true)

  const loadClubAccess = useCallback(async (accessToken: string, clubId: string | null) => {
    if (!clubId) {
      setClubAccess(null)
      return
    }
    try {
      const access = await apiRequest<ClubAccessSummary>(`/clubs/${clubId}/me/access`, {}, accessToken)
      setClubAccess({ ...access, commissions: access.commissions ?? [] })
    } catch {
      setClubAccess(null)
    }
  }, [])

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
    await loadClubAccess(accessToken, nextClub)
  }, [loadClubAccess])

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
    setClubAccess(null)
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
    const existing = getToken()
    if (existing) {
      loadClubAccess(existing, clubId).catch(() => setClubAccess(null))
    }
  }, [loadClubAccess])

  const isHead = clubAccess?.member_role === 'head'
  const hasPermission = useCallback((key: string) => {
    if (!clubAccess) return false
    if (clubAccess.member_role === 'head') return true
    return clubAccess.permissions.includes(key)
  }, [clubAccess])
  const uiCaps = useMemo(
    () => resolveClubUiCaps(clubAccess, activeClubId),
    [clubAccess, activeClubId],
  )

  const value = useMemo(
    () => ({
      user,
      clubs,
      activeClubId,
      clubAccess,
      loading,
      login,
      logout,
      refresh,
      selectClub,
      token,
      isHead,
      hasPermission,
      uiCaps,
    }),
    [user, clubs, activeClubId, clubAccess, loading, login, logout, refresh, selectClub, token, isHead, hasPermission, uiCaps],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth(): AuthState {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}
