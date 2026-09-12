import { useEffect, useState, type FormEvent } from 'react'
import { Link, useParams } from 'react-router-dom'
import { apiRequest } from '../api/client'
import type {
  AccessRequest,
  ChatGroup,
  ClubMember,
  ClubRole,
  ClubRoleAssignment,
  Commission,
  CommissionMember,
  EmailInvite,
} from '../api/types'
import { ClubDiaryPanel } from '../components/ClubDiaryPanel'
import { useAuth } from '../auth/AuthContext'

export function HeadManagePage() {
  const { clubId } = useParams<{ clubId: string }>()
  const { token, uiCaps } = useAuth()
  const sections = uiCaps.manageSections

  const [members, setMembers] = useState<ClubMember[]>([])
  const [roles, setRoles] = useState<ClubRole[]>([])
  const [invites, setInvites] = useState<EmailInvite[]>([])
  const [commissions, setCommissions] = useState<Commission[]>([])
  const [requests, setRequests] = useState<AccessRequest[]>([])
  const [assignments, setAssignments] = useState<ClubRoleAssignment[]>([])
  const [groups, setGroups] = useState<ChatGroup[]>([])
  const [selectedCommission, setSelectedCommission] = useState<string>('')
  const [roleForm, setRoleForm] = useState({ name: '', description: '' })
  const [groupForm, setGroupForm] = useState({ name: '', member_ids: [] as string[] })
  const [rolePick, setRolePick] = useState<Record<string, string>>({})
  const [commissionMembers, setCommissionMembers] = useState<CommissionMember[]>([])

  const [inviteEmail, setInviteEmail] = useState('')
  const [commissionForm, setCommissionForm] = useState({ name: '', description: '' })
  const [editCommission, setEditCommission] = useState({ name: '', description: '' })
  const [addMemberForm, setAddMemberForm] = useState({ user_id: '', member_role: 'member' as 'president' | 'secretary' | 'member' })
  const [error, setError] = useState('')
  const [message, setMessage] = useState('')

  async function reload() {
    if (!token || !clubId) return
    const tasks: Promise<void>[] = []
    if (sections.members || sections.groups || sections.commissions) {
      tasks.push(apiRequest<ClubMember[]>(`/clubs/${clubId}/members`, {}, token).then(setMembers))
    }
    if (sections.roles || sections.members) {
      tasks.push(apiRequest<ClubRole[]>(`/clubs/${clubId}/roles`, {}, token).then(setRoles))
    }
    if (sections.invites) {
      tasks.push(apiRequest<EmailInvite[]>(`/clubs/${clubId}/email-invites`, {}, token).then(setInvites))
    }
    if (sections.commissions) {
      tasks.push(apiRequest<Commission[]>(`/clubs/${clubId}/commissions`, {}, token).then(setCommissions))
    }
    if (sections.accessRequests) {
      tasks.push(apiRequest<AccessRequest[]>(`/clubs/${clubId}/access-requests`, {}, token).then(setRequests))
    }
    if (sections.members && sections.roles) {
      tasks.push(apiRequest<ClubRoleAssignment[]>(`/clubs/${clubId}/role-assignments`, {}, token).then(setAssignments))
    }
    if (sections.groups) {
      tasks.push(apiRequest<ChatGroup[]>(`/clubs/${clubId}/chat/groups`, {}, token).then(setGroups))
    }
    await Promise.all(tasks)
  }

  useEffect(() => {
    if (!uiCaps.nav.manage) return
    reload().catch((err) => setError(err instanceof Error ? err.message : 'Erreur'))
  }, [token, clubId, uiCaps.nav.manage, sections.members, sections.invites, sections.roles, sections.groups, sections.commissions, sections.accessRequests])

  useEffect(() => {
    if (!token || !clubId || !selectedCommission) return
    apiRequest<{ commission: Commission; members: CommissionMember[] }>(
      `/clubs/${clubId}/commissions/${selectedCommission}`, {}, token,
    ).then((res) => {
      setCommissionMembers(res.members)
      setEditCommission({ name: res.commission.name, description: res.commission.description ?? '' })
    }).catch(() => setCommissionMembers([]))
  }, [token, clubId, selectedCommission])

  if (!uiCaps.nav.manage) {
    return <p className="error">Accès non autorisé.</p>
  }

  async function sendInvite(e: FormEvent) {
    e.preventDefault()
    if (!token || !clubId) return
    setError('')
    try {
      await apiRequest(`/clubs/${clubId}/email-invites`, {
        method: 'POST',
        body: JSON.stringify({ email: inviteEmail }),
      }, token)
      setInviteEmail('')
      setMessage('Invitation envoyée')
      await reload()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erreur')
    }
  }

  async function revokeInvite(id: string) {
    if (!token || !clubId) return
    await apiRequest(`/clubs/${clubId}/email-invites/${id}`, { method: 'DELETE' }, token)
    await reload()
  }

  async function createCommission(e: FormEvent) {
    e.preventDefault()
    if (!token || !clubId) return
    await apiRequest(`/clubs/${clubId}/commissions`, {
      method: 'POST',
      body: JSON.stringify(commissionForm),
    }, token)
    setCommissionForm({ name: '', description: '' })
    await reload()
  }

  async function updateCommission(e: FormEvent) {
    e.preventDefault()
    if (!token || !clubId || !selectedCommission) return
    await apiRequest(`/clubs/${clubId}/commissions/${selectedCommission}`, {
      method: 'PATCH',
      body: JSON.stringify(editCommission),
    }, token)
    setMessage('Commission mise à jour')
    await reload()
  }

  async function deleteCommission(id: string) {
    if (!token || !clubId) return
    const target = commissions.find((c) => c.id === id)
    if (target?.is_system) {
      setError('Cette commission système ne peut pas être supprimée')
      return
    }
    if (!confirm('Supprimer cette commission ?')) return
    await apiRequest(`/clubs/${clubId}/commissions/${id}`, { method: 'DELETE' }, token)
    if (selectedCommission === id) setSelectedCommission('')
    await reload()
  }

  async function addCommissionMember(e: FormEvent) {
    e.preventDefault()
    if (!token || !clubId || !selectedCommission) return
    await apiRequest(`/clubs/${clubId}/commissions/${selectedCommission}/members`, {
      method: 'POST',
      body: JSON.stringify(addMemberForm),
    }, token)
    setAddMemberForm({ user_id: '', member_role: 'member' })
    const res = await apiRequest<{ commission: Commission; members: CommissionMember[] }>(
      `/clubs/${clubId}/commissions/${selectedCommission}`, {}, token,
    )
    setCommissionMembers(res.members)
  }

  async function removeCommissionMember(userId: string) {
    if (!token || !clubId || !selectedCommission) return
    await apiRequest(`/clubs/${clubId}/commissions/${selectedCommission}/members/${userId}`, {
      method: 'DELETE',
    }, token)
    const res = await apiRequest<{ commission: Commission; members: CommissionMember[] }>(
      `/clubs/${clubId}/commissions/${selectedCommission}`, {}, token,
    )
    setCommissionMembers(res.members)
  }

  async function approveRequest(id: string) {
    if (!token || !clubId) return
    const membre = roles.find((r) => r.name === 'Membre')
    await apiRequest(`/clubs/${clubId}/access-requests/${id}/approve`, {
      method: 'POST',
      body: JSON.stringify({ role_id: membre?.id }),
    }, token)
    await reload()
  }

  async function createRole(e: FormEvent) {
    e.preventDefault()
    if (!token || !clubId) return
    await apiRequest(`/clubs/${clubId}/roles`, {
      method: 'POST',
      body: JSON.stringify({ ...roleForm, permission_keys: [] }),
    }, token)
    setRoleForm({ name: '', description: '' })
    setMessage('Rôle créé')
    await reload()
  }

  async function assignRole(userId: string) {
    if (!token || !clubId) return
    const roleId = rolePick[userId]
    if (!roleId) return
    await apiRequest(`/clubs/${clubId}/members/${userId}/roles`, {
      method: 'POST',
      body: JSON.stringify({ role_id: roleId }),
    }, token)
    setMessage('Rôle assigné')
    await reload()
  }

  async function unassignRole(userId: string, roleId: string) {
    if (!token || !clubId) return
    await apiRequest(`/clubs/${clubId}/members/${userId}/roles/${roleId}`, { method: 'DELETE' }, token)
    await reload()
  }

  async function createGroup(e: FormEvent) {
    e.preventDefault()
    if (!token || !clubId) return
    await apiRequest(`/clubs/${clubId}/chat/groups`, {
      method: 'POST',
      body: JSON.stringify({
        name: groupForm.name,
        member_ids: groupForm.member_ids,
      }),
    }, token)
    setGroupForm({ name: '', member_ids: [] })
    setMessage('Groupe créé')
    await reload()
  }

  async function rejectRequest(id: string) {
    if (!token || !clubId) return
    await apiRequest(`/clubs/${clubId}/access-requests/${id}/reject`, {
      method: 'POST',
      body: JSON.stringify({}),
    }, token)
    await reload()
  }

  const pendingInvites = invites.filter((i) => !i.used_at)
  const pendingRequests = requests.filter((r) => r.status === 'pending')
  const rolesByUser = assignments.reduce<Record<string, ClubRoleAssignment[]>>((acc, item) => {
    acc[item.user_id] = acc[item.user_id] ?? []
    acc[item.user_id].push(item)
    return acc
  }, {})

  return (
    <div className="stack gap-lg">
      {uiCaps.nav.club ? (
        <Link to={`/clubs/${clubId}`} className="back">← Retour au club</Link>
      ) : (
        <Link to="/home" className="back">← Retour à l'accueil</Link>
      )}
      <h2>{uiCaps.profile === 'secretary' ? 'Secrétariat du club' : 'Gestion du club'}</h2>
      {error && <p className="error">{error}</p>}
      {message && <p className="success">{message}</p>}

      {sections.members && (
      <section className="card stack">
        <h3>Membres ({members.length})</h3>
        <ul className="list">
          {members.map((m) => (
            <li key={m.id} className="stack gap-sm">
              <div>
                <strong>{m.user?.first_name} {m.user?.last_name}</strong>
                <span className="muted small"> · {m.member_role === 'head' ? 'Responsable' : 'Membre'}</span>
              </div>
              {sections.roles && (
                <>
                  <div className="role-tags">
                    {(rolesByUser[m.user_id] ?? []).map((item) => (
                      <span key={item.id} className="role-tag">
                        {item.role?.name ?? 'Rôle'}
                        <button type="button" className="btn-ghost btn-small" onClick={() => unassignRole(m.user_id, item.club_role_id)}>×</button>
                      </span>
                    ))}
                  </div>
                  {m.member_role !== 'head' && (
                    <div className="actions">
                      <select value={rolePick[m.user_id] ?? ''} onChange={(e) => setRolePick((prev) => ({ ...prev, [m.user_id]: e.target.value }))}>
                        <option value="">Assigner un rôle…</option>
                        {roles.filter((role) => role.name !== 'Responsable de club').map((role) => (
                          <option key={role.id} value={role.id}>{role.name}</option>
                        ))}
                      </select>
                      <button type="button" className="btn-small" onClick={() => assignRole(m.user_id)}>Assigner</button>
                    </div>
                  )}
                </>
              )}
            </li>
          ))}
        </ul>
      </section>
      )}

      {sections.roles && (
      <section className="card stack">
        <h3>Rôles du club</h3>
        <p className="muted small">Les rôles déterminent les permissions et l’interface de chaque membre.</p>
        <form onSubmit={createRole} className="stack">
          <label>Nom<input value={roleForm.name} onChange={(e) => setRoleForm({ ...roleForm, name: e.target.value })} required /></label>
          <label>Description<input value={roleForm.description} onChange={(e) => setRoleForm({ ...roleForm, description: e.target.value })} /></label>
          <button type="submit" className="btn-primary">Créer un rôle</button>
        </form>
        <ul className="list">
          {roles.map((role) => (
            <li key={role.id}>{role.name}{role.description ? ` — ${role.description}` : ''}</li>
          ))}
        </ul>
      </section>
      )}

      {sections.groups && (
      <section className="card stack">
        <h3>Groupes de discussion</h3>
        <form onSubmit={createGroup} className="stack">
          <label>Nom du groupe<input value={groupForm.name} onChange={(e) => setGroupForm({ ...groupForm, name: e.target.value })} required /></label>
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
              {members.map((m) => (
                <option key={m.user_id} value={m.user_id}>{m.user?.first_name} {m.user?.last_name}</option>
              ))}
            </select>
          </label>
          <button type="submit" className="btn-primary">Créer le groupe</button>
        </form>
        <ul className="list links">
          {groups.map((g) => (
            <li key={g.id}>
              <Link to={`/chat/${g.id}`}>{g.name}</Link>
              <span className="badge">{g.group_type}</span>
            </li>
          ))}
        </ul>
      </section>
      )}

      {sections.invites && (
      <section className="card stack">
        <h3>Invitations email</h3>
        <form onSubmit={sendInvite} className="stack">
          <label>Email<input type="email" value={inviteEmail} onChange={(e) => setInviteEmail(e.target.value)} required /></label>
          <button type="submit" className="btn-primary">Envoyer</button>
        </form>
        <ul className="list">
          {pendingInvites.map((i) => (
            <li key={i.id} className="row-between">
              <span>{i.email}</span>
              <button type="button" className="btn-ghost btn-small" onClick={() => revokeInvite(i.id)}>Révoquer</button>
            </li>
          ))}
        </ul>
      </section>
      )}

      {sections.commissions && (
      <section className="card stack">
        <h3>Commissions</h3>
        <p className="muted">Chaque commission a un président, un secrétaire (choisi par le président) et des membres.</p>
        <form onSubmit={createCommission} className="stack">
          <label>Nom<input value={commissionForm.name} onChange={(e) => setCommissionForm({ ...commissionForm, name: e.target.value })} required /></label>
          <label>Description<input value={commissionForm.description} onChange={(e) => setCommissionForm({ ...commissionForm, description: e.target.value })} /></label>
          <button type="submit" className="btn-primary">Créer</button>
        </form>
        <ul className="list">
          {commissions.map((c) => (
            <li key={c.id} className="row-between">
              <button type="button" className="btn-ghost" onClick={() => setSelectedCommission(c.id)}>
                {c.name}{c.code ? ` (${c.code})` : ''}{c.is_system && ' · système'}
              </button>
              {!c.is_system && (
                <button type="button" className="btn-ghost btn-small" onClick={() => deleteCommission(c.id)}>Supprimer</button>
              )}
            </li>
          ))}
        </ul>
        {selectedCommission && (
          <div className="stack nested">
            {commissions.find((c) => c.id === selectedCommission)?.description && (
              <p className="muted">{commissions.find((c) => c.id === selectedCommission)?.description}</p>
            )}
            <form onSubmit={updateCommission} className="stack">
              <label>Nom<input value={editCommission.name} onChange={(e) => setEditCommission({ ...editCommission, name: e.target.value })} /></label>
              <label>Description<input value={editCommission.description} onChange={(e) => setEditCommission({ ...editCommission, description: e.target.value })} /></label>
              <button type="submit" className="btn-primary">Mettre à jour</button>
            </form>
            <form onSubmit={addCommissionMember} className="stack">
              <label>Membre
                <select value={addMemberForm.user_id} onChange={(e) => setAddMemberForm({ ...addMemberForm, user_id: e.target.value })} required>
                  <option value="">Choisir…</option>
                  {members.map((m) => (
                    <option key={m.user_id} value={m.user_id}>{m.user?.first_name} {m.user?.last_name}</option>
                  ))}
                </select>
              </label>
              <label>Rôle
                <select value={addMemberForm.member_role} onChange={(e) => setAddMemberForm({ ...addMemberForm, member_role: e.target.value as 'president' | 'secretary' | 'member' })}>
                  <option value="member">Membre</option>
                  <option value="secretary">Secrétaire</option>
                  <option value="president">Président</option>
                </select>
              </label>
              <button type="submit" className="btn-primary">Ajouter</button>
            </form>
            <ul className="list">
              {commissionMembers.map((m) => (
                <li key={m.id} className="row-between">
                  <span>{m.user?.first_name} {m.user?.last_name} ({m.member_role === 'president' ? 'Président' : m.member_role === 'secretary' ? 'Secrétaire' : 'Membre'})</span>
                  <button type="button" className="btn-ghost btn-small" onClick={() => removeCommissionMember(m.user_id)}>Retirer</button>
                </li>
              ))}
            </ul>
          </div>
        )}
      </section>
      )}

      {sections.accessRequests && (
      <section className="card stack">
        <h3>Demandes d'accès ({pendingRequests.length})</h3>
        <ul className="list">
          {pendingRequests.map((r) => (
            <li key={r.id} className="stack">
              <span>{r.first_name} {r.last_name} — {r.email}</span>
              <div className="actions">
                <button type="button" className="btn-small" onClick={() => approveRequest(r.id)}>Approuver</button>
                <button type="button" className="btn-ghost btn-small" onClick={() => rejectRequest(r.id)}>Rejeter</button>
              </div>
            </li>
          ))}
        </ul>
      </section>
      )}

      {sections.diary && token && clubId && <ClubDiaryPanel clubId={clubId} token={token} />}
    </div>
  )
}
