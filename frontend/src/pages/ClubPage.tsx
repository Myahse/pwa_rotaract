import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { apiRequest } from '../api/client'
import type { BirthdayMember, ChatGroup, Club, ClubDiaryResponse } from '../api/types'
import { useAuth } from '../auth/AuthContext'
import { DiaryTreeView } from '../components/DiaryTreeView'

export function ClubPage() {
  const { clubId } = useParams<{ clubId: string }>()
  const { token, clubs } = useAuth()
  const membership = clubs.find((c) => c.club_id === clubId)
  const [club, setClub] = useState<Club | null>(null)
  const [groups, setGroups] = useState<ChatGroup[]>([])
  const [birthdays, setBirthdays] = useState<BirthdayMember[]>([])
  const [diary, setDiary] = useState<ClubDiaryResponse | null>(null)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!token || !clubId) return
    Promise.all([
      apiRequest<Club>(`/clubs/${clubId}/`, {}, token),
      apiRequest<ChatGroup[]>(`/clubs/${clubId}/chat/groups`, {}, token),
      apiRequest<BirthdayMember[]>(`/clubs/${clubId}/birthdays/today`, {}, token),
      apiRequest<ClubDiaryResponse>(`/clubs/${clubId}/diary`, {}, token),
    ])
      .then(([c, g, b, d]) => {
        setClub(c)
        setGroups(g)
        setBirthdays(b)
        setDiary(d)
      })
      .catch((err) => setError(err instanceof Error ? err.message : 'Chargement impossible'))
  }, [token, clubId])

  if (error) return <p className="error">{error}</p>
  if (!club) return <p className="muted">Chargement…</p>

  return (
    <div className="stack gap-lg">
      <section>
        <h2>{club.name}</h2>
        {club.city && <p className="muted">{club.city}</p>}
        {membership?.member_role === 'head' && (
          <p><Link to={`/clubs/${clubId}/manage`} className="btn-primary inline">Gérer le club</Link></p>
        )}
      </section>

      <section className="card">
        <h3>Carnet du club</h3>
        {diary ? <DiaryTreeView nodes={diary.tree} /> : <p className="muted">Chargement…</p>}
      </section>

      <section className="card">
        <h3>Anniversaires du jour</h3>
        {birthdays.length === 0 ? (
          <p className="muted">Aucun anniversaire aujourd'hui</p>
        ) : (
          <ul className="list">
            {birthdays.map((u) => (
              <li key={u.user_id}>{u.first_name} {u.last_name}</li>
            ))}
          </ul>
        )}
      </section>

      <section className="card">
        <h3>Groupes de discussion</h3>
        <ul className="list links">
          {groups.map((g) => (
            <li key={g.id}>
              <Link to={`/chat/${g.id}`}>{g.name}</Link>
              <span className="badge">{g.group_type}</span>
            </li>
          ))}
        </ul>
      </section>
    </div>
  )
}
