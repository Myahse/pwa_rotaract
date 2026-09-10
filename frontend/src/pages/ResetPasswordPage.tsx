import { useState, type FormEvent } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { apiRequest } from '../api/client'

export function ResetPasswordPage() {
  const [params] = useSearchParams()
  const navigate = useNavigate()
  const token = params.get('token') ?? ''
  const [password, setPassword] = useState('')
  const [confirmation, setConfirmation] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [done, setDone] = useState(false)

  async function onSubmit(event: FormEvent) {
    event.preventDefault()
    setError('')
    if (password.length < 8) { setError('Le mot de passe doit contenir au moins 8 caractères.'); return }
    if (password !== confirmation) { setError('Les mots de passe ne correspondent pas.'); return }
    setLoading(true)
    try {
      await apiRequest('/auth/reset-password', { method: 'POST', body: JSON.stringify({ token, password }) })
      setDone(true)
      setTimeout(() => navigate('/'), 1600)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Lien invalide ou expiré')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="auth-page">
      <div className="auth-card">
        <div className="auth-brand"><img src="/logo.png" alt="Rotaract IUGB Club" /></div>
        <h1>Nouveau mot de passe</h1>
        {done ? <p className="success auth-message">Mot de passe mis à jour. Redirection vers la connexion…</p> : (
          <form onSubmit={onSubmit} className="stack">
            <label>Nouveau mot de passe<input type="password" value={password} onChange={(event) => setPassword(event.target.value)} minLength={8} required autoComplete="new-password" /></label>
            <label>Confirmer le mot de passe<input type="password" value={confirmation} onChange={(event) => setConfirmation(event.target.value)} minLength={8} required autoComplete="new-password" /></label>
            {error && <p className="error">{error}</p>}
            <button type="submit" className="btn-primary btn-block" disabled={loading || !token}>{loading ? 'Mise à jour…' : 'Enregistrer le mot de passe'}</button>
          </form>
        )}
        <p className="muted center"><Link to="/">Retour à la connexion</Link></p>
      </div>
    </div>
  )
}