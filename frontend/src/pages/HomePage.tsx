import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { apiRequest } from '../api/client'
import type { BirthdayWidget } from '../api/types'
import { useAuth } from '../auth/AuthContext'

export function HomePage() {
  const { user, token, activeClubId, clubs } = useAuth()
  const [widget, setWidget] = useState<BirthdayWidget | null>(null)

  useEffect(() => {
    if (!token) return
    apiRequest<BirthdayWidget>('/widgets/birthday', {}, token)
      .then(setWidget)
      .catch(() => setWidget(null))
  }, [token])

  const membership = clubs.find((c) => c.club_id === activeClubId)

  return (
    <div className="stack gap-lg">
      <section className="hero-card">
        <p className="eyebrow">Bienvenue</p>
        <h2>{user?.first_name} {user?.last_name}</h2>
        {membership?.club && (
          <p className="muted">{membership.member_role === 'head' ? 'Responsable' : 'Membre'} · {membership.club.name}</p>
        )}
      </section>

      {widget?.is_my_birthday && (
        <section className="birthday-card">
          <span className="birthday-emoji">{widget.emoji ?? '🎂'}</span>
          <div>
            <h3>{widget.title}</h3>
            <p>{widget.message}</p>
          </div>
        </section>
      )}

      {activeClubId ? (
        <section className="card-grid">
          <Link to={`/clubs/${activeClubId}`} className="tile">
            <h3>Mon club</h3>
            <p>Membres, groupes, anniversaires</p>
          </Link>
        </section>
      ) : (
        <p className="muted">Vous n'êtes membre d'aucun club pour le moment.</p>
      )}
    </div>
  )
}
