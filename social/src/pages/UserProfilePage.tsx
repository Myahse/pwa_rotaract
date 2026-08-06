import { useCallback, useEffect, useState } from 'react'
import { Link, useLocation, useNavigate, useParams } from 'react-router-dom'
import { apiRequest } from '../api/client'
import type { SocialPost, SocialUserProfile } from '../api/types'
import { useAuth } from '../auth/AuthContext'
import { getToken } from '../auth/storage'
import { SkeletonProfile } from '../components/Skeleton'

export function UserProfilePage() {
  const { userID } = useParams()
  const location = useLocation()
  const { user, openLogin, requireAuth } = useAuth()
  const navigate = useNavigate()
  const [profile, setProfile] = useState<SocialUserProfile | null>(null)
  const [posts, setPosts] = useState<SocialPost[]>([])
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)

  const isOwnRoute = location.pathname === '/me' || userID === 'me'
  const id = isOwnRoute ? user?.id : userID

  const load = useCallback(async () => {
    if (!id) {
      setLoading(false)
      return
    }
    setLoading(true)
    setError('')
    try {
      const [p, feed] = await Promise.all([
        apiRequest<SocialUserProfile>(`/social/users/${id}`, {}, getToken()),
        apiRequest<{ posts: SocialPost[] }>(`/social/users/${id}/posts?limit=20`, {}, getToken()),
      ])
      setProfile(p)
      setPosts(feed.posts)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Profil introuvable')
    } finally {
      setLoading(false)
    }
  }, [id])

  useEffect(() => { void load() }, [load])

  if (isOwnRoute && !user) {
    return (
      <div className="guest-banner">
        <div>
          <strong>Profil</strong>
          <p>Connectez-vous pour voir votre profil.</p>
        </div>
        <button type="button" className="btn-primary" onClick={() => openLogin()}>Se connecter</button>
      </div>
    )
  }

  if (loading) {
    return (
      <div className="stack gap-md">
        <SkeletonProfile />
      </div>
    )
  }
  if (error || !profile) return <p className="error">{error || 'Profil introuvable'}</p>

  function friendAction() {
    if (!profile) return
    requireAuth(async () => {
      const t = getToken()
      if (profile.incoming_request) {
        await apiRequest(`/social/friends/${profile.id}/accept`, { method: 'POST', body: '{}' }, t)
      } else if (profile.friendship_status === 'accepted') {
        if (!confirm('Retirer cet ami ?')) return
        await apiRequest(`/social/friends/${profile.id}`, { method: 'DELETE' }, t)
      } else if (profile.friendship_status === 'pending') {
        return
      } else {
        await apiRequest(`/social/friends/${profile.id}/request`, { method: 'POST', body: '{}' }, t)
      }
      await load()
    }, 'Connectez-vous pour gérer vos amis')
  }

  function followToggle() {
    if (!profile) return
    requireAuth(async () => {
      const t = getToken()
      if (profile.followed_by_me) {
        await apiRequest(`/social/users/${profile.id}/follow`, { method: 'DELETE' }, t)
      } else {
        await apiRequest(`/social/users/${profile.id}/follow`, { method: 'POST', body: '{}' }, t)
      }
      await load()
    }, 'Connectez-vous pour vous abonner')
  }

  function openMessage() {
    requireAuth(async () => {
      const conv = await apiRequest<{ id: string }>('/social/conversations', {
        method: 'POST',
        body: JSON.stringify({ user_id: profile!.id }),
      }, getToken())
      navigate(`/messages/${conv.id}`)
    }, 'Connectez-vous pour envoyer un message')
  }

  const friendLabel = profile.is_me
    ? null
    : profile.incoming_request
      ? 'Accepter'
      : profile.friendship_status === 'accepted'
        ? 'Amis'
        : profile.friendship_status === 'pending'
          ? 'Demande envoyée'
          : 'Ajouter ami'

  return (
    <div className="stack gap-md">
      <section className="profile-hero panel">
        <div className="avatar-circle xl">
          {profile.avatar_url ? <img src={profile.avatar_url} alt="" /> : profile.first_name[0]}
        </div>
        <div className="stack" style={{ gap: '0.35rem', flex: 1 }}>
          <h1>{profile.first_name} {profile.last_name}</h1>
          <p className="muted small">
            {profile.friends_count} amis · {profile.followers_count} abonnés · {profile.following_count} abonnements
          </p>
          {!profile.is_me && (
            <div className="actions">
              {friendLabel && (
                <button
                  type="button"
                  className={profile.friendship_status === 'accepted' ? 'chip' : 'btn-primary btn-tiny'}
                  onClick={friendAction}
                  disabled={profile.friendship_status === 'pending' && !profile.incoming_request}
                >
                  {friendLabel}
                </button>
              )}
              <button type="button" className="chip" onClick={followToggle}>
                {profile.followed_by_me ? 'Abonné' : 'Suivre'}
              </button>
              {profile.friendship_status === 'accepted' && (
                <button type="button" className="btn-primary btn-tiny" onClick={openMessage}>Message</button>
              )}
            </div>
          )}
          {profile.is_me && (
            <div className="actions">
              <Link to="/friends" className="btn-primary btn-tiny">Mes amis</Link>
              <Link to="/messages" className="chip">Messages</Link>
            </div>
          )}
        </div>
      </section>

      <section className="panel stack">
        <h2>Clubs</h2>
        {!profile.clubs?.length ? (
          <p className="muted">Aucun club associé.</p>
        ) : (
          <ul className="club-list">
            {profile.clubs.map((c) => <li key={c.id}>{c.name}</li>)}
          </ul>
        )}
      </section>

      <section className="panel stack">
        <h2>Publications</h2>
        {posts.length === 0 && <p className="muted">Aucune publication.</p>}
        <div className="feed">
          {posts.map((post) => (
            <article key={post.id} className="post-card">
              <Link to={`/posts/${post.id}`} className="stack" style={{ gap: '0.35rem' }}>
                <p className="muted small">{new Date(post.created_at).toLocaleString('fr-FR')}</p>
                {post.body && <p className="post-body">{post.body}</p>}
                <p className="muted small">{post.reaction_count} j&apos;aime · {post.comment_count} commentaires</p>
              </Link>
            </article>
          ))}
        </div>
      </section>
    </div>
  )
}
