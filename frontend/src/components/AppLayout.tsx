import { Link, Outlet, useLocation } from 'react-router-dom'
import { Building2, Coins, House, Settings2, UserCircle } from 'lucide-react'
import { useAuth } from '../auth/AuthContext'

export function AppLayout() {
  const { user, clubs, activeClubId, logout, selectClub, uiCaps } = useAuth()
  const location = useLocation()
  const activeClub = clubs.find((c) => c.club_id === activeClubId)

  //todo : change the brand mark and the logo of th e app and the wifdget to the wheel itself at the center of the app
  return (
    <div className="app-shell">
      <header className="app-header">
        <div className="brand">
          <picture>
            <source srcSet="/logo-white.png" media="(prefers-color-scheme: dark)" />
            <img className="brand-logo" src="/logo.png" alt="Rotaract IUGB Club" />
          </picture>
          <div>
            {activeClub?.club && <p>{activeClub.club.name}</p>}
          </div>
        </div>
        <div className="header-actions">
          {clubs.length > 1 && (
            <select
              value={activeClubId ?? ''}
              onChange={(e) => selectClub(e.target.value)}
              className="club-select"
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
          <button type="button" className="btn-ghost" onClick={logout}>
            Déconnexion
          </button>
        </div>
      </header>

      <nav className="bottom-nav">
        {uiCaps.nav.home && (
          <Link to="/home" className={location.pathname === '/home' ? 'active' : ''}>
            <House className="nav-icon" aria-hidden="true" />
            <span>Accueil</span>
          </Link>
        )}
        {activeClubId && uiCaps.nav.club && (
          <Link
            to={`/clubs/${activeClubId}`}
            className={
              location.pathname.startsWith(`/clubs/${activeClubId}`)
              && !location.pathname.includes('/manage')
              && !location.pathname.includes('/cotisations')
                ? 'active'
                : ''
            }
          >
            <Building2 className="nav-icon" aria-hidden="true" />
            <span>{uiCaps.profile === 'commission_president' ? 'Commission' : 'Club'}</span>
          </Link>
        )}
        {activeClubId && uiCaps.nav.cotisations && (
          <Link
            to={`/clubs/${activeClubId}/cotisations`}
            className={location.pathname.includes('/cotisations') ? 'active' : ''}
          >
            <Coins className="nav-icon" aria-hidden="true" />
            <span>Cotisations</span>
          </Link>
        )}
        {activeClubId && uiCaps.nav.manage && (
          <Link to={`/clubs/${activeClubId}/manage`} className={location.pathname.includes('/manage') ? 'active' : ''}>
            <Settings2 className="nav-icon" aria-hidden="true" />
            <span>Gérer</span>
          </Link>
        )}
        {uiCaps.nav.profile && (
          <Link to="/profile" className={location.pathname === '/profile' ? 'active' : ''}>
            <UserCircle className="nav-icon" aria-hidden="true" />
            <span>Profil</span>
          </Link>
        )}
      </nav>
      <main className="app-main">
        <Outlet />
      </main>
    </div>
  )
}
