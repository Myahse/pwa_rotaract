import { useState, type FormEvent } from 'react'
import { Link } from 'react-router-dom'
import { Eye, EyeOff } from 'lucide-react'
import { apiRequest } from '../api/client'

export function AccessRequestPage() {
  const [form, setForm] = useState({
    email: '',
    password: '',
    first_name: '',
    last_name: '',
    club_name: '',
    profession: '',
  })
  const [error, setError] = useState('')
  const [done, setDone] = useState(false)
  const [loading, setLoading] = useState(false)
  const [showPassword, setShowPassword] = useState(false)

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    setLoading(true)
    setError('')
    try {
      await apiRequest('/access-requests', {
        method: 'POST',
        body: JSON.stringify(form),
      })
      setDone(true)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Envoi impossible')
    } finally {
      setLoading(false)
    }
  }

  if (done) {
    return (
      <div className="auth-page">
        <div className="auth-card">
          <h1>Demande envoyée</h1>
          <p>Le responsable du club ou l'administrateur examinera votre demande.</p>
          <Link to="/" className="btn-primary inline">Retour à la connexion</Link>
        </div>
      </div>
    )
  }

  return (
    <div className="auth-page">
      <div className="auth-card wide">
        <Link to="/" className="auth-back">← Connexion</Link>
        <h1>Demander l&apos;accès</h1>
        <p className="muted">Sans invitation email, décrivez le club que vous souhaitez rejoindre.</p>
        <form onSubmit={onSubmit} className="stack">
          <label>Prénom<input value={form.first_name} onChange={(e) => setForm({ ...form, first_name: e.target.value })} required /></label>
          <label>Nom<input value={form.last_name} onChange={(e) => setForm({ ...form, last_name: e.target.value })} required /></label>
          <label>Email<input type="email" value={form.email} onChange={(e) => setForm({ ...form, email: e.target.value })} required /></label>
          <label>Mot de passe
            <div className="password-field">
              <input type={showPassword ? 'text' : 'password'} value={form.password} onChange={(e) => setForm({ ...form, password: e.target.value })} required minLength={8} autoComplete="new-password" />
              <button
                type="button"
                className="password-toggle icon-button"
                onClick={() => setShowPassword((value) => !value)}
                aria-label={showPassword ? 'Masquer le mot de passe' : 'Afficher le mot de passe'}
                title={showPassword ? 'Masquer le mot de passe' : 'Afficher le mot de passe'}
              >
                {showPassword ? <EyeOff size={18} aria-hidden="true" /> : <Eye size={18} aria-hidden="true" />}
              </button>
            </div>
          </label>
          <label>Nom du club<input value={form.club_name} onChange={(e) => setForm({ ...form, club_name: e.target.value })} required /></label>
          <label>Profession<input value={form.profession} onChange={(e) => setForm({ ...form, profession: e.target.value })} /></label>
          {error && <p className="error">{error}</p>}
          <button type="submit" className="btn-primary" disabled={loading}>Envoyer</button>
        </form>
        <p className="muted center"><Link to="/login">Retour</Link></p>
      </div>
    </div>
  )
}
