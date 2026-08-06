import { useCallback, useEffect, useState, type FormEvent } from 'react'
import { Link, useParams } from 'react-router-dom'
import { apiRequest } from '../api/client'
import type {
  SocialGroup,
  SocialGroupJoinRequest,
  SocialGroupMember,
  SocialGroupMessage,
} from '../api/types'
import { useAuth } from '../auth/AuthContext'
import { getToken } from '../auth/storage'
import { SkeletonList, SkeletonProfile, SkeletonThread } from '../components/Skeleton'

export function GroupDetailPage() {
  const { groupID } = useParams()
  const { user, openLogin, requireAuth } = useAuth()
  const [group, setGroup] = useState<SocialGroup | null>(null)
  const [members, setMembers] = useState<SocialGroupMember[]>([])
  const [requests, setRequests] = useState<SocialGroupJoinRequest[]>([])
  const [messages, setMessages] = useState<SocialGroupMessage[]>([])
  const [joinMessage, setJoinMessage] = useState('')
  const [chatBody, setChatBody] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)
  const [loadingChat, setLoadingChat] = useState(false)
  const [tab, setTab] = useState<'chat' | 'members' | 'requests'>('chat')

  const load = useCallback(async () => {
    if (!groupID) return
    setError('')
    try {
      const t = getToken()
      const [g, m] = await Promise.all([
        apiRequest<SocialGroup>(`/social/groups/${groupID}`, {}, t),
        apiRequest<{ members: SocialGroupMember[] }>(`/social/groups/${groupID}/members`, {}, t),
      ])
      setGroup(g)
      setMembers(m.members)
      if (g.my_role === 'admin') {
        const reqs = await apiRequest<{ requests: SocialGroupJoinRequest[] }>(
          `/social/groups/${groupID}/requests`,
          {},
          t,
        )
        setRequests(reqs.requests)
      } else {
        setRequests([])
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Groupe introuvable')
      setGroup((prev) => prev ?? null)
    } finally {
      setLoading(false)
    }
  }, [groupID])

  const loadChat = useCallback(async () => {
    if (!groupID || !user || group?.join_status !== 'member') {
      setMessages([])
      return
    }
    try {
      const data = await apiRequest<{ messages: SocialGroupMessage[] }>(
        `/social/groups/${groupID}/messages?limit=80`,
        {},
        getToken(),
      )
      setMessages(data.messages)
    } catch {
      /* keep existing */
    } finally {
      setLoadingChat(false)
    }
  }, [groupID, user, group?.join_status])

  useEffect(() => {
    setLoading(true)
    void load()
  }, [load])

  useEffect(() => {
    setLoadingChat(true)
    void loadChat()
  }, [loadChat])

  function join() {
    requireAuth(async () => {
      const g = await apiRequest<SocialGroup>(`/social/groups/${groupID}/join`, {
        method: 'POST',
        body: JSON.stringify({ message: joinMessage.trim() }),
      }, getToken())
      setGroup(g)
      setJoinMessage('')
      void load()
    }, 'Connectez-vous pour rejoindre ce groupe')
  }

  function leave() {
    if (!confirm('Quitter ce groupe ?')) return
    requireAuth(async () => {
      await apiRequest(`/social/groups/${groupID}/leave`, { method: 'DELETE' }, getToken())
      setGroup((prev) => prev ? { ...prev, join_status: 'none', my_role: undefined } : prev)
      setMessages([])
      void load()
    })
  }

  async function review(userId: string, approve: boolean) {
    const path = approve ? 'approve' : 'reject'
    setRequests((prev) => prev.filter((r) => r.user_id !== userId))
    try {
      await apiRequest(`/social/groups/${groupID}/requests/${userId}/${path}`, {
        method: 'POST',
        body: '{}',
      }, getToken())
      void load()
    } catch {
      void load()
    }
  }

  async function sendChat(e: FormEvent) {
    e.preventDefault()
    const body = chatBody.trim()
    if (!body || !user) return
    const optimistic: SocialGroupMessage = {
      id: `tmp-${crypto.randomUUID()}`,
      group_id: groupID!,
      sender_id: user.id,
      body,
      created_at: new Date().toISOString(),
      sender: {
        id: user.id,
        first_name: user.first_name,
        last_name: user.last_name,
        avatar_url: user.avatar_url,
      },
    }
    setChatBody('')
    setMessages((prev) => [...prev, optimistic])
    try {
      const created = await apiRequest<SocialGroupMessage>(`/social/groups/${groupID}/messages`, {
        method: 'POST',
        body: JSON.stringify({ body }),
      }, getToken())
      setMessages((prev) => prev.map((m) => (m.id === optimistic.id ? created : m)))
    } catch {
      setMessages((prev) => prev.filter((m) => m.id !== optimistic.id))
      setChatBody(body)
    }
  }

  if (loading && !group) {
    return (
      <div className="stack gap-md">
        <SkeletonProfile />
      </div>
    )
  }

  if (error || !group) {
    return (
      <div className="stack gap-md">
        <Link to="/groups" className="btn-text">← Groupes</Link>
        <p className="error">{error || 'Groupe introuvable'}</p>
      </div>
    )
  }

  const isMember = group.join_status === 'member'
  const isAdmin = group.my_role === 'admin'

  return (
    <div className="stack gap-md">
      <Link to="/groups" className="btn-text">← Groupes</Link>

      <section className="profile-hero panel">
        <div className="avatar-circle xl">{group.name[0]}</div>
        <div className="stack" style={{ gap: '0.35rem', flex: 1 }}>
          <h1>{group.name}</h1>
          <p className="muted small">
            {group.member_count} membres · {group.privacy === 'open' ? 'Ouvert' : 'Sur demande'}
            {isAdmin && group.pending_count ? ` · ${group.pending_count} demande(s)` : ''}
          </p>
          {group.description && <p>{group.description}</p>}
          <div className="actions">
            {!user && (
              <button type="button" className="btn-primary btn-tiny" onClick={() => openLogin()}>
                Se connecter
              </button>
            )}
            {user && !isMember && group.join_status !== 'pending' && (
              <>
                {group.privacy === 'approval' && (
                  <input
                    value={joinMessage}
                    onChange={(e) => setJoinMessage(e.target.value)}
                    placeholder="Message pour l’admin (optionnel)"
                    style={{ maxWidth: 220 }}
                  />
                )}
                <button type="button" className="btn-primary btn-tiny" onClick={join}>
                  {group.privacy === 'open' ? 'Rejoindre' : 'Demander à rejoindre'}
                </button>
              </>
            )}
            {group.join_status === 'pending' && <span className="chip">Demande en attente</span>}
            {isMember && (
              <button type="button" className="chip" onClick={leave}>Quitter</button>
            )}
          </div>
        </div>
      </section>

      {isMember && (
        <div className="feed-tabs">
          <button type="button" className={tab === 'chat' ? 'active' : ''} onClick={() => setTab('chat')}>Discussion</button>
          <button type="button" className={tab === 'members' ? 'active' : ''} onClick={() => setTab('members')}>
            Membres ({members.length})
          </button>
          {isAdmin && (
            <button type="button" className={tab === 'requests' ? 'active' : ''} onClick={() => setTab('requests')}>
              Demandes ({requests.length})
            </button>
          )}
        </div>
      )}

      {isMember && tab === 'chat' && (
        <section className="dm-panel">
          {loadingChat && messages.length === 0 ? (
            <SkeletonThread />
          ) : (
            <div className="dm-thread">
              {messages.length === 0 && <p className="muted">Aucun message. Lancez la conversation.</p>}
              {messages.map((m) => (
                <div key={m.id} className={m.sender_id === user?.id ? 'dm-bubble mine' : 'dm-bubble'}>
                  {m.sender_id !== user?.id && (
                    <strong className="small">{m.sender?.first_name} {m.sender?.last_name}</strong>
                  )}
                  <p>{m.body}</p>
                  <time className="muted small">{new Date(m.created_at).toLocaleString('fr-FR')}</time>
                </div>
              ))}
            </div>
          )}
          <form onSubmit={sendChat} className="comment-form">
            <input
              value={chatBody}
              onChange={(e) => setChatBody(e.target.value)}
              placeholder="Écrire dans le groupe…"
            />
            <button type="submit" className="btn-primary btn-tiny">Envoyer</button>
          </form>
        </section>
      )}

      {(!isMember || tab === 'members') && (
        <section className="panel stack">
          <h2>Membres</h2>
          {members.length === 0 ? <SkeletonList count={3} /> : (
            <ul className="suggest-list">
              {members.map((m) => (
                <li key={m.user_id}>
                  <Link to={`/users/${m.user_id}`} className="suggest-user">
                    <div className="avatar-circle">
                      {m.user?.avatar_url ? <img src={m.user.avatar_url} alt="" /> : m.user?.first_name?.[0]}
                    </div>
                    <div>
                      <strong>{m.user?.first_name} {m.user?.last_name}</strong>
                      <p className="muted small">{m.role === 'admin' ? 'Admin' : 'Membre'}</p>
                    </div>
                  </Link>
                </li>
              ))}
            </ul>
          )}
        </section>
      )}

      {isAdmin && tab === 'requests' && (
        <section className="panel stack">
          <h2>Demandes d&apos;adhésion</h2>
          {requests.length === 0 && <p className="muted">Aucune demande en attente.</p>}
          <ul className="suggest-list">
            {requests.map((r) => (
              <li key={r.id}>
                <Link to={`/users/${r.user_id}`} className="suggest-user">
                  <div className="avatar-circle">
                    {r.user?.avatar_url ? <img src={r.user.avatar_url} alt="" /> : r.user?.first_name?.[0]}
                  </div>
                  <div>
                    <strong>{r.user?.first_name} {r.user?.last_name}</strong>
                    {r.message && <p className="muted small">{r.message}</p>}
                  </div>
                </Link>
                <div className="actions">
                  <button type="button" className="btn-primary btn-tiny" onClick={() => review(r.user_id, true)}>Accepter</button>
                  <button type="button" className="chip" onClick={() => review(r.user_id, false)}>Refuser</button>
                </div>
              </li>
            ))}
          </ul>
        </section>
      )}

      {!isMember && group.privacy === 'approval' && (
        <p className="muted small">La discussion de groupe est réservée aux membres. Envoyez une demande à un admin.</p>
      )}
    </div>
  )
}
