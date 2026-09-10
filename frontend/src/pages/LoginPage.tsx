import { useState, type FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { Eye, EyeOff } from 'lucide-react'
import { useAuth } from '../auth/AuthContext'

export function LoginPage() {
  const { login } = useAuth()
  const navigate = useNavigate()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      await login(email, password)
      navigate('/home')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Connexion impossible')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="login-screen">
      <div className="login-glow" aria-hidden />
      <div className="login-panel">
        <div className="login-brand">
          <picture>
            <source srcSet="/logo-white.png" media="(prefers-color-scheme: dark)" />
            <img className="login-logo" src="/logo.png" alt="Rotaract IUGB Club" />
          </picture>
          <h1>Rotaract CIV</h1>
          <p>Connectez-vous pour accéder à votre club</p>
        </div>

        <form onSubmit={onSubmit} className="login-form stack">
          <label>
            Email
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
              autoComplete="username"
              placeholder="vous@email.com"
              autoFocus
            />
          </label>
          <label>
            Mot de passe
            <div className="password-field">
              <input
                type={showPassword ? 'text' : 'password'}
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                autoComplete="current-password"
                placeholder="••••••••"
              />
              <button
                type="button"
                className="password-toggle"
                onClick={() => setShowPassword((v) => !v)}
                aria-label={showPassword ? 'Masquer le mot de passe' : 'Afficher le mot de passe'}
                title={showPassword ? 'Masquer le mot de passe' : 'Afficher le mot de passe'}
              >
                {showPassword ? <EyeOff size={18} aria-hidden="true" /> : <Eye size={18} aria-hidden="true" />}
              </button>
            </div>
          </label>
          <Link className="forgot-link" to="/forgot-password">Mot de passe oublié ?</Link>
          {error && <p className="error" role="alert">{error}</p>}
          <button type="submit" className="btn-primary btn-lg btn-block" disabled={loading}>
            {loading ? 'Connexion…' : 'Entrer dans mon club'}
          </button>
        </form>

        <div className="login-links">
          <Link to="/register">Invitation</Link>
          <span aria-hidden>·</span>
          <Link to="/access-request">Demander l&apos;accès</Link>
        </div>
      </div>
    </div>
  )
}
