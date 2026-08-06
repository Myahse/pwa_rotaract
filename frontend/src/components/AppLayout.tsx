import { Link, Outlet, useLocation } from 'react-router-dom'
import { useAuth } from '../auth/AuthContext'

export function AppLayout() {
  const { user, clubs, activeClubId, logout, selectClub } = useAuth()
  const location = useLocation()
  const activeClub = clubs.find((c) => c.club_id === activeClubId)

  //todo : change the brand mark and the logo of th e app and the wifdget to the wheel itself at the center of the app
  return (
    <div className="app-shell">
      <header className="app-header">
        <div className="brand">
          <span className="brand-dot" aria-hidden />
          <div>
            <strong>Rotaract CIV</strong>
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
          <span className="user-chip">{user?.first_name}</span>
          <button type="button" className="btn-ghost" onClick={logout}>
            Déconnexion
          </button>
        </div>
      </header>

      <nav className="bottom-nav">
        <Link to="/home" className={location.pathname === '/home' ? 'active' : ''}>Accueil</Link>
        {activeClubId && (
          <Link to={`/clubs/${activeClubId}`} className={location.pathname.includes('/clubs/') ? 'active' : ''}>
            Club
          </Link>
        )}
        <Link to="/profile" className={location.pathname === '/profile' ? 'active' : ''}>Profil</Link>
      </nav>
      <main className="app-main">
        <Outlet />
      </main>
    </div>
  )
}
