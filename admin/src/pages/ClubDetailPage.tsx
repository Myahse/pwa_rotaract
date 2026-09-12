import { useEffect, useState, type FormEvent } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { apiRequest } from '../api/client'
import type { Club, ClubMember } from '../api/types'
import { ClubDiaryPanel } from '../components/ClubDiaryPanel'
import { useAuth } from '../auth/AuthContext'

function toISODate(value: string) {
  return value ? `${value}T00:00:00.000Z` : undefined
}

export function ClubDetailPage() {
  const { clubId } = useParams<{ clubId: string }>()
  const navigate = useNavigate()
  const { token } = useAuth()
  const [club, setClub] = useState<Club | null>(null)
  const [editForm, setEditForm] = useState({
    name: '', slug: '', country: '', city: '', commune: '', description: '', founded_at: '', is_active: true,
  })
  const [headForm, setHeadForm] = useState({
    email: '', password: '', first_name: '', last_name: '', phone: '', birth_date: '', profession: '', member_since: '',
  })
  const [message, setMessage] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [members, setMembers] = useState<ClubMember[]>([])
  const [membersLoading, setMembersLoading] = useState(false)
  const [section, setSection] = useState<'info' | 'head' | 'members' | 'diary'>('info')

  useEffect(() => {
    if (!token || !clubId) return
    apiRequest<Club>(`/clubs/${clubId}/`, {}, token)
      .then((c) => {
        setClub(c)
        setEditForm({
          name: c.name,
          slug: c.slug,
          country: c.country ?? '',
          city: c.city ?? '',
          commune: c.commune ?? '',
          description: c.description ?? '',
          founded_at: c.founded_at ? c.founded_at.slice(0, 10) : '',
          is_active: c.is_active,
        })
      })
      .catch((err) => setError(err instanceof Error ? err.message : 'Erreur'))
  }, [token, clubId])

  useEffect(() => {
    if (!token || !clubId) return
    setMembersLoading(true)
    apiRequest<ClubMember[]>(`/clubs/${clubId}/members`, {}, token)
      .then(setMembers)
      .catch((err) => setError(err instanceof Error ? err.message : 'Impossible de charger les membres'))
      .finally(() => setMembersLoading(false))
  }, [token, clubId])

  async function onUpdateClub(e: FormEvent) {
    e.preventDefault()
    if (!token || !clubId) return
    setLoading(true)
    setError('')
    setMessage('')
    try {
      const updated = await apiRequest<Club>(`/admin/clubs/${clubId}`, {
        method: 'PATCH',
        body: JSON.stringify({
          name: editForm.name,
          slug: editForm.slug || undefined,
          country: editForm.country || undefined,
          city: editForm.city || undefined,
          commune: editForm.commune || undefined,
          description: editForm.description || undefined,
          founded_at: toISODate(editForm.founded_at),
          is_active: editForm.is_active,
        }),
      }, token)
      setClub(updated)
      setMessage('Club mis à jour')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Mise à jour impossible')
    } finally {
      setLoading(false)
    }
  }

  async function onDeleteClub() {
    if (!token || !clubId || !confirm('Désactiver ce club ?')) return
    await apiRequest(`/admin/clubs/${clubId}`, { method: 'DELETE' }, token)
    navigate('/clubs')
  }

  async function onAssignHead(e: FormEvent) {
    e.preventDefault()
    if (!token || !clubId) return
    setLoading(true)
    setError('')
    setMessage('')
    try {
      await apiRequest(`/admin/clubs/${clubId}/head`, {
        method: 'POST',
        body: JSON.stringify({
          email: headForm.email,
          password: headForm.password,
          first_name: headForm.first_name,
          last_name: headForm.last_name,
          phone: headForm.phone || undefined,
          birth_date: toISODate(headForm.birth_date),
          profession: headForm.profession || undefined,
          member_since: toISODate(headForm.member_since),
        }),
      }, token)
      setMessage('Président créé et assigné')
      setHeadForm({
        email: '', password: '', first_name: '', last_name: '', phone: '', birth_date: '', profession: '', member_since: '',
      })
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Assignation impossible')
    } finally {
      setLoading(false)
    }
  }

  if (!club && !error) return <p className="muted">Chargement…</p>
  if (!club) return <p className="error">{error}</p>

  return (
    <div className="stack gap-lg">
      <Link to="/clubs" className="back">← Clubs</Link>

      <header className="page-header club-header">
        {club.logo_url ? (
          <img src={club.logo_url} alt="" className="club-logo-lg" />
        ) : (
          <span className="brand-mark lg">{club.name[0]}</span>
        )}
        <div>
          <h1>{club.name}</h1>
          <p className="muted">
            Code d&apos;invitation : <code>{club.invite_code}</code>
          </p>
          {club.founded_at && (
            <p className="muted small">Créé le {new Date(club.founded_at).toLocaleDateString('fr-FR')}</p>
          )}
          {!club.is_active && <span className="badge inactive">Inactif</span>}
        </div>
      </header>

      <div className="tabs" role="tablist">
        <button type="button" className={`tab ${section === 'info' ? 'active' : ''}`} onClick={() => setSection('info')}>
          Infos
        </button>
        <button type="button" className={`tab ${section === 'head' ? 'active' : ''}`} onClick={() => setSection('head')}>
          Président
        </button>
        <button type="button" className={`tab ${section === 'members' ? 'active' : ''}`} onClick={() => setSection('members')}>
          Membres{members.length > 0 ? ` (${members.length})` : ''}
        </button>
        <button type="button" className={`tab ${section === 'diary' ? 'active' : ''}`} onClick={() => setSection('diary')}>
          Journal
        </button>
      </div>

      {error && <p className="error" role="alert">{error}</p>}
      {message && <p className="success" role="status">{message}</p>}

      {section === 'info' && (
        <form onSubmit={onUpdateClub} className="panel stack">
          <h2>Modifier le club</h2>
          <div className="grid-2">
            <label>Nom<input value={editForm.name} onChange={(e) => setEditForm({ ...editForm, name: e.target.value })} required /></label>
            <label>Slug<input value={editForm.slug} onChange={(e) => setEditForm({ ...editForm, slug: e.target.value })} /></label>
            <label>Pays<input value={editForm.country} onChange={(e) => setEditForm({ ...editForm, country: e.target.value })} required /></label>
            <label>Ville<input value={editForm.city} onChange={(e) => setEditForm({ ...editForm, city: e.target.value })} required /></label>
            <label>Commune<input value={editForm.commune} onChange={(e) => setEditForm({ ...editForm, commune: e.target.value })} required /></label>
            <label>Date de création<input type="date" value={editForm.founded_at} onChange={(e) => setEditForm({ ...editForm, founded_at: e.target.value })} /></label>
            <label className="span-2">Description<input value={editForm.description} onChange={(e) => setEditForm({ ...editForm, description: e.target.value })} /></label>
          </div>
          <label className="checkbox">
            <input type="checkbox" checked={editForm.is_active} onChange={(e) => setEditForm({ ...editForm, is_active: e.target.checked })} />
            Club actif
          </label>
          <div className="actions">
            <button type="submit" className="btn-primary" disabled={loading}>Enregistrer</button>
            <button type="button" className="btn-danger" onClick={onDeleteClub}>Désactiver</button>
          </div>
        </form>
      )}

      {section === 'head' && (
        <form onSubmit={onAssignHead} className="panel stack">
          <h2>Assigner un président</h2>
          <p className="muted small">Pour les clubs créés sans président (ancien flux).</p>
          <div className="grid-2">
            <label>Prénom<input value={headForm.first_name} onChange={(e) => setHeadForm({ ...headForm, first_name: e.target.value })} required /></label>
            <label>Nom<input value={headForm.last_name} onChange={(e) => setHeadForm({ ...headForm, last_name: e.target.value })} required /></label>
            <label>Email<input type="email" value={headForm.email} onChange={(e) => setHeadForm({ ...headForm, email: e.target.value })} required /></label>
            <label>Mot de passe<input type="password" value={headForm.password} onChange={(e) => setHeadForm({ ...headForm, password: e.target.value })} required minLength={8} /></label>
            <label>Téléphone<input value={headForm.phone} onChange={(e) => setHeadForm({ ...headForm, phone: e.target.value })} /></label>
            <label>Date de naissance<input type="date" value={headForm.birth_date} onChange={(e) => setHeadForm({ ...headForm, birth_date: e.target.value })} /></label>
            <label>Profession<input value={headForm.profession} onChange={(e) => setHeadForm({ ...headForm, profession: e.target.value })} /></label>
            <label>Membre Rotaract depuis<input type="date" value={headForm.member_since} onChange={(e) => setHeadForm({ ...headForm, member_since: e.target.value })} /></label>
          </div>
          <button type="submit" className="btn-primary" disabled={loading}>Créer le président</button>
        </form>
      )}

      {section === 'members' && (
        <section className="panel stack">
          <header className="row-between">
            <div>
              <h2>Membres du club</h2>
              <p className="muted small">{members.length} membre{members.length !== 1 ? 's' : ''} inscrit{members.length !== 1 ? 's' : ''}</p>
            </div>
            <Link to={`/member-cards?club_id=${clubId}`} className="btn-small btn-outline">
              Cartes membre
            </Link>
          </header>

          {membersLoading ? (
            <p className="muted">Chargement…</p>
          ) : members.length === 0 ? (
            <div className="empty-state">
              <strong>Aucun membre</strong>
              <p>Les membres apparaîtront ici après inscription ou approbation d&apos;une demande d&apos;accès.</p>
            </div>
          ) : (
            <ul className="member-list">
              {members.map((member) => (
                <li key={member.id} className="member-row">
                  <div className="member-main">
                    <strong>
                      {member.user?.first_name} {member.user?.last_name}
                    </strong>
                    <span className={`badge ${member.member_role === 'head' ? 'approved' : ''}`}>
                      {member.member_role === 'head' ? 'Responsable' : 'Membre'}
                    </span>
                  </div>
                  <p className="muted small">{member.user?.email}</p>
                  <p className="muted small">
                    {member.user?.phone ? `${member.user.phone} · ` : ''}
                    Inscrit le {new Date(member.joined_at).toLocaleDateString('fr-FR')}
                    {member.user?.member_since
                      ? ` · Rotaract depuis ${new Date(member.user.member_since).toLocaleDateString('fr-FR')}`
                      : ''}
                  </p>
                  {member.user?.profession && (
                    <p className="muted small">{member.user.profession}</p>
                  )}
                </li>
              ))}
            </ul>
          )}
        </section>
      )}

      {section === 'diary' && token && clubId && (
        <div className="panel">
          <ClubDiaryPanel clubId={clubId} token={token} />
        </div>
      )}
    </div>
  )
}
