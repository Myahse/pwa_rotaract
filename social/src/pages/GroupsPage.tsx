import { useCallback, useEffect, useState, type FormEvent } from 'react'
import { Link } from 'react-router-dom'
import { apiRequest } from '../api/client'
import type { SocialGroup, SocialGroupPrivacy } from '../api/types'
import { useAuth } from '../auth/AuthContext'
import { getToken } from '../auth/storage'
import { SkeletonList } from '../components/Skeleton'

export function GroupsPage() {
  const { user, openLogin, requireAuth, token } = useAuth()
  const [tab, setTab] = useState<'discover' | 'mine'>('discover')
  const [groups, setGroups] = useState<SocialGroup[]>([])
  const [ready, setReady] = useState(false)
  const [error, setError] = useState('')
  const [creating, setCreating] = useState(false)
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [privacy, setPrivacy] = useState<SocialGroupPrivacy>('approval')

  const load = useCallback(async () => {
    setError('')
    try {
      const qs = tab === 'mine' ? '?mine=1' : ''
      const data = await apiRequest<{ groups: SocialGroup[] }>(`/social/groups${qs}`, {}, token)
      setGroups(data.groups)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Impossible de charger les groupes')
    } finally {
      setReady(true)
    }
  }, [tab, token])

  useEffect(() => {
    if (tab === 'mine' && !user) {
      setGroups([])
      setReady(true)
      return
    }
    void load()
  }, [load, tab, user])

  function createGroup(e: FormEvent) {
    e.preventDefault()
    requireAuth(async () => {
      setCreating(true)
      setError('')
      try {
        const created = await apiRequest<SocialGroup>('/social/groups', {
          method: 'POST',
          body: JSON.stringify({ name: name.trim(), description: description.trim(), privacy }),
        }, getToken())
        setName('')
        setDescription('')
        setPrivacy('approval')
        setTab('mine')
        setGroups((prev) => [created, ...prev.filter((g) => g.id !== created.id)])
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Création impossible')
      } finally {
        setCreating(false)
      }
    }, 'Connectez-vous pour créer un groupe')
  }

  return (
    <div className="stack gap-md">
      <header className="page-header">
        <h1>Groupes</h1>
        <p className="muted">Rejoignez une communauté, discutez et demandez l&apos;accès aux admins</p>
      </header>

      <div className="feed-tabs">
        <button type="button" className={tab === 'discover' ? 'active' : ''} onClick={() => setTab('discover')}>
          Découvrir
        </button>
        <button
          type="button"
          className={tab === 'mine' ? 'active' : ''}
          onClick={() => {
            if (!user) openLogin('Connectez-vous pour voir vos groupes')
            else setTab('mine')
          }}
        >
          Mes groupes
        </button>
      </div>

      {user && (
        <section className="panel stack">
          <h2>Créer un groupe</h2>
          <form onSubmit={createGroup} className="stack">
            <label>
              Nom
              <input value={name} onChange={(e) => setName(e.target.value)} required minLength={2} maxLength={80} placeholder="Ex. Rotaract Abidjan Projects" />
            </label>
            <label>
              Description
              <textarea value={description} onChange={(e) => setDescription(e.target.value)} maxLength={1000} placeholder="À propos du groupe…" />
            </label>
            <label>
              Accès
              <select value={privacy} onChange={(e) => setPrivacy(e.target.value as SocialGroupPrivacy)}>
                <option value="approval">Sur demande (admin valide)</option>
                <option value="open">Ouvert (rejoindre directement)</option>
              </select>
            </label>
            <button type="submit" className="btn-primary" disabled={creating || name.trim().length < 2}>
              {creating ? 'Création…' : 'Créer le groupe'}
            </button>
          </form>
        </section>
      )}

      {error && <p className="error">{error}</p>}

      <section className="panel stack">
        <h2>{tab === 'mine' ? 'Vos groupes' : 'Groupes à rejoindre'}</h2>
        {(!ready && groups.length === 0) && <SkeletonList count={4} />}
        {ready && groups.length === 0 && (
          <p className="muted">
            {tab === 'mine' ? 'Vous n’avez pas encore de groupe.' : 'Aucun groupe pour le moment. Créez le premier !'}
          </p>
        )}
        {groups.length > 0 && (
          <ul className="suggest-list">
            {groups.map((g) => (
              <li key={g.id}>
                <Link to={`/groups/${g.id}`} className="suggest-user">
                  <div className="avatar-circle">{g.name[0]}</div>
                  <div>
                    <strong>{g.name}</strong>
                    <p className="muted small">
                      {g.member_count} membres · {g.privacy === 'open' ? 'Ouvert' : 'Sur demande'}
                      {g.join_status === 'member' ? ' · Membre' : ''}
                      {g.join_status === 'pending' ? ' · Demande en cours' : ''}
                    </p>
                  </div>
                </Link>
                <Link to={`/groups/${g.id}`} className="btn-primary btn-tiny">Ouvrir</Link>
              </li>
            ))}
          </ul>
        )}
      </section>
    </div>
  )
}
