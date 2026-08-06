import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { apiRequest } from '../api/client'
import type { FriendProfile, SocialComment } from '../api/types'
import { getToken } from '../auth/storage'

type Props = {
  comment: SocialComment | null
  onClose: () => void
}

export function ShareCommentModal({ comment, onClose }: Props) {
  const navigate = useNavigate()
  const [friends, setFriends] = useState<FriendProfile[]>([])
  const [note, setNote] = useState('Regarde ce commentaire')
  const [error, setError] = useState('')
  const [sending, setSending] = useState(false)

  useEffect(() => {
    if (!comment) return
    void apiRequest<{ friends: FriendProfile[] }>('/social/friends', {}, getToken())
      .then((d) => setFriends(d.friends))
      .catch((err) => setError(err instanceof Error ? err.message : 'Impossible de charger les amis'))
  }, [comment])

  if (!comment) return null

  async function share(friendId: string) {
    setSending(true)
    setError('')
    try {
      const msg = await apiRequest<{ conversation_id: string }>(`/social/comments/${comment!.id}/share`, {
        method: 'POST',
        body: JSON.stringify({ friend_user_id: friendId, note: note.trim() }),
      }, getToken())
      onClose()
      navigate(`/messages/${msg.conversation_id}`)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Partage impossible')
    } finally {
      setSending(false)
    }
  }

  return (
    <div className="modal-backdrop" onClick={onClose} role="presentation">
      <div className="modal-card" role="dialog" aria-modal="true" onClick={(e) => e.stopPropagation()}>
        <button type="button" className="modal-close" onClick={onClose} aria-label="Fermer">×</button>
        <div className="modal-head">
          <h2>Partager à un ami</h2>
          <p className="muted small">« {comment.body.slice(0, 120)}{comment.body.length > 120 ? '…' : ''} »</p>
        </div>
        <label>
          Message
          <input value={note} onChange={(e) => setNote(e.target.value)} />
        </label>
        {error && <p className="error">{error}</p>}
        {friends.length === 0 ? (
          <p className="muted">Ajoutez des amis pour partager un commentaire.</p>
        ) : (
          <ul className="suggest-list">
            {friends.map((f) => (
              <li key={f.id}>
                <div className="suggest-user">
                  <div className="avatar-circle">
                    {f.avatar_url ? <img src={f.avatar_url} alt="" /> : f.first_name[0]}
                  </div>
                  <strong>{f.first_name} {f.last_name}</strong>
                </div>
                <button
                  type="button"
                  className="btn-primary btn-tiny"
                  disabled={sending}
                  onClick={() => share(f.id)}
                >
                  Envoyer
                </button>
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  )
}
