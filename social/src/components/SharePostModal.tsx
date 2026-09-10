import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Repeat2, Share2, UserRound } from 'lucide-react'
import { apiRequest } from '../api/client'
import type { FriendProfile, SocialPost } from '../api/types'
import { getToken } from '../auth/storage'

type Props = {
  post: SocialPost | null
  onClose: () => void
  onRepost: (quoteBody: string) => void
}

export function SharePostModal({ post, onClose, onRepost }: Props) {
  const navigate = useNavigate()
  const [showFriends, setShowFriends] = useState(false)
  const [friends, setFriends] = useState<FriendProfile[]>([])
  const [error, setError] = useState('')
  const [sending, setSending] = useState(false)
  const [quoteBody, setQuoteBody] = useState('')

  useEffect(() => {
    if (!post || !showFriends) return
    void apiRequest<{ friends: FriendProfile[] }>('/social/friends', {}, getToken())
      .then((data) => setFriends(data.friends))
      .catch((err) => setError(err instanceof Error ? err.message : 'Impossible de charger les amis'))
  }, [post, showFriends])

  if (!post) return null
  const currentPost = post

  async function shareWithFriend(friendID: string) {
    setSending(true)
    setError('')
    try {
      const conversation = await apiRequest<{ id: string }>('/social/conversations', {
        method: 'POST',
        body: JSON.stringify({ user_id: friendID }),
      }, getToken())
      await apiRequest(`/social/conversations/${conversation.id}/messages`, {
        method: 'POST',
        body: JSON.stringify({ body: '', shared_post_id: currentPost.id }),
      }, getToken())
      onClose()
      navigate(`/messages/${conversation.id}`)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Partage impossible')
    } finally {
      setSending(false)
    }
  }

  async function shareElsewhere() {
    const url = `${window.location.origin}/posts/${currentPost.id}`
    try {
      if (navigator.share) await navigator.share({ title: 'Rotaract Social', url })
      else {
        await navigator.clipboard.writeText(url)
        alert('Lien copié')
      }
      onClose()
    } catch { /* cancelled */ }
  }

  return (
    <div className="modal-backdrop" onClick={onClose} role="presentation">
      <div className="modal-card share-post-modal" role="dialog" aria-modal="true" onClick={(event) => event.stopPropagation()}>
        <button type="button" className="modal-close" onClick={onClose} aria-label="Fermer">×</button>
        <div className="modal-head">
          <h2>Repartager</h2>
          <p className="muted small">Choisissez comment partager cette publication.</p>
        </div>

        {!showFriends ? (
          <div className="share-quote-stack">
            <label className="share-quote-label">
              Ajouter un texte <span className="muted small">(facultatif)</span>
              <textarea value={quoteBody} onChange={(event) => setQuoteBody(event.target.value)} maxLength={5000} rows={2} />
            </label>
            <div className="share-options">
              <button type="button" className="share-option" onClick={() => { onRepost(quoteBody.trim()); onClose() }}>
                <span className="share-option-icon"><Repeat2 size={19} aria-hidden="true" /></span>
                <strong>Reposter</strong>
              </button>
              <button type="button" className="share-option" onClick={() => setShowFriends(true)}>
                <span className="share-option-icon"><UserRound size={19} aria-hidden="true" /></span>
                <strong>Avec un ami</strong>
              </button>
              <button type="button" className="share-option" onClick={shareElsewhere}>
                <span className="share-option-icon"><Share2 size={19} aria-hidden="true" /></span>
                <strong>Autres plateformes</strong>
              </button>
            </div>
          </div>
        ) : (
          <div className="share-friends">
            <button type="button" className="btn-text" onClick={() => setShowFriends(false)}>← Retour</button>
            {error && <p className="error">{error}</p>}
            {friends.length === 0 ? (
              <p className="muted">Ajoutez des amis pour leur envoyer une publication.</p>
            ) : (
              <ul className="suggest-list">
                {friends.map((friend) => (
                  <li key={friend.id}>
                    <div className="suggest-user">
                      <div className="avatar-circle">
                        {friend.avatar_url ? <img src={friend.avatar_url} alt="" /> : friend.first_name[0]}
                      </div>
                      <strong>{friend.first_name} {friend.last_name}</strong>
                    </div>
                    <button type="button" className="btn-primary btn-tiny" disabled={sending} onClick={() => shareWithFriend(friend.id)}>
                      Envoyer
                    </button>
                  </li>
                ))}
              </ul>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
