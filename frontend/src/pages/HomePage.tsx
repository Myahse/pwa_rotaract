import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { apiRequest } from '../api/client'
import type { BirthdayWidget } from '../api/types'
import { useAuth } from '../auth/AuthContext'

export function HomePage() {
  const { user, token, activeClubId, clubs, uiCaps } = useAuth()
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
          <p className="muted">{uiCaps.profileLabel} · {membership.club.name}</p>
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
          {uiCaps.nav.club && (
            <Link to={`/clubs/${activeClubId}`} className="tile card">
              <h3>{uiCaps.profile === 'commission_president' ? 'Ma commission' : 'Mon club'}</h3>
              <p className="muted small">
                {uiCaps.clubView === 'groups_only'
                  ? 'Discussion de votre commission'
                  : 'Carnet, anniversaires, groupes'}
              </p>
            </Link>
          )}
          {uiCaps.nav.cotisations && (
            <Link to={`/clubs/${activeClubId}/cotisations`} className="tile card">
              <h3>Cotisations</h3>
              <p className="muted small">
                {uiCaps.canManageDues
                  ? 'Valider les reçus Wave des membres'
                  : 'Payer et envoyer votre reçu Wave'}
              </p>
            </Link>
          )}
          {uiCaps.nav.manage && (
            <Link to={`/clubs/${activeClubId}/manage`} className="tile card">
              <h3>Gestion</h3>
              <p className="muted small">
                {uiCaps.profile === 'secretary'
                  ? 'Membres et invitations'
                  : 'Membres, rôles, commissions'}
              </p>
            </Link>
          )}
        </section>
      ) : (
        <p className="muted">Vous n'êtes membre d'aucun club pour le moment.</p>
      )}
    </div>
  )
}
