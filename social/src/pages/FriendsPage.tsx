import { useCallback, useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { apiRequest } from '../api/client'
import type { FriendProfile } from '../api/types'
import { useAuth } from '../auth/AuthContext'
import { getToken } from '../auth/storage'
import { SkeletonList } from '../components/Skeleton'

function name(u: FriendProfile) {
  return `${u.first_name} ${u.last_name}`
}

export function FriendsPage() {
  const { user, openLogin } = useAuth()
  const navigate = useNavigate()
  const [friends, setFriends] = useState<FriendProfile[]>([])
  const [requests, setRequests] = useState<FriendProfile[]>([])
  const [error, setError] = useState('')
  const [ready, setReady] = useState(false)

  const load = useCallback(async () => {
    if (!user) {
      setReady(true)
      return
    }
    setError('')
    try {
      const t = getToken()
      const [f, r] = await Promise.all([
        apiRequest<{ friends: FriendProfile[] }>('/social/friends', {}, t),
        apiRequest<{ requests: FriendProfile[] }>('/social/friends/requests', {}, t),
      ])
      setFriends(f.friends)
      setRequests(r.requests)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Impossible de charger les amis')
    } finally {
      setReady(true)
    }
  }, [user])

  useEffect(() => { void load() }, [load])

  async function accept(id: string) {
    const req = requests.find((r) => r.id === id)
    setRequests((prev) => prev.filter((r) => r.id !== id))
    if (req) setFriends((prev) => [req, ...prev])
    try {
      await apiRequest(`/social/friends/${id}/accept`, { method: 'POST', body: '{}' }, getToken())
    } catch {
      await load()
    }
  }

  async function decline(id: string) {
    const snapshot = requests
    setRequests((prev) => prev.filter((r) => r.id !== id))
    try {
      await apiRequest(`/social/friends/${id}/decline`, { method: 'POST', body: '{}' }, getToken())
    } catch {
      setRequests(snapshot)
    }
  }

  async function remove(id: string) {
    if (!confirm('Retirer cet ami ?')) return
    const snapshot = friends
    setFriends((prev) => prev.filter((f) => f.id !== id))
    try {
      await apiRequest(`/social/friends/${id}`, { method: 'DELETE' }, getToken())
    } catch {
      setFriends(snapshot)
    }
  }

  async function message(id: string) {
    const conv = await apiRequest<{ id: string }>('/social/conversations', {
      method: 'POST',
      body: JSON.stringify({ user_id: id }),
    }, getToken())
    navigate(`/messages/${conv.id}`)
  }

  if (!user) {
    return (
      <div className="guest-banner">
        <div>
          <strong>Amis</strong>
          <p>Connectez-vous pour gérer vos amis Rotaract.</p>
        </div>
        <button type="button" className="btn-primary" onClick={() => openLogin()}>Se connecter</button>
      </div>
    )
  }

  return (
    <div className="stack gap-md">
      <header className="page-header">
        <h1>Amis</h1>
        <p className="muted">Demandes, liste d&apos;amis et messagerie</p>
      </header>

      {error && <p className="error">{error}</p>}

      {(!ready && friends.length === 0 && requests.length === 0) && (
        <section className="panel stack">
          <h2>Mes amis</h2>
          <SkeletonList count={4} />
        </section>
      )}

      {ready && requests.length > 0 && (
        <section className="panel stack">
          <h2>Demandes reçues</h2>
          <ul className="suggest-list">
            {requests.map((u) => (
              <li key={u.id}>
                <Link to={`/users/${u.id}`} className="suggest-user">
                  <div className="avatar-circle">
                    {u.avatar_url ? <img src={u.avatar_url} alt="" /> : u.first_name[0]}
                  </div>
                  <div>
                    <strong>{name(u)}</strong>
                    <p className="muted small">Souhaite devenir ami</p>
                  </div>
                </Link>
                <div className="actions">
                  <button type="button" className="btn-primary btn-tiny" onClick={() => accept(u.id)}>Accepter</button>
                  <button type="button" className="chip" onClick={() => decline(u.id)}>Refuser</button>
                </div>
              </li>
            ))}
          </ul>
        </section>
      )}

      {(ready || friends.length > 0) && (
        <section className="panel stack">
          <h2>Mes amis ({friends.length})</h2>
          {friends.length === 0 ? (
            <p className="muted">Aucun ami pour le moment. Suivez des membres puis envoyez une demande d&apos;ami.</p>
          ) : (
            <ul className="suggest-list">
              {friends.map((u) => (
                <li key={u.id}>
                  <Link to={`/users/${u.id}`} className="suggest-user">
                    <div className="avatar-circle">
                      {u.avatar_url ? <img src={u.avatar_url} alt="" /> : u.first_name[0]}
                    </div>
                    <strong>{name(u)}</strong>
                  </Link>
                  <div className="actions">
                    <button type="button" className="btn-primary btn-tiny" onClick={() => message(u.id)}>Message</button>
                    <button type="button" className="chip" onClick={() => remove(u.id)}>Retirer</button>
                  </div>
                </li>
              ))}
            </ul>
          )}
        </section>
      )}
    </div>
  )
}
