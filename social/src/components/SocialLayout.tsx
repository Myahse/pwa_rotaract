import { Link, Outlet, useLocation } from 'react-router-dom'
import { useAuth } from '../auth/AuthContext'
import { LoginModal } from './LoginModal'

export function SocialLayout() {
  const { user, logout, openLogin, loading } = useAuth()
  const location = useLocation()

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

          <nav className="top-nav" aria-label="Principal">
            <Link to="/" className={isFeed || isPostDetail ? 'active' : ''}>Fil</Link>
            <Link to="/groups" className={isGroups ? 'active' : ''}>Groupes</Link>
            <Link to="/friends" className={isFriends ? 'active' : ''}>Amis</Link>
            <Link to="/messages" className={isMessages ? 'active' : ''}>Messages</Link>
            <Link to="/me" className={isProfile ? 'active' : ''}>Profil</Link>
          </nav>

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
        <Link to="/" className={isFeed ? 'active' : ''}>Fil</Link>
        <Link to="/groups" className={isGroups ? 'active' : ''}>Groupes</Link>
        <Link to="/friends" className={isFriends ? 'active' : ''}>Amis</Link>
        <Link to="/messages" className={isMessages ? 'active' : ''}>Msg</Link>
        <Link to="/me" className={isProfile ? 'active' : ''}>Profil</Link>
      </nav>

      <LoginModal />
    </div>
  )
}
