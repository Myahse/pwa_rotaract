import { useCallback, useEffect, useMemo, useRef, useState, type FormEvent } from 'react'
import { Link, useParams } from 'react-router-dom'
import { ArrowLeft, Info, Send, X } from 'lucide-react'
import { apiRequest, chatWebSocketURL } from '../api/client'
import type { ChatGroup, ChatGroupMember, ChatMessage, ClubMember } from '../api/types'
import { useAuth } from '../auth/AuthContext'
import { ChatAvatar } from '../components/ChatAvatar'

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
  const [groupMembers, setGroupMembers] = useState<ChatGroupMember[]>([])
  const [clubMembers, setClubMembers] = useState<ClubMember[]>([])
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [content, setContent] = useState('')
  const [editingId, setEditingId] = useState<string | null>(null)
  const [editText, setEditText] = useState('')
  const [error, setError] = useState('')
  const [connected, setConnected] = useState(false)
  const [showInfo, setShowInfo] = useState(false)
  const [addMemberIds, setAddMemberIds] = useState<string[]>([])
  const [busy, setBusy] = useState(false)
  const bottomRef = useRef<HTMLDivElement>(null)
  const wsRef = useRef<WebSocket | null>(null)
  const reconnectTimer = useRef<number | null>(null)

  const isDirect = group?.group_type === 'custom' && groupMembers.length === 2
  const isCustomGroup = group?.group_type === 'custom' && groupMembers.length > 2

  const loadGroup = useCallback(async () => {
    if (!token || !groupId) return
    const res = await apiRequest<{ group: ChatGroup; members: ChatGroupMember[] }>(
      `/chat/groups/${groupId}`,
      {},
      token,
    )
    setGroup(res.group)
    setGroupMembers(res.members)
    if (res.group.club_id) {
      const roster = await apiRequest<ClubMember[]>(`/clubs/${res.group.club_id}/members`, {}, token)
      setClubMembers(roster)
    }
  }, [token, groupId])

  const loadMessages = useCallback(async () => {
    if (!token || !groupId) return
    const res = await apiRequest<{ messages: ChatMessage[] }>(`/chat/groups/${groupId}/messages`, {}, token)
    setMessages(res.messages)
  }, [token, groupId])

  useEffect(() => {
    if (!token || !groupId) return
    loadGroup().catch((err) => setError(err instanceof Error ? err.message : 'Erreur'))
    loadMessages().catch((err) => setError(err instanceof Error ? err.message : 'Erreur'))
  }, [token, groupId, loadGroup, loadMessages])

  useEffect(() => {
    if (!token || !groupId) return

    let closed = false
    let attempt = 0

    function connect() {
      if (closed || !token || !groupId) return
      const ws = new WebSocket(chatWebSocketURL(groupId, token))
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

  const candidatesToAdd = useMemo(() => {
    const inGroup = new Set(groupMembers.map((m) => m.user_id))
    return clubMembers.filter((m) => m.user_id !== user?.id && !inGroup.has(m.user_id))
  }, [clubMembers, groupMembers, user?.id])

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

  async function addMembers() {
    if (!token || !groupId || addMemberIds.length === 0 || busy) return
    setBusy(true)
    try {
      await apiRequest(`/chat/groups/${groupId}/members`, {
        method: 'POST',
        body: JSON.stringify({ member_ids: addMemberIds }),
      }, token)
      setAddMemberIds([])
      await loadGroup()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ajout impossible')
    } finally {
      setBusy(false)
    }
  }

  const backLink = activeClubId ? `/clubs/${activeClubId}/messages` : '/home'
  const title = group?.name ?? 'Discussion'
  const subtitle = connected ? 'En ligne' : 'Reconnexion…'

  return (
    <div className="wa-chat">
      <header className="wa-chat-header">
        <Link to={backLink} className="wa-chat-back" aria-label="Retour aux discussions">
          <ArrowLeft size={22} strokeWidth={2} />
        </Link>
        <ChatAvatar name={title} isGroup={!isDirect} size="md" />
        <div className="wa-chat-header-text">
          <h1>{title}</h1>
          <p className="muted small">{subtitle}</p>
        </div>
        <button
          type="button"
          className="wa-icon-btn"
          aria-label="Infos du groupe"
          onClick={() => setShowInfo(true)}
        >
          <Info size={22} strokeWidth={1.85} />
        </button>
      </header>

      {error && <p className="error wa-chat-error">{error}</p>}

      <div className="wa-chat-body">
        <div className="wa-chat-messages">
          {messages.map((m) => {
            const mine = m.user_id === user?.id
            return (
              <div key={m.id} className={`wa-bubble-row${mine ? ' mine' : ''}`}>
                <div className={`wa-bubble${mine ? ' mine' : ''}`}>
                  {!mine && <span className="wa-bubble-author">{m.user?.first_name ?? 'Membre'}</span>}
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
                  <time>
                    {new Date(m.created_at).toLocaleTimeString('fr-FR', { hour: '2-digit', minute: '2-digit' })}
                  </time>
                  {mine && editingId !== m.id && (
                    <div className="wa-bubble-actions">
                      <button type="button" className="btn-ghost btn-small" onClick={() => { setEditingId(m.id); setEditText(m.content) }}>
                        Modifier
                      </button>
                      <button type="button" className="btn-ghost btn-small" onClick={() => deleteMessage(m.id)}>
                        Supprimer
                      </button>
                    </div>
                  )}
                </div>
              </div>
            )
          })}
          <div ref={bottomRef} />
        </div>
      </div>

      <form onSubmit={send} className="wa-composer">
        <input
          value={content}
          onChange={(e) => setContent(e.target.value)}
          placeholder="Message"
          aria-label="Votre message"
          autoComplete="off"
        />
        <button type="submit" className="wa-send-btn" disabled={!content.trim()} aria-label="Envoyer">
          <Send size={20} strokeWidth={2} />
        </button>
      </form>

      {showInfo && (
        <div className="wa-sheet-backdrop" role="presentation" onClick={() => setShowInfo(false)}>
          <div
            className="wa-sheet wa-sheet--info"
            role="dialog"
            aria-modal="true"
            aria-labelledby="chat-info-title"
            onClick={(e) => e.stopPropagation()}
          >
            <header className="wa-sheet-header">
              <h2 id="chat-info-title">Détails</h2>
              <button type="button" className="wa-icon-btn" aria-label="Fermer" onClick={() => setShowInfo(false)}>
                <X size={22} />
              </button>
            </header>
            <div className="wa-info-hero">
              <ChatAvatar name={title} isGroup={!isDirect} size="lg" />
              <p className="wa-info-name">{title}</p>
              {group && (
                <p className="muted small">
                  {group.group_type === 'club' ? 'Discussion générale du club' : group.group_type === 'commission' ? 'Groupe de commission' : isDirect ? 'Conversation privée' : 'Groupe personnalisé'}
                </p>
              )}
            </div>
            <section className="wa-info-section">
              <h3>Participants ({groupMembers.length})</h3>
              <ul className="wa-member-pick-list">
                {groupMembers.map((m) => {
                  const name = `${m.user?.first_name ?? ''} ${m.user?.last_name ?? ''}`.trim() || 'Membre'
                  return (
                    <li key={m.id} className="wa-member-pick-row wa-member-pick-row--static">
                      <ChatAvatar name={name} size="md" />
                      <span>{name}</span>
                    </li>
                  )
                })}
              </ul>
            </section>
            {isCustomGroup && candidatesToAdd.length > 0 && (
              <section className="wa-info-section stack">
                <h3>Ajouter des membres</h3>
                <ul className="wa-member-pick-list wa-member-pick-list--scroll">
                  {candidatesToAdd.map((m) => {
                    const name = `${m.user?.first_name ?? ''} ${m.user?.last_name ?? ''}`.trim() || 'Membre'
                    const checked = addMemberIds.includes(m.user_id)
                    return (
                      <li key={m.id}>
                        <label className="wa-member-check-row">
                          <input
                            type="checkbox"
                            checked={checked}
                            onChange={() => {
                              setAddMemberIds((prev) => (
                                checked ? prev.filter((id) => id !== m.user_id) : [...prev, m.user_id]
                              ))
                            }}
                          />
                          <ChatAvatar name={name} size="md" />
                          <span>{name}</span>
                        </label>
                      </li>
                    )
                  })}
                </ul>
                <button
                  type="button"
                  className="btn-primary"
                  disabled={busy || addMemberIds.length === 0}
                  onClick={() => addMembers()}
                >
                  Ajouter au groupe
                </button>
              </section>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
