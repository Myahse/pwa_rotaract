import { useCallback, useEffect, useRef, useState, type FormEvent } from 'react'
import { Link, useParams } from 'react-router-dom'
import { apiRequest } from '../api/client'
import type { ChatGroup, ChatMessage } from '../api/types'
import { useAuth } from '../auth/AuthContext'

function upsertMessage(list: ChatMessage[], message: ChatMessage): ChatMessage[] {
  const idx = list.findIndex((m) => m.id === message.id)
  if (idx >= 0) {
    const next = [...list]
    next[idx] = message
    return next
  }
  return [...list, message]
}

export function ChatPage() {
  const { groupId } = useParams<{ groupId: string }>()
  const { token, user, activeClubId } = useAuth()
  const [group, setGroup] = useState<ChatGroup | null>(null)
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [content, setContent] = useState('')
  const [editingId, setEditingId] = useState<string | null>(null)
  const [editText, setEditText] = useState('')
  const [error, setError] = useState('')
  const [connected, setConnected] = useState(false)
  const bottomRef = useRef<HTMLDivElement>(null)
  const wsRef = useRef<WebSocket | null>(null)
  const reconnectTimer = useRef<number | null>(null)

  const loadMessages = useCallback(async () => {
    if (!token || !groupId) return
    const res = await apiRequest<{ messages: ChatMessage[] }>(`/chat/groups/${groupId}/messages`, {}, token)
    setMessages(res.messages)
  }, [token, groupId])

  useEffect(() => {
    if (!token || !groupId) return
    apiRequest<{ group: ChatGroup }>(`/chat/groups/${groupId}`, {}, token)
      .then((res) => setGroup(res.group))
      .catch((err) => setError(err instanceof Error ? err.message : 'Erreur'))
    loadMessages().catch((err) => setError(err instanceof Error ? err.message : 'Erreur'))
  }, [token, groupId, loadMessages])

  useEffect(() => {
    if (!token || !groupId) return

    let closed = false
    let attempt = 0

    function connect() {
      if (closed) return
      const protocol = window.location.protocol === 'https:' ? 'wss' : 'ws'
      const ws = new WebSocket(`${protocol}://${window.location.host}/api/v1/ws/chat?token=${token}&group_id=${groupId}`)
      wsRef.current = ws

      ws.onopen = () => {
        attempt = 0
        setConnected(true)
        void loadMessages()
      }

      ws.onmessage = (event) => {
        try {
          const payload = JSON.parse(event.data) as {
            type: string
            message?: ChatMessage
            message_id?: string
          }
          if (payload.type === 'message' || payload.type === 'message_updated') {
            if (payload.message) {
              setMessages((prev) => upsertMessage(prev, payload.message!))
            }
            return
          }
          if (payload.type === 'message_deleted' && payload.message_id) {
            setMessages((prev) => prev.filter((m) => m.id !== payload.message_id))
          }
        } catch { /* ignore */ }
      }

      ws.onclose = () => {
        setConnected(false)
        if (closed) return
        attempt += 1
        const delay = Math.min(1000 * 2 ** attempt, 15000)
        reconnectTimer.current = window.setTimeout(connect, delay)
      }

      ws.onerror = () => ws.close()
    }

    connect()

    return () => {
      closed = true
      if (reconnectTimer.current) window.clearTimeout(reconnectTimer.current)
      wsRef.current?.close()
    }
  }, [token, groupId, loadMessages])

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  async function send(e: FormEvent) {
    e.preventDefault()
    if (!token || !groupId || !content.trim()) return
    const text = content.trim()
    setContent('')
    try {
      const msg = await apiRequest<ChatMessage>(`/chat/groups/${groupId}/messages`, {
        method: 'POST',
        body: JSON.stringify({ content: text }),
      }, token)
      setMessages((prev) => upsertMessage(prev, msg))
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Envoi impossible')
    }
  }

  async function deleteMessage(id: string) {
    if (!token || !groupId) return
    try {
      await apiRequest(`/chat/groups/${groupId}/messages/${id}`, { method: 'DELETE' }, token)
      setMessages((prev) => prev.filter((m) => m.id !== id))
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Suppression impossible')
    }
  }

  async function saveEdit(id: string) {
    if (!token || !groupId || !editText.trim()) return
    try {
      const updated = await apiRequest<ChatMessage>(`/chat/groups/${groupId}/messages/${id}`, {
        method: 'PATCH',
        body: JSON.stringify({ content: editText.trim() }),
      }, token)
      setMessages((prev) => upsertMessage(prev, updated))
      setEditingId(null)
      setEditText('')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Modification impossible')
    }
  }

  const backLink = activeClubId ? `/clubs/${activeClubId}/messages` : '/home'

  return (
    <div className="chat-page">
      <div className="chat-header">
        <Link to={backLink} className="back">← Messages</Link>
        <div>
          <h2>{group?.name ?? 'Discussion'}</h2>
          <p className="muted small">{connected ? 'En ligne' : 'Reconnexion…'}</p>
        </div>
      </div>
      {error && <p className="error">{error}</p>}
      <div className="chat-messages">
        {messages.map((m) => (
          <div key={m.id} className={`bubble ${m.user_id === user?.id ? 'mine' : ''}`}>
            <strong>{m.user?.first_name ?? 'Membre'}</strong>
            {editingId === m.id ? (
              <div className="stack">
                <input value={editText} onChange={(e) => setEditText(e.target.value)} />
                <div className="actions">
                  <button type="button" className="btn-small" onClick={() => saveEdit(m.id)}>OK</button>
                  <button type="button" className="btn-ghost btn-small" onClick={() => setEditingId(null)}>Annuler</button>
                </div>
              </div>
            ) : (
              <p>{m.content}</p>
            )}
            <time>{new Date(m.created_at).toLocaleTimeString('fr-FR', { hour: '2-digit', minute: '2-digit' })}</time>
            {m.user_id === user?.id && editingId !== m.id && (
              <div className="actions">
                <button type="button" className="btn-ghost btn-small" onClick={() => { setEditingId(m.id); setEditText(m.content) }}>Modifier</button>
                <button type="button" className="btn-ghost btn-small" onClick={() => deleteMessage(m.id)}>Supprimer</button>
              </div>
            )}
          </div>
        ))}
        <div ref={bottomRef} />
      </div>
      <form onSubmit={send} className="chat-input">
        <input value={content} onChange={(e) => setContent(e.target.value)} placeholder="Votre message…" />
        <button type="submit" className="btn-primary">Envoyer</button>
      </form>
    </div>
  )
}
