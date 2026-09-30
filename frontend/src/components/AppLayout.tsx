import type { CSSProperties } from 'react'
import { Link, Outlet, useLocation } from 'react-router-dom'
import { Building2, Coins, House, LogOut, MessageSquare, Settings2, UserCircle } from 'lucide-react'
import { useAuth } from '../auth/AuthContext'

export function AppLayout() {
  const { user, clubs, activeClubId, logout, selectClub, uiCaps } = useAuth()
  const location = useLocation()
  const inChat = location.pathname.startsWith('/chat/')
  const inMessages = /\/messages$/.test(location.pathname)
  const activeClub = clubs.find((c) => c.club_id === activeClubId)
  const clubLabel = uiCaps.profile === 'commission_president' ? 'Commission' : 'Club'
  const clubLabelShort = uiCaps.profile === 'commission_president' ? 'Comm.' : 'Club'

  const navItems: {
    to: string
    label: string
    short: string
    icon: typeof House
    active: boolean
  }[] = []

  if (uiCaps.nav.home) {
    navItems.push({
      to: '/home',
      label: 'Accueil',
      short: 'Accueil',
      icon: House,
      active: location.pathname === '/home',
    })
  }
  if (activeClubId && uiCaps.nav.club) {
    navItems.push({
      to: `/clubs/${activeClubId}`,
      label: clubLabel,
      short: clubLabelShort,
      icon: Building2,
      active:
        location.pathname.startsWith(`/clubs/${activeClubId}`)
        && !location.pathname.includes('/manage')
        && !location.pathname.includes('/cotisations')
        && !location.pathname.includes('/messages'),
    })
  }
  if (activeClubId && uiCaps.nav.messages) {
    navItems.push({
      to: `/clubs/${activeClubId}/messages`,
      label: 'Messages',
      short: 'Messages',
      icon: MessageSquare,
      active: location.pathname.includes('/messages') || location.pathname.startsWith('/chat/'),
    })
  }
  if (activeClubId && uiCaps.nav.cotisations) {
    navItems.push({
      to: `/clubs/${activeClubId}/cotisations`,
      label: 'Cotisations',
      short: 'Cotis.',
      icon: Coins,
      active: location.pathname.includes('/cotisations'),
    })
  }
  if (activeClubId && uiCaps.nav.manage) {
    navItems.push({
      to: `/clubs/${activeClubId}/manage`,
      label: 'Gérer le club',
      short: 'Gérer',
      icon: Settings2,
      active: location.pathname.includes('/manage'),
    })
  }
  if (uiCaps.nav.profile) {
    navItems.push({
      to: '/profile',
      label: 'Profil',
      short: 'Profil',
      icon: UserCircle,
      active: location.pathname === '/profile',
    })
  }

  return (
    <div className="app-shell">
      <header className="app-header">
        <div className="brand">
          <picture>
            <source srcSet="/logo-white.png" media="(prefers-color-scheme: dark)" />
            <img className="brand-logo" src="/logo.png" alt="Rotaract IUGB Club" />
          </picture>
          {activeClub?.club && clubs.length <= 1 && (
            <p className="header-club-name">{activeClub.club.name}</p>
          )}
        </div>
        <div className="header-actions">
          {clubs.length > 1 && (
            <select
              value={activeClubId ?? ''}
              onChange={(e) => selectClub(e.target.value)}
              className="club-select"
              aria-label="Choisir le club"
            >
              {clubs.map((m) => (
                <option key={m.club_id} value={m.club_id}>
                  {m.club?.name ?? m.club_id}
                </option>
              ))}
            </select>
          )}
          <span className="user-chip">
            {user?.first_name}
            {uiCaps.profile !== 'member' || uiCaps.profileLabel !== 'Membre' ? ` · ${uiCaps.profileLabel}` : ''}
          </span>
          <button type="button" className="btn-ghost header-logout" onClick={logout}>
            <LogOut className="header-logout-icon" aria-hidden="true" />
            <span className="header-logout-text">Déconnexion</span>
          </button>
        </div>
      </header>

      <main
        className={[
          'app-main',
          inChat ? 'app-main--chat' : '',
          inMessages ? 'app-main--messages' : '',
        ].filter(Boolean).join(' ')}
      >
        <Outlet />
      </main>

      {navItems.length > 0 && !inChat && (
        <nav
          className="bottom-nav"
          aria-label="Navigation principale"
          style={{ '--nav-count': navItems.length } as CSSProperties}
        >
          {navItems.map((item) => {
            const Icon = item.icon
            return (
              <Link
                key={item.to}
                to={item.to}
                className={item.active ? 'active' : ''}
                aria-current={item.active ? 'page' : undefined}
                aria-label={item.label}
              >
                <Icon className="nav-icon" aria-hidden="true" />
                <span className="nav-label">{item.short}</span>
              </Link>
            )
          })}
        </nav>
      )}
    </div>
  )
}
