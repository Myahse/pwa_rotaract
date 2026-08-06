import { useEffect, useState, type FormEvent } from 'react'
import { Link, useParams } from 'react-router-dom'
import { apiRequest } from '../api/client'
import type {
  AccessRequest,
  ClubMember,
  ClubRole,
  Commission,
  CommissionMember,
  EmailInvite,
} from '../api/types'
import { ClubDiaryPanel } from '../components/ClubDiaryPanel'
import { useAuth } from '../auth/AuthContext'

export function HeadManagePage() {
  const { clubId } = useParams<{ clubId: string }>()
  const { token, clubs } = useAuth()
  const membership = clubs.find((c) => c.club_id === clubId)
  const isHead = membership?.member_role === 'head'

  const [members, setMembers] = useState<ClubMember[]>([])
  const [roles, setRoles] = useState<ClubRole[]>([])
  const [invites, setInvites] = useState<EmailInvite[]>([])
  const [commissions, setCommissions] = useState<Commission[]>([])
  const [requests, setRequests] = useState<AccessRequest[]>([])
  const [selectedCommission, setSelectedCommission] = useState<string>('')
  const [commissionMembers, setCommissionMembers] = useState<CommissionMember[]>([])

  const [inviteEmail, setInviteEmail] = useState('')
  const [commissionForm, setCommissionForm] = useState({ name: '', description: '' })
  const [editCommission, setEditCommission] = useState({ name: '', description: '' })
  const [addMemberForm, setAddMemberForm] = useState({ user_id: '', member_role: 'member' as 'president' | 'secretary' | 'member' })
  const [error, setError] = useState('')
  const [message, setMessage] = useState('')

  async function reload() {
    if (!token || !clubId) return
    const [m, r, i, c, req] = await Promise.all([
      apiRequest<ClubMember[]>(`/clubs/${clubId}/members`, {}, token),
      apiRequest<ClubRole[]>(`/clubs/${clubId}/roles`, {}, token),
      apiRequest<EmailInvite[]>(`/clubs/${clubId}/email-invites`, {}, token),
      apiRequest<Commission[]>(`/clubs/${clubId}/commissions`, {}, token),
      apiRequest<AccessRequest[]>(`/clubs/${clubId}/access-requests`, {}, token),
    ])
    setMembers(m)
    setRoles(r)
    setInvites(i)
    setCommissions(c)
    setRequests(req)
  }

  useEffect(() => {
    if (!isHead) return
    reload().catch((err) => setError(err instanceof Error ? err.message : 'Erreur'))
  }, [token, clubId, isHead])

  useEffect(() => {
    if (!token || !clubId || !selectedCommission) return
    apiRequest<{ commission: Commission; members: CommissionMember[] }>(
      `/clubs/${clubId}/commissions/${selectedCommission}`, {}, token,
    ).then((res) => {
      setCommissionMembers(res.members)
      setEditCommission({ name: res.commission.name, description: res.commission.description ?? '' })
    }).catch(() => setCommissionMembers([]))
  }, [token, clubId, selectedCommission])

  if (!isHead) {
    return <p className="error">Réservé au responsable du club.</p>
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

  return (
    <div className="stack gap-lg">
      <Link to={`/clubs/${clubId}`} className="back">← Retour au club</Link>
      <h2>Gestion du club</h2>
      {error && <p className="error">{error}</p>}
      {message && <p className="success">{message}</p>}

      <section className="card stack">
        <h3>Membres ({members.length})</h3>
        <ul className="list">
          {members.map((m) => (
            <li key={m.id}>
              {m.user?.first_name} {m.user?.last_name} — {m.member_role}
            </li>
          ))}
        </ul>
      </section>

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

      {token && clubId && <ClubDiaryPanel clubId={clubId} token={token} />}
    </div>
  )
}
