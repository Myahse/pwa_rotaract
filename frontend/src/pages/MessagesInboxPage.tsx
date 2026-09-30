import { useEffect, useMemo, useState, type FormEvent } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { MessageCirclePlus, Search, Users, X } from 'lucide-react'
import { apiRequest } from '../api/client'
import type { ChatInboxGroup, ClubMember } from '../api/types'
import { useAuth } from '../auth/AuthContext'
import { ChatAvatar } from '../components/ChatAvatar'

type Filter = 'all' | 'direct' | 'groups'
type Panel = null | 'new-dm' | 'new-group'

function formatPreview(content?: string | null) {
  if (!content) return 'Aucun message'
  return content.length > 56 ? `${content.slice(0, 56)}…` : content
}

function formatWhen(value?: string | null) {
  if (!value) return ''
  const date = new Date(value)
  const now = new Date()
  if (date.toDateString() === now.toDateString()) {
    return date.toLocaleTimeString('fr-FR', { hour: '2-digit', minute: '2-digit' })
  }
  const diffDays = Math.floor((now.getTime() - date.getTime()) / 86400000)
  if (diffDays < 7) {
    return date.toLocaleDateString('fr-FR', { weekday: 'short' })
  }
  return date.toLocaleDateString('fr-FR', { day: '2-digit', month: '2-digit' })
}

function groupTypeLabel(item: ChatInboxGroup): string | null {
  if (item.is_direct) return null
  if (item.group_type === 'club') return 'Général'
  if (item.group_type === 'commission') return 'Commission'
  return 'Groupe'
}

function sortInbox(items: ChatInboxGroup[]) {
  return [...items].sort((a, b) => {
    const ta = new Date(a.last_message_at ?? a.created_at).getTime()
    const tb = new Date(b.last_message_at ?? b.created_at).getTime()
    return tb - ta
  })
}

