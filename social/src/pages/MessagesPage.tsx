import { useCallback, useEffect, useState, type FormEvent } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { apiRequest } from '../api/client'
import type { SocialConversation, SocialMessage } from '../api/types'
import { useAuth } from '../auth/AuthContext'
import { getToken } from '../auth/storage'
import { SkeletonList, SkeletonThread } from '../components/Skeleton'

function peerName(c: SocialConversation) {
  const p = c.peer
  if (!p) return 'Conversation'
  return `${p.first_name} ${p.last_name}`
}

export function MessagesPage() {
  const { conversationID } = useParams()
  const { user, openLogin } = useAuth()
  const navigate = useNavigate()
  const [conversations, setConversations] = useState<SocialConversation[]>([])
  const [messages, setMessages] = useState<SocialMessage[]>([])
  const [body, setBody] = useState('')
  const [error, setError] = useState('')
  const [loadingList, setLoadingList] = useState(true)
  const [loadingThread, setLoadingThread] = useState(false)

  const loadConversations = useCallback(async () => {
    if (!user) {
      setLoadingList(false)
      return
    }
    setLoadingList(true)
    try {
      const data = await apiRequest<{ conversations: SocialConversation[] }>('/social/conversations', {}, getToken())
      setConversations(data.conversations)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Impossible de charger les messages')
    } finally {
      setLoadingList(false)
    }
  }, [user])

  const loadMessages = useCallback(async () => {
    if (!conversationID || !user) {
      setMessages([])
      setLoadingThread(false)
      return
    }
    setLoadingThread(true)
    try {
      const data = await apiRequest<{ messages: SocialMessage[] }>(
        `/social/conversations/${conversationID}/messages?limit=80`,
        {},
        getToken(),
      )
      setMessages(data.messages)
    } finally {
      setLoadingThread(false)
    }
  }, [conversationID, user])

  useEffect(() => { void loadConversations() }, [loadConversations])
  useEffect(() => { void loadMessages() }, [loadMessages])

  async function send(e: FormEvent) {
    e.preventDefault()
    if (!conversationID || !body.trim()) return
    await apiRequest(`/social/conversations/${conversationID}/messages`, {
      method: 'POST',
      body: JSON.stringify({ body: body.trim() }),
    }, getToken())
    setBody('')
    await loadMessages()
    await loadConversations()
  }

  if (!user) {
    return (
      <div className="guest-banner">
        <div>
          <strong>Messages</strong>
          <p>Connectez-vous pour écrire à vos amis.</p>
        </div>
        <button type="button" className="btn-primary" onClick={() => openLogin()}>Se connecter</button>
      </div>
    )
  }

  const active = conversations.find((c) => c.id === conversationID)

  return (
    <div className="stack gap-md">
      <header className="page-header">
        <h1>Messages</h1>
        <p className="muted">Discussions privées entre amis</p>
      </header>

      {error && <p className="error">{error}</p>}

      {!conversationID && (
        <section className="panel stack">
          {loadingList && <SkeletonList count={5} />}
          {!loadingList && conversations.length === 0 && (
            <div className="empty">
              <strong>Aucune conversation</strong>
              <p>Ajoutez un ami puis ouvrez un message depuis Amis ou un profil.</p>
              <Link to="/friends" className="btn-primary btn-tiny">Voir mes amis</Link>
            </div>
          )}
          {!loadingList && (
            <ul className="suggest-list">
              {conversations.map((c) => (
                <li key={c.id}>
                  <button type="button" className="suggest-user conv-row" onClick={() => navigate(`/messages/${c.id}`)}>
                    <div className="avatar-circle">
                      {c.peer?.avatar_url
                        ? <img src={c.peer.avatar_url} alt="" />
                        : (c.peer?.first_name?.[0] ?? '?')}
                    </div>
                    <div>
                      <strong>{peerName(c)}</strong>
                      <p className="muted small">
                        {c.last_message?.shared_comment_id
                          ? 'Commentaire partagé'
                          : (c.last_message?.body || 'Nouvelle conversation')}
                      </p>
                    </div>
                  </button>
                </li>
              ))}
            </ul>
          )}
        </section>
      )}

      {conversationID && (
        <section className="dm-panel">
          <div className="dm-head">
            <button type="button" className="btn-text" onClick={() => navigate('/messages')}>← Retour</button>
            <strong>{active ? peerName(active) : 'Conversation'}</strong>
          </div>
          {loadingThread ? (
            <SkeletonThread />
          ) : (
            <div className="dm-thread">
              {messages.map((m) => (
                <div key={m.id} className={m.sender_id === user.id ? 'dm-bubble mine' : 'dm-bubble'}>
                  <p>{m.body}</p>
                  {m.shared_comment && (
                    <div className="share-card">
                      <p className="muted small">
                        Commentaire de {m.shared_comment.author?.first_name} {m.shared_comment.author?.last_name}
                      </p>
                      <p>{m.shared_comment.body}</p>
                      {m.shared_post_id && (
                        <Link to={`/posts/${m.shared_post_id}`} className="muted small">Voir la publication</Link>
                      )}
                    </div>
                  )}
                  <time className="muted small">{new Date(m.created_at).toLocaleString('fr-FR')}</time>
                </div>
              ))}
            </div>
          )}
          <form onSubmit={send} className="comment-form">
            <input
              value={body}
              onChange={(e) => setBody(e.target.value)}
              placeholder="Écrire un message…"
            />
            <button type="submit" className="btn-primary btn-tiny">Envoyer</button>
          </form>
        </section>
      )}
    </div>
  )
}
