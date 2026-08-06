const TOKEN_KEY = 'rotaract_token'

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

export function setToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token)
}

export function clearToken(): void {
  localStorage.removeItem(TOKEN_KEY)
}

const CLUB_KEY = 'rotaract_active_club'

export function getActiveClubId(): string | null {
  return localStorage.getItem(CLUB_KEY)
}

export function setActiveClubId(clubId: string): void {
  localStorage.setItem(CLUB_KEY, clubId)
}
