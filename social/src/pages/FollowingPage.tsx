import { useCallback, useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { apiRequest } from '../api/client'
import type { FollowProfile } from '../api/types'
import { useAuth } from '../auth/AuthContext'
import { getToken } from '../auth/storage'
import { SkeletonList } from '../components/Skeleton'

export function FollowingPage() {
  const { user, openLogin, requireAuth, token } = useAuth()
  const navigate = useNavigate()
  const [following, setFollowing] = useState<FollowProfile[]>([])
  const [suggestions, setSuggestions] = useState<FollowProfile[]>([])
  const [error, setError] = useState('')
  const [followingReady, setFollowingReady] = useState(!token)
  const [suggestionsReady, setSuggestionsReady] = useState(false)

  const load = useCallback(async () => {
    try {
      const sug = await apiRequest<{ users: FollowProfile[] }>('/social/suggestions', {}, token)
      setSuggestions(sug.users)
      if (token) {
        const data = await apiRequest<{ users: FollowProfile[] }>('/social/following', {}, token)
        setFollowing(data.users)
      } else {
        setFollowing([])
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erreur')
    } finally {
      setSuggestionsReady(true)
      setFollowingReady(true)
    }
  }, [token])

  useEffect(() => { void load() }, [load])

  function toggle(u: FollowProfile) {
    requireAuth(async () => {
      const t = getToken()
      const next = !u.followed_by_me
      if (next) {
        setSuggestions((prev) => prev.filter((x) => x.id !== u.id))
        setFollowing((prev) => [{ ...u, followed_by_me: true }, ...prev.filter((x) => x.id !== u.id)])
      } else {
        setFollowing((prev) => prev.filter((x) => x.id !== u.id))
        setSuggestions((prev) => [{ ...u, followed_by_me: false }, ...prev.filter((x) => x.id !== u.id)])
      }
      try {
        if (u.followed_by_me) {
          await apiRequest(`/social/users/${u.id}/follow`, { method: 'DELETE' }, t)
        } else {
          await apiRequest(`/social/users/${u.id}/follow`, { method: 'POST', body: '{}' }, t)
        }
      } catch {
        await load()
      }
    }, 'Connectez-vous pour gérer vos abonnements')
  }

  function addFriend(id: string) {
    requireAuth(async () => {
      setFollowing((prev) => prev.map((u) => (
        u.id === id ? { ...u, friendship_status: 'pending' } : u
      )))
      try {
        await apiRequest(`/social/friends/${id}/request`, { method: 'POST', body: '{}' }, getToken())
      } catch {
        await load()
      }
    }, 'Connectez-vous pour ajouter un ami')
  }

  function message(id: string) {
    requireAuth(async () => {
      const conv = await apiRequest<{ id: string }>('/social/conversations', {
        method: 'POST',
        body: JSON.stringify({ user_id: id }),
      }, getToken())
      navigate(`/messages/${conv.id}`)
    }, 'Connectez-vous pour envoyer un message')
  }

  return (
    <div className="stack gap-md">
      <header className="page-header">
        <h1>Abonnements</h1>
        <p className="muted">Ouvrez un profil, ajoutez en ami, puis envoyez un message</p>
      </header>

      {!user && (
        <div className="guest-banner">
          <div>
            <strong>Mode invité</strong>
            <p>Connectez-vous pour suivre des membres et filtrer le fil « Abonnements ».</p>
          </div>
          <button type="button" className="btn-primary" onClick={() => openLogin()}>Se connecter</button>
        </div>
      )}

      {error && <p className="error">{error}</p>}

      {user && (
        <section className="panel stack">
          <h2>Vous suivez {followingReady ? `(${following.length})` : ''}</h2>
          {(!followingReady && following.length === 0) && <SkeletonList count={4} />}
          {followingReady && following.length === 0 && (
            <p className="muted">Vous ne suivez personne pour le moment.</p>
          )}
          {following.length > 0 && (
            <ul className="suggest-list">
              {following.map((u) => (
                <li key={u.id}>
                  <Link to={`/users/${u.id}`} className="suggest-user">
                    <div className="avatar-circle">{u.avatar_url ? <img src={u.avatar_url} alt="" /> : u.first_name[0]}</div>
                    <div>
                      <strong>{u.first_name} {u.last_name}</strong>
                      <p className="muted small">{u.followers_count} abonnés · Voir le profil</p>
                    </div>
                  </Link>
                  <div className="actions">
                    {u.friendship_status === 'accepted' ? (
                      <button type="button" className="btn-primary btn-tiny" onClick={() => message(u.id)}>Message</button>
                    ) : u.friendship_status === 'pending' ? (
                      <span className="chip">Demande envoyée</span>
                    ) : (
                      <button type="button" className="btn-primary btn-tiny" onClick={() => addFriend(u.id)}>Ajouter ami</button>
                    )}
                    <button type="button" className="chip" onClick={() => toggle({ ...u, followed_by_me: true })}>Ne plus suivre</button>
                  </div>
                </li>
              ))}
            </ul>
          )}
        </section>
      )}

      <section className="panel stack">
        <h2>À découvrir</h2>
        {(!suggestionsReady && suggestions.length === 0) && <SkeletonList count={3} />}
        {suggestions.length > 0 && (
          <ul className="suggest-list">
            {suggestions.map((u) => (
              <li key={u.id}>
                <Link to={`/users/${u.id}`} className="suggest-user">
                  <div className="avatar-circle">{u.avatar_url ? <img src={u.avatar_url} alt="" /> : u.first_name[0]}</div>
                  <div>
                    <strong>{u.first_name} {u.last_name}</strong>
                    <p className="muted small">{u.followers_count} abonnés · Voir le profil</p>
                  </div>
                </Link>
                <button type="button" className="btn-primary btn-tiny" onClick={() => toggle(u)}>Suivre</button>
              </li>
            ))}
          </ul>
        )}
      </section>
    </div>
  )
}
