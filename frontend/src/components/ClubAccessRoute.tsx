import { Navigate } from 'react-router-dom'
import type { ReactNode } from 'react'
import { useAuth } from '../auth/AuthContext'

type Section = 'club' | 'manage' | 'cotisations'

export function ClubAccessRoute({ section, children }: { section: Section; children: ReactNode }) {
  const { uiCaps, activeClubId } = useAuth()

  const allowed =
    section === 'club' ? uiCaps.nav.club
    : section === 'manage' ? uiCaps.nav.manage
    : uiCaps.nav.cotisations

  if (!allowed) {
    return <Navigate to={uiCaps.defaultPath || (activeClubId ? `/clubs/${activeClubId}/cotisations` : '/home')} replace />
  }

  return children
}
