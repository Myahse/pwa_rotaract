import { Link, Outlet, useLocation } from 'react-router-dom'
import { useAuth } from '../auth/AuthContext'
import { IconAccess, IconCalendar, IconClubs, IconDashboard, IconDonate, IconRegister, IconWebsite } from './Icons'

const nav = [
  { to: '/dashboard', label: 'Accueil', short: 'Accueil', icon: IconDashboard, match: (p: string) => p === '/dashboard' || p === '/' },
  { to: '/clubs', label: 'Clubs', short: 'Clubs', icon: IconClubs, match: (p: string) => p.startsWith('/clubs') },
  { to: '/events', label: 'Événements', short: 'Agenda', icon: IconCalendar, match: (p: string) => p.startsWith('/events') },
  { to: '/website-home', label: 'Accueil site', short: 'Site', icon: IconWebsite, match: (p: string) => p.startsWith('/website-home') },
  { to: '/access-requests', label: "Demandes d'accès", short: 'Accès', icon: IconAccess, match: (p: string) => p.startsWith('/access-requests') },
  { to: '/donations', label: 'Dons', short: 'Dons', icon: IconDonate, match: (p: string) => p.startsWith('/donations') },
  { to: '/club-registrations', label: 'Nouveaux clubs', short: 'Clubs+', icon: IconRegister, match: (p: string) => p.startsWith('/club-registrations') },
]

export function AdminLayout() {
  const { user, logout } = useAuth()
  const location = useLocation()
  const initials = `${user?.first_name?.[0] ?? ''}${user?.last_name?.[0] ?? ''}`.toUpperCase() || 'A'

  return (
    <div className="admin-shell">
      <header className="admin-topbar">
        <div className="brand">
          <span className="brand-dot" aria-hidden />
          <div>
            <strong>Rotaract Admin</strong>
            <p>Côte d&apos;Ivoire</p>
          </div>
        </div>
        <div className="user-chip">
          <span className="avatar-dot" aria-hidden>{initials}</span>
          <span>{user?.first_name || user?.email}</span>
        </div>
      </header>

      <aside className="admin-sidebar">
        <div className="brand">
          <span className="brand-dot" aria-hidden />
          <div>
            <strong>Rotaract Admin</strong>
            <p>Côte d&apos;Ivoire</p>
          </div>
        </div>
        <nav aria-label="Navigation principale">
          {nav.map((item) => {
            const Icon = item.icon
            const active = item.match(location.pathname)
            return (
              <Link key={item.to} to={item.to} className={active ? 'active' : ''} aria-current={active ? 'page' : undefined}>
                <Icon />
                {item.label}
              </Link>
            )
          })}
        </nav>
        <div className="sidebar-footer">
          <div className="user-chip" style={{ maxWidth: '100%', color: 'inherit' }}>
            <span className="avatar-dot">{initials}</span>
            <span style={{ color: 'rgba(255,255,255,0.85)' }}>{user?.email}</span>
          </div>
          <button type="button" className="logout-btn" onClick={logout}>Déconnexion</button>
        </div>
      </aside>

      <main className="admin-main">
        <Outlet />
      </main>

      <nav className="admin-bottom-nav" aria-label="Navigation mobile">
        {nav.map((item) => {
          const Icon = item.icon
          const active = item.match(location.pathname)
          return (
            <Link key={item.to} to={item.to} className={`nav-item ${active ? 'active' : ''}`} aria-current={active ? 'page' : undefined}>
              <Icon />
              {item.short}
            </Link>
          )
        })}
      </nav>
    </div>
  )
}
