import { useEffect, useRef, useState } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { apiRequest, ApiClientError } from '../api/client'
import { useAuth } from '../auth/AuthContext'
import { setToken } from '../auth/storage'

type AccessPreview = {
  club_name: string
  contact_email: string
  contact_first_name: string
  contact_last_name: string
  expires_at: string
}

type CompleteResponse = {
  access_token: string
  club: { id: string }
}

declare global {
  interface Window {
    google?: {
      accounts: {
        id: {
          initialize: (config: {
            client_id: string
            callback: (response: { credential: string }) => void
            auto_select?: boolean
            cancel_on_tap_outside?: boolean
          }) => void
          renderButton: (
            parent: HTMLElement,
            options: {
              theme?: string
              size?: string
              text?: string
              shape?: string
              width?: number
            },
          ) => void
        }
      }
    }
  }
}

function loadGoogleScript(): Promise<void> {
  if (window.google?.accounts?.id) return Promise.resolve()
  const existing = document.querySelector<HTMLScriptElement>('script[data-google-gsi]')
  if (existing) {
    return new Promise((resolve, reject) => {
      existing.addEventListener('load', () => resolve())
      existing.addEventListener('error', () => reject(new Error('Google script failed')))
    })
  }
  return new Promise((resolve, reject) => {
    const script = document.createElement('script')
    script.src = 'https://accounts.google.com/gsi/client'
    script.async = true
    script.defer = true
    script.dataset.googleGsi = 'true'
    script.onload = () => resolve()
    script.onerror = () => reject(new Error('Impossible de charger Google Sign-In'))
    document.head.appendChild(script)
  })
}

export function ClubRegisterCompletePage() {
  const [params] = useSearchParams()
  const token = params.get('token') ?? ''
  const navigate = useNavigate()
  const { refresh } = useAuth()
  const buttonRef = useRef<HTMLDivElement>(null)

  const [preview, setPreview] = useState<AccessPreview | null>(null)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)
  const [completing, setCompleting] = useState(false)

  useEffect(() => {
    let cancelled = false

    async function setup() {
      if (!token) {
        setError("Lien d'inscription manquant")
        setLoading(false)
        return
      }

      try {
        const [access, google] = await Promise.all([
          apiRequest<AccessPreview>(`/club-registration/access/${token}`),
          apiRequest<{ client_id: string }>('/auth/google-client-id'),
        ])
        if (cancelled) return
        setPreview(access)

        await loadGoogleScript()
        if (cancelled || !buttonRef.current || !window.google) return

        window.google.accounts.id.initialize({
          client_id: google.client_id,
          callback: async (response) => {
            setCompleting(true)
            setError('')
            try {
              const result = await apiRequest<CompleteResponse>('/club-registration/complete', {
                method: 'POST',
                body: JSON.stringify({
                  token,
                  google_id_token: response.credential,
                }),
              })
              setToken(result.access_token)
              await refresh()
              navigate(result.club?.id ? `/clubs/${result.club.id}` : '/', { replace: true })

            } catch (err) {
              setError(err instanceof Error ? err.message : 'Inscription impossible')
            } finally {
              setCompleting(false)
            }
          },
        })

        buttonRef.current.innerHTML = ''
        window.google.accounts.id.renderButton(buttonRef.current, {
          theme: 'outline',
          size: 'large',
          text: 'continue_with',
          shape: 'rectangular',
          width: 320,
        })
      } catch (err) {
        if (cancelled) return
        if (err instanceof ApiClientError) {
          setError(err.message)
        } else {
          setError(err instanceof Error ? err.message : 'Lien invalide')
        }
      } finally {
        if (!cancelled) setLoading(false)
      }
    }

    setup()
    return () => {
      cancelled = true
    }
  }, [token, navigate, refresh])

  return (
    <div className="auth-page">
      <div className="auth-card">
        <div className="brand hero-brand">
          <span className="brand-dot lg" aria-hidden />
          <h1>Rotaract CIV</h1>
          <p>Finalisez l&apos;inscription de votre club</p>
        </div>

        {loading && <p className="muted center">Chargement…</p>}

        {!loading && preview && (
          <div className="stack">
            <p>
              Club <strong>{preview.club_name}</strong>
            </p>
            <p className="muted">
              Connectez-vous avec le compte Google <strong>{preview.contact_email}</strong> pour créer
              le compte président.
            </p>
            <div ref={buttonRef} className="google-btn-wrap" />
            {completing && <p className="muted center">Création du club…</p>}
          </div>
        )}

        {error && <p className="error">{error}</p>}

        <p className="muted center">
          Déjà inscrit ? <Link to="/login">Se connecter</Link>
        </p>
      </div>
    </div>
  )
}
