import { createPortal } from 'react-dom'
import { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { apiRequest } from '../api/client'
import type { FollowProfile } from '../api/types'
import { useAuth } from '../auth/AuthContext'
import { getToken } from '../auth/storage'
import { SkeletonSuggestions } from './Skeleton'

export function SuggestionsPanel() {
  const { token, requireAuth, user } = useAuth()
  const [users, setUsers] = useState<FollowProfile[]>([])
  const [ready, setReady] = useState(false)
  const [slot, setSlot] = useState<HTMLElement | null>(null)

  const load = useCallback(async () => {
    try {
      const data = await apiRequest<{ users: FollowProfile[] }>('/social/suggestions', {}, token)
      setUsers(data.users)
    } catch {
      /* keep previous suggestions on soft refresh */
    } finally {
      setReady(true)
    }
  }, [token])

  useEffect(() => { void load() }, [load])
  useEffect(() => {
    setSlot(document.getElementById('suggestions-slot'))
  }, [])

  function follow(id: string) {
    requireAuth(async () => {
      const removed = users.find((u) => u.id === id)
      setUsers((prev) => prev.filter((u) => u.id !== id))
      try {
        await apiRequest(`/social/users/${id}/follow`, { method: 'POST', body: '{}' }, getToken())
      } catch {
        if (removed) setUsers((prev) => [removed, ...prev])
      }
    }, 'Connectez-vous pour vous abonner')
  }

  const content = !ready && users.length === 0 ? (
    <SkeletonSuggestions count={3} />
  ) : (
    <div className="suggest-card">
      <h3>Suggestions</h3>
      <p className="muted small">Rotaractiens à suivre</p>
      {users.length === 0 && <p className="muted small">Aucune suggestion pour le moment.</p>}
      <ul className="suggest-list">
        {users.map((u) => (
          <li key={u.id}>
            <Link to={`/users/${u.id}`} className="suggest-user">
              <div className="avatar-circle">
                {u.avatar_url ? <img src={u.avatar_url} alt="" /> : u.first_name[0]}
              </div>
              <div>
                <strong>{u.first_name} {u.last_name}</strong>
                <p className="muted small">{u.followers_count} abonnés</p>
              </div>
            </Link>
            {user?.id !== u.id && (
              <button
                type="button"
                className={u.followed_by_me ? 'chip' : 'btn-primary btn-tiny'}
                onClick={() => follow(u.id)}
              >
                {u.followed_by_me ? 'Suivi' : 'Suivre'}
              </button>
            )}
          </li>
        ))}
      </ul>
    </div>
  )

  if (!slot) return <div className="suggest-mobile">{content}</div>
  return createPortal(content, slot)
}
