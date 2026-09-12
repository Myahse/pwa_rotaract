import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { apiRequest } from '../api/client'
import type { BirthdayMember, Club, ClubDiaryResponse } from '../api/types'
import { useAuth } from '../auth/AuthContext'
import { DiaryTreeView } from '../components/DiaryTreeView'

export function ClubPage() {
  const { clubId } = useParams<{ clubId: string }>()
  const { token, uiCaps } = useAuth()
  const [club, setClub] = useState<Club | null>(null)
  const [birthdays, setBirthdays] = useState<BirthdayMember[]>([])
  const [diary, setDiary] = useState<ClubDiaryResponse | null>(null)
  const [error, setError] = useState('')

  const groupsOnly = uiCaps.clubView === 'groups_only'

  useEffect(() => {
    if (!token || !clubId) return
    const requests: Promise<unknown>[] = [
      apiRequest<Club>(`/clubs/${clubId}/`, {}, token),
    ]
    if (!groupsOnly) {
      requests.push(
        apiRequest<BirthdayMember[]>(`/clubs/${clubId}/birthdays/today`, {}, token),
        apiRequest<ClubDiaryResponse>(`/clubs/${clubId}/diary`, {}, token),
      )
    }
    Promise.all(requests)
      .then((results) => {
        setClub(results[0] as Club)
        if (!groupsOnly) {
          setBirthdays(results[1] as BirthdayMember[])
          setDiary(results[2] as ClubDiaryResponse)
        }
      })
      .catch((err) => setError(err instanceof Error ? err.message : 'Chargement impossible'))
  }, [token, clubId, groupsOnly])

  if (error) return <p className="error">{error}</p>
  if (!club) return <p className="muted">Chargement…</p>

  return (
    <div className="stack gap-lg">
      <section>
        <h2>{groupsOnly ? 'Ma commission' : club.name}</h2>
        {!groupsOnly && club.city && <p className="muted">{club.city}</p>}
        <div className="actions">
          {uiCaps.nav.cotisations && (
            <Link to={`/clubs/${clubId}/cotisations`} className="btn-outline">Cotisations mensuelles</Link>
          )}
          {uiCaps.nav.manage && (
            <Link to={`/clubs/${clubId}/manage`} className="btn-primary inline">Gérer le club</Link>
          )}
        </div>
      </section>

      {!groupsOnly && (
        <>
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
        </>
      )}

      <section className="card">
        <h3>{groupsOnly ? 'Groupe de discussion' : 'Messages'}</h3>
        <p className="muted">Conversations privées, groupes du club et fil général.</p>
        <Link to={`/clubs/${clubId}/messages`} className="btn-primary inline">Ouvrir les messages</Link>
      </section>
    </div>
  )
}
