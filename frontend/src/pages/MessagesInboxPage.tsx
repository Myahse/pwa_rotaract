import { useEffect, useMemo, useState, type FormEvent } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { apiRequest } from '../api/client'
import type { ChatInboxGroup, ClubMember } from '../api/types'
import { useAuth } from '../auth/AuthContext'

type Tab = 'direct' | 'groups'

function formatPreview(content?: string | null) {
  if (!content) return 'Aucun message'
  return content.length > 72 ? `${content.slice(0, 72)}…` : content
}

function formatWhen(value?: string | null) {
  if (!value) return ''
  const date = new Date(value)
  const now = new Date()
  if (date.toDateString() === now.toDateString()) {
    return date.toLocaleTimeString('fr-FR', { hour: '2-digit', minute: '2-digit' })
  }
  return date.toLocaleDateString('fr-FR', { day: '2-digit', month: 'short' })
}

export function MessagesInboxPage() {
  const { clubId } = useParams<{ clubId: string }>()
  const { token, user, uiCaps } = useAuth()
  const navigate = useNavigate()
  const [tab, setTab] = useState<Tab>('direct')
  const [inbox, setInbox] = useState<ChatInboxGroup[]>([])
  const [members, setMembers] = useState<ClubMember[]>([])
  const [error, setError] = useState('')
  const [showNewGroup, setShowNewGroup] = useState(false)
  const [groupForm, setGroupForm] = useState({ name: '', member_ids: [] as string[] })

  const groupsOnly = uiCaps.clubView === 'groups_only'
  const visibleInbox = useMemo(() => {
    if (!groupsOnly) return inbox
    const ids = new Set(uiCaps.presidentCommissionIds)
    return inbox.filter((item) => item.commission_id && ids.has(item.commission_id))
  }, [inbox, groupsOnly, uiCaps.presidentCommissionIds])

  const directChats = useMemo(
    () => visibleInbox.filter((item) => item.is_direct),
    [visibleInbox],
  )
  const groupChats = useMemo(
    () => visibleInbox.filter((item) => !item.is_direct),
    [visibleInbox],
  )
  const list = tab === 'direct' ? directChats : groupChats

  const otherMembers = useMemo(
    () => members.filter((m) => m.user_id !== user?.id),
    [members, user?.id],
  )

  async function reload() {
    if (!token || !clubId) return
    const [inboxRes, membersRes] = await Promise.all([
      apiRequest<ChatInboxGroup[]>(`/clubs/${clubId}/chat/inbox`, {}, token),
      apiRequest<ClubMember[]>(`/clubs/${clubId}/members`, {}, token),
    ])
    setInbox(inboxRes)
    setMembers(membersRes)
  }

  useEffect(() => {
    reload().catch((err) => setError(err instanceof Error ? err.message : 'Chargement impossible'))
    const timer = window.setInterval(() => {
      reload().catch(() => { /* ignore background refresh errors */ })
    }, 15000)
    return () => window.clearInterval(timer)
  }, [token, clubId])

  async function startDirect(memberId: string) {
    if (!token || !clubId) return
    try {
      const group = await apiRequest<{ id: string }>(`/clubs/${clubId}/chat/direct`, {
        method: 'POST',
        body: JSON.stringify({ user_id: memberId }),
      }, token)
      navigate(`/chat/${group.id}`)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Conversation impossible')
    }
  }

  async function createGroup(e: FormEvent) {
    e.preventDefault()
    if (!token || !clubId) return
    try {
      const group = await apiRequest<{ id: string }>(`/clubs/${clubId}/chat/groups`, {
        method: 'POST',
        body: JSON.stringify({
          name: groupForm.name,
          member_ids: groupForm.member_ids,
        }),
      }, token)
      setGroupForm({ name: '', member_ids: [] })
      setShowNewGroup(false)
      await reload()
      navigate(`/chat/${group.id}`)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Création impossible')
    }
  }

  return (
    <div className="stack gap-lg">
      <section>
        <h2>Messages</h2>
        <p className="muted">Conversations privées et groupes du club.</p>
      </section>

      {error && <p className="error">{error}</p>}

      <div className="tab-row">
        <button type="button" className={tab === 'direct' ? 'tab active' : 'tab'} onClick={() => setTab('direct')}>
          Messages ({directChats.length})
        </button>
        <button type="button" className={tab === 'groups' ? 'tab active' : 'tab'} onClick={() => setTab('groups')}>
          Groupes ({groupChats.length})
        </button>
      </div>

      {tab === 'direct' && (
        <section className="card stack">
          <h3>Nouvelle conversation</h3>
          {otherMembers.length === 0 ? (
            <p className="muted">Aucun autre membre disponible.</p>
          ) : (
            <ul className="list">
              {otherMembers.map((m) => (
                <li key={m.id} className="inbox-row">
                  <span>{m.user?.first_name} {m.user?.last_name}</span>
                  <button type="button" className="btn-small" onClick={() => startDirect(m.user_id)}>Message</button>
                </li>
              ))}
            </ul>
          )}
        </section>
      )}

      {tab === 'groups' && (
        <section className="card stack">
          <div className="actions">
            <h3>Créer un groupe</h3>
            <button type="button" className="btn-outline btn-small" onClick={() => setShowNewGroup((v) => !v)}>
              {showNewGroup ? 'Annuler' : 'Nouveau groupe'}
            </button>
          </div>
          {showNewGroup && (
            <form onSubmit={createGroup} className="stack">
              <label>
                Nom du groupe
                <input
                  value={groupForm.name}
                  onChange={(e) => setGroupForm({ ...groupForm, name: e.target.value })}
                  required
                />
              </label>
              <label>
                Membres
                <select
                  multiple
                  value={groupForm.member_ids}
                  onChange={(e) => setGroupForm({
                    ...groupForm,
                    member_ids: Array.from(e.target.selectedOptions).map((opt) => opt.value),
                  })}
                >
                  {otherMembers.map((m) => (
                    <option key={m.user_id} value={m.user_id}>
                      {m.user?.first_name} {m.user?.last_name}
                    </option>
                  ))}
                </select>
              </label>
              <button type="submit" className="btn-primary">Créer et ouvrir</button>
            </form>
          )}
        </section>
      )}

      <section className="card stack">
        <h3>{tab === 'direct' ? 'Conversations' : 'Groupes'}</h3>
        {list.length === 0 ? (
          <p className="muted">Aucune conversation pour le moment.</p>
        ) : (
          <ul className="list links inbox-list">
            {list.map((item) => (
              <li key={item.id}>
                <Link to={`/chat/${item.id}`} className="inbox-item">
                  <div className="inbox-item-head">
                    <strong>{item.name}</strong>
                    <span className="muted small">{formatWhen(item.last_message_at ?? item.created_at)}</span>
                  </div>
                  <p className="muted small inbox-preview">{formatPreview(item.last_message_content)}</p>
                  {!item.is_direct && (
                    <span className="badge">{item.group_type === 'club' ? 'Général' : item.group_type}</span>
                  )}
                </Link>
              </li>
            ))}
          </ul>
        )}
      </section>
    </div>
  )
}
