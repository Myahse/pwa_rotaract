import { useEffect, useRef, useState, type FormEvent } from 'react'
import { useParams } from 'react-router-dom'
import { apiRequest } from '../api/client'
import type { ChatMessage } from '../api/types'
import { useAuth } from '../auth/AuthContext'

export function ChatPage() {
  const { groupId } = useParams<{ groupId: string }>()
  const { token, user } = useAuth()
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [content, setContent] = useState('')
  const [editingId, setEditingId] = useState<string | null>(null)
  const [editText, setEditText] = useState('')
  const [error, setError] = useState('')
  const bottomRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!token || !groupId) return
    apiRequest<{ messages: ChatMessage[] }>(`/chat/groups/${groupId}/messages`, {}, token)
      .then((res) => setMessages([...res.messages].reverse()))
      .catch((err) => setError(err instanceof Error ? err.message : 'Erreur'))
  }, [token, groupId])

  useEffect(() => {
    if (!token || !groupId) return
    const protocol = window.location.protocol === 'https:' ? 'wss' : 'ws'
    const ws = new WebSocket(`${protocol}://${window.location.host}/api/v1/ws/chat?token=${token}&group_id=${groupId}`)
    ws.onmessage = (event) => {
      try {
        const payload = JSON.parse(event.data) as { type: string; message: ChatMessage }
        if (payload.type === 'message') {
          setMessages((prev) => [...prev, payload.message])
        }
      } catch { /* ignore */ }
    }
    return () => ws.close()
  }, [token, groupId])

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
      setMessages((prev) => {
        if (prev.some((m) => m.id === msg.id)) return prev
        return [...prev, msg]
      })
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
      setMessages((prev) => prev.map((m) => (m.id === id ? updated : m)))
      setEditingId(null)
      setEditText('')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Modification impossible')
    }
  }

  return (
    <div className="chat-page">
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