export function MessagesInboxPage() {
  const { clubId } = useParams<{ clubId: string }>()
  const { token, user, uiCaps } = useAuth()
  const navigate = useNavigate()
  const [filter, setFilter] = useState<Filter>('all')
  const [search, setSearch] = useState('')
  const [panel, setPanel] = useState<Panel>(null)
  const [pickerSearch, setPickerSearch] = useState('')
  const [inbox, setInbox] = useState<ChatInboxGroup[]>([])
  const [members, setMembers] = useState<ClubMember[]>([])
  const [error, setError] = useState('')
  const [groupForm, setGroupForm] = useState({ name: '', member_ids: [] as string[] })
  const [busy, setBusy] = useState(false)

  const groupsOnly = uiCaps.clubView === 'groups_only'
  const visibleInbox = useMemo(() => {
    if (!groupsOnly) return inbox
    const ids = new Set(uiCaps.presidentCommissionIds)
    return inbox.filter((item) => item.commission_id && ids.has(item.commission_id))
  }, [inbox, groupsOnly, uiCaps.presidentCommissionIds])

  const filteredList = useMemo(() => {
    let list = visibleInbox
    if (filter === 'direct') list = list.filter((item) => item.is_direct)
    if (filter === 'groups') list = list.filter((item) => !item.is_direct)
    const q = search.trim().toLowerCase()
    if (q) {
      list = list.filter((item) => {
        const hay = `${item.name} ${item.last_message_content ?? ''}`.toLowerCase()
        return hay.includes(q)
      })
    }
    return sortInbox(list)
  }, [visibleInbox, filter, search])

  const otherMembers = useMemo(
    () => members.filter((m) => m.user_id !== user?.id),
    [members, user?.id],
  )

  const pickerMembers = useMemo(() => {
    const q = pickerSearch.trim().toLowerCase()
    if (!q) return otherMembers
    return otherMembers.filter((m) => {
      const name = `${m.user?.first_name ?? ''} ${m.user?.last_name ?? ''}`.toLowerCase()
      return name.includes(q)
    })
  }, [otherMembers, pickerSearch])

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
      reload().catch(() => { /* ignore */ })
    }, 15000)
    return () => window.clearInterval(timer)
  }, [token, clubId])

  function closePanel() {
    setPanel(null)
    setPickerSearch('')
    setGroupForm({ name: '', member_ids: [] })
  }

  async function startDirect(memberId: string) {
    if (!token || !clubId || busy) return
    setBusy(true)
    try {
      const group = await apiRequest<{ id: string }>(`/clubs/${clubId}/chat/direct`, {
        method: 'POST',
        body: JSON.stringify({ user_id: memberId }),
      }, token)
      closePanel()
      navigate(`/chat/${group.id}`)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Conversation impossible')
    } finally {
      setBusy(false)
    }
  }

  async function createGroup(e: FormEvent) {
    e.preventDefault()
    if (!token || !clubId || busy) return
    setBusy(true)
    try {
      const group = await apiRequest<{ id: string }>(`/clubs/${clubId}/chat/groups`, {
        method: 'POST',
        body: JSON.stringify({
          name: groupForm.name.trim(),
          member_ids: groupForm.member_ids,
        }),
      }, token)
      closePanel()
      await reload()
      navigate(`/chat/${group.id}`)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Création impossible')
    } finally {
      setBusy(false)
    }
  }

  function toggleGroupMember(userId: string) {
    setGroupForm((prev) => {
      const has = prev.member_ids.includes(userId)
      return {
        ...prev,
        member_ids: has
          ? prev.member_ids.filter((id) => id !== userId)
          : [...prev.member_ids, userId],
      }
    })
  }

  const directCount = visibleInbox.filter((i) => i.is_direct).length
  const groupCount = visibleInbox.filter((i) => !i.is_direct).length

  return (
    <div className="wa-messages">
      <header className="wa-messages-toolbar">
        <div className="wa-messages-toolbar-top">
          <h1>Discussions</h1>
          <div className="wa-messages-actions">
            <button
              type="button"
              className="wa-icon-btn"
              aria-label="Nouvelle conversation"
              onClick={() => setPanel('new-dm')}
            >
              <MessageCirclePlus size={22} strokeWidth={1.85} />
            </button>
            <button
              type="button"
              className="wa-icon-btn"
              aria-label="Nouveau groupe"
              onClick={() => setPanel('new-group')}
            >
              <Users size={22} strokeWidth={1.85} />
            </button>
          </div>
        </div>
        <div className="wa-search">
          <Search className="wa-search-icon" size={18} aria-hidden="true" />
          <input
            type="search"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Rechercher une discussion…"
            aria-label="Rechercher"
          />
        </div>
        <div className="wa-filter-row" role="tablist" aria-label="Filtrer les discussions">
          <button
            type="button"
            role="tab"
            aria-selected={filter === 'all'}
            className={filter === 'all' ? 'wa-chip active' : 'wa-chip'}
            onClick={() => setFilter('all')}
          >
            Toutes
          </button>
          <button
            type="button"
            role="tab"
            aria-selected={filter === 'direct'}
            className={filter === 'direct' ? 'wa-chip active' : 'wa-chip'}
            onClick={() => setFilter('direct')}
          >
            Privées ({directCount})
          </button>
          <button
            type="button"
            role="tab"
            aria-selected={filter === 'groups'}
            className={filter === 'groups' ? 'wa-chip active' : 'wa-chip'}
            onClick={() => setFilter('groups')}
          >
            Groupes ({groupCount})
          </button>
        </div>
      </header>

      {error && <p className="error wa-messages-error">{error}</p>}

      <ul className="wa-chat-list">
        {filteredList.length === 0 ? (
          <li className="wa-chat-list-empty">
            <p className="muted">Aucune discussion.</p>
            <button type="button" className="btn-outline btn-small" onClick={() => setPanel('new-dm')}>
              Démarrer une conversation
            </button>
          </li>
        ) : (
          filteredList.map((item) => {
            const typeLabel = groupTypeLabel(item)
            return (
              <li key={item.id}>
                <Link to={`/chat/${item.id}`} className="wa-chat-row">
                  <ChatAvatar name={item.name} isGroup={!item.is_direct} />
                  <div className="wa-chat-row-body">
                    <div className="wa-chat-row-head">
                      <span className="wa-chat-row-title">{item.name}</span>
                      <time className="wa-chat-row-time">{formatWhen(item.last_message_at ?? item.created_at)}</time>
                    </div>
                    <div className="wa-chat-row-foot">
                      <p className="wa-chat-row-preview">{formatPreview(item.last_message_content)}</p>
                      {typeLabel && <span className="wa-chat-row-badge">{typeLabel}</span>}
                      {!item.is_direct && item.member_count > 0 && (
                        <span className="wa-chat-row-meta muted">{item.member_count} membres</span>
                      )}
                    </div>
                  </div>
                </Link>
              </li>
            )
          })
        )}
      </ul>

      {panel && (
        <div className="wa-sheet-backdrop" role="presentation" onClick={closePanel}>
          <div
            className="wa-sheet"
            role="dialog"
            aria-modal="true"
            aria-labelledby="wa-sheet-title"
            onClick={(e) => e.stopPropagation()}
          >
            <header className="wa-sheet-header">
              <h2 id="wa-sheet-title">
                {panel === 'new-dm' ? 'Nouvelle conversation' : 'Nouveau groupe'}
              </h2>
              <button type="button" className="wa-icon-btn" aria-label="Fermer" onClick={closePanel}>
                <X size={22} />
              </button>
            </header>

            <div className="wa-search wa-search--inset">
              <Search className="wa-search-icon" size={18} aria-hidden="true" />
              <input
                type="search"
                value={pickerSearch}
                onChange={(e) => setPickerSearch(e.target.value)}
                placeholder="Rechercher un membre…"
                aria-label="Rechercher un membre"
              />
            </div>

            {panel === 'new-dm' && (
              <ul className="wa-member-pick-list">
                {pickerMembers.length === 0 ? (
                  <li className="muted wa-member-pick-empty">Aucun membre trouvé.</li>
                ) : (
                  pickerMembers.map((m) => {
                    const name = `${m.user?.first_name ?? ''} ${m.user?.last_name ?? ''}`.trim() || 'Membre'
                    return (
                      <li key={m.id}>
                        <button
                          type="button"
                          className="wa-member-pick-row"
                          disabled={busy}
                          onClick={() => startDirect(m.user_id)}
                        >
                          <ChatAvatar name={name} size="md" />
                          <span>{name}</span>
                        </button>
                      </li>
                    )
                  })
                )}
              </ul>
            )}

            {panel === 'new-group' && (
              <form onSubmit={createGroup} className="wa-group-form stack">
                <label>
                  Nom du groupe
                  <input
                    value={groupForm.name}
                    onChange={(e) => setGroupForm({ ...groupForm, name: e.target.value })}
                    placeholder="Ex. Projet solidarité"
                    required
                    maxLength={80}
                  />
                </label>
                <p className="small muted">Membres ({groupForm.member_ids.length} sélectionné{groupForm.member_ids.length > 1 ? 's' : ''})</p>
                <ul className="wa-member-pick-list wa-member-pick-list--scroll">
                  {pickerMembers.map((m) => {
                    const name = `${m.user?.first_name ?? ''} ${m.user?.last_name ?? ''}`.trim() || 'Membre'
                    const checked = groupForm.member_ids.includes(m.user_id)
                    return (
                      <li key={m.id}>
                        <label className="wa-member-check-row">
                          <input
                            type="checkbox"
                            checked={checked}
                            onChange={() => toggleGroupMember(m.user_id)}
                          />
                          <ChatAvatar name={name} size="md" />
                          <span>{name}</span>
                        </label>
                      </li>
                    )
                  })}
                </ul>
                <button type="submit" className="btn-primary" disabled={busy || !groupForm.name.trim()}>
                  Créer le groupe
                </button>
              </form>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
