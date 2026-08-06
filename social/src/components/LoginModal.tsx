import { useState, type FormEvent } from 'react'
import { useAuth } from '../auth/AuthContext'

export function LoginModal() {
  const { loginOpen, closeLogin, login, loginReason } = useAuth()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  if (!loginOpen) return null

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      await login(email, password)
      setEmail('')
      setPassword('')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Connexion impossible')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="modal-backdrop" onClick={closeLogin} role="presentation">
      <div className="modal-card" role="dialog" aria-modal="true" aria-labelledby="login-title" onClick={(e) => e.stopPropagation()}>
        <button type="button" className="modal-close" onClick={closeLogin} aria-label="Fermer">×</button>
        <div className="modal-head">
          <div className="brand-dot lg" aria-hidden />
          <h2 id="login-title">Connexion</h2>
          <p className="muted">{loginReason}</p>
        </div>
        <form onSubmit={onSubmit} className="stack">
          <label>
            Email
            <input type="email" value={email} onChange={(e) => setEmail(e.target.value)} required autoComplete="username" autoFocus />
          </label>
          <label>
            Mot de passe
            <input type="password" value={password} onChange={(e) => setPassword(e.target.value)} required autoComplete="current-password" />
          </label>
          {error && <p className="error" role="alert">{error}</p>}
          <button type="submit" className="btn-primary btn-block" disabled={loading}>
            {loading ? 'Connexion…' : 'Continuer'}
          </button>
        </form>
      </div>
    </div>
  )
}
