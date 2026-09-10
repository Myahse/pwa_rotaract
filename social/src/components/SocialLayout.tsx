import { Link, Outlet, useLocation } from 'react-router-dom'
import { Building2, CircleUserRound, MessageCircle, Newspaper, Users } from 'lucide-react'
import { useEffect, useState } from 'react'
import { useAuth } from '../auth/AuthContext'
import { LoginModal } from './LoginModal'

export function SocialLayout() {
  const { user, logout, openLogin, loading } = useAuth()
  const location = useLocation()
  const [darkMode, setDarkMode] = useState(() => window.matchMedia('(prefers-color-scheme: dark)').matches)

  useEffect(() => {
    document.documentElement.dataset.theme = darkMode ? 'dark' : 'light'
  }, [darkMode])

  useEffect(() => {
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
    const handleThemeChange = (event: MediaQueryListEvent) => setDarkMode(event.matches)
    mediaQuery.addEventListener('change', handleThemeChange)
    return () => mediaQuery.removeEventListener('change', handleThemeChange)
  }, [])

  const path = location.pathname
  const isFeed = path === '/'
  const isFriends = path.startsWith('/friends')
  const isGroups = path.startsWith('/groups')
  const isMessages = path.startsWith('/messages')
  const isProfile = path === '/me' || path.startsWith('/users/')
  const isFollowing = path.startsWith('/following')
  const isPostDetail = path.startsWith('/posts/')

  return (
    <div className="app-shell">
      <header className="topbar">
        <div className="topbar-inner">
          <Link to="/" className="brand">
            <span className="brand-dot" aria-hidden />
            <span>Rotaract</span>
          </Link>

          <div className="top-actions">
            {loading ? null : user ? (
              <>
                <Link to="/me" className="user-link">
                  {user.avatar_url
                    ? <img src={user.avatar_url} alt="" />
                    : <span>{user.first_name[0]}</span>}
                  <span className="hide-sm">{user.first_name}</span>
                </Link>
                <button type="button" className="btn-text" onClick={logout}>Déconnexion</button>
              </>
            ) : (
              <button type="button" className="btn-primary" onClick={() => openLogin('Connectez-vous à Rotaract Social')}>
                Connexion
              </button>
            )}
          </div>
        </div>
      </header>

      <div className={`page${isPostDetail ? ' page-post' : ''}`}>
        {!isPostDetail && (
          <aside className="rail left hide-mobile">
            <p className="rail-label">Navigation</p>
            <Link to="/" className={isFeed ? 'rail-link active' : 'rail-link'}>Fil d&apos;actualité</Link>
            <Link to="/groups" className={isGroups ? 'rail-link active' : 'rail-link'}>Groupes</Link>
            <Link to="/friends" className={isFriends ? 'rail-link active' : 'rail-link'}>Amis</Link>
            <Link to="/messages" className={isMessages ? 'rail-link active' : 'rail-link'}>Messages</Link>
            <Link to="/following" className={isFollowing ? 'rail-link active' : 'rail-link'}>Abonnements</Link>
            <Link to="/me" className={isProfile ? 'rail-link active' : 'rail-link'}>Profil</Link>
          </aside>
        )}

        <main className={`main${isPostDetail ? ' main-post' : ''}`}>
          <Outlet />
        </main>

        {!isPostDetail && (
          <aside className="rail right hide-mobile" id="suggestions-slot" />
        )}
      </div>

      <nav className={`bottom-nav show-mobile nav-5${isPostDetail ? ' hide-on-post' : ''}`} aria-label="Mobile">
        <Link to="/" className={isFeed ? 'active' : ''}>
          <Newspaper className="nav-icon" aria-hidden="true" />
          <span>Fil</span>
        </Link>
        <Link to="/groups" className={isGroups ? 'active' : ''}>
          <Building2 className="nav-icon" aria-hidden="true" />
          <span>Groupes</span>
        </Link>
        <Link to="/friends" className={isFriends ? 'active' : ''}>
          <Users className="nav-icon" aria-hidden="true" />
          <span>Amis</span>
        </Link>
        <Link to="/messages" className={isMessages ? 'active' : ''}>
          <MessageCircle className="nav-icon" aria-hidden="true" />
          <span>Msg</span>
        </Link>
        <Link to="/me" className={isProfile ? 'active' : ''}>
          <CircleUserRound className="nav-icon" aria-hidden="true" />
          <span>Profil</span>
        </Link>
      </nav>

      <LoginModal />
    </div>
  )
}
