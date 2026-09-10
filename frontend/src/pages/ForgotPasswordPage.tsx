import { useState, type FormEvent } from 'react'
import { Link } from 'react-router-dom'
import { apiRequest } from '../api/client'

export function ForgotPasswordPage() {
  const [email, setEmail] = useState('')
  const [sent, setSent] = useState(false)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  async function onSubmit(event: FormEvent) {
    event.preventDefault()
    setError('')
    setLoading(true)
    try {
      await apiRequest('/auth/forgot-password', { method: 'POST', body: JSON.stringify({ email }) })
      setSent(true)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Impossible de traiter la demande')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="auth-page">
      <div className="auth-card">
        <Link to="/" className="auth-back">← Connexion</Link>
        <div className="auth-brand"><img src="/logo.png" alt="Rotaract IUGB Club" /></div>
        <h1>Mot de passe oublié ?</h1>
        {sent ? (
          <p className="success auth-message">Si cette adresse existe, vous recevrez un lien de réinitialisation.</p>
        ) : (
          <form onSubmit={onSubmit} className="stack">
            <p className="muted">Saisissez votre adresse email pour recevoir un lien sécurisé.</p>
            <label>Email<input type="email" value={email} onChange={(event) => setEmail(event.target.value)} required autoComplete="email" /></label>
            {error && <p className="error">{error}</p>}
            <button type="submit" className="btn-primary btn-block" disabled={loading}>{loading ? 'Envoi…' : 'Recevoir le lien'}</button>
          </form>
        )}
      </div>
    </div>
  )
}