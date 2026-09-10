import { useEffect, useState, type FormEvent } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { apiRequest } from '../api/client'
import type { ClubRole } from '../api/types'

type InvitePreview = {
  club: { id: string; name: string }
  roles: ClubRole[]
  email?: string
}

export function RegisterPage() {
  const [params] = useSearchParams()
  const navigate = useNavigate()
  const token = params.get('token') ?? ''
  const code = params.get('code') ?? ''

  const [preview, setPreview] = useState<InvitePreview | null>(null)
  const [roleId, setRoleId] = useState('')
  const [form, setForm] = useState({
    email: '',
    password: '',
    first_name: '',
    last_name: '',
    profession: '',
  })
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const [inviteCodeInput, setInviteCodeInput] = useState('')

  useEffect(() => {
    async function load() {
      try {
        setError('')
        if (token) {
          const data = await apiRequest<InvitePreview & { email: string }>(`/invite/token/${token}`)
          setPreview(data)
          setForm((f) => ({ ...f, email: data.email }))
          const membre = data.roles.find((r) => r.name === 'Membre')
          if (membre) setRoleId(membre.id)
        } else if (code || inviteCodeInput) {
          const targetCode = code || inviteCodeInput
          const data = await apiRequest<InvitePreview>(`/invite/${targetCode}`)
          setPreview(data)
          const membre = data.roles.find((r) => r.name === 'Membre')
          if (membre) setRoleId(membre.id)
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Invitation invalide')
        setPreview(null)
      }
    }
    if (token || code || inviteCodeInput.length >= 4) load()
  }, [token, code, inviteCodeInput])

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    if (!preview || !roleId) return
    setLoading(true)
    setError('')
    try {
      await apiRequest('/register', {
        method: 'POST',
        body: JSON.stringify({
          ...form,
          token: token || undefined,
          invite_code: code || inviteCodeInput || undefined,
          role_id: roleId,
        }),
      })
      navigate('/login')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Inscription impossible')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="auth-page">
      <div className="auth-card wide">
        <Link to="/" className="auth-back">← Connexion</Link>
        <h1>Créer un compte</h1>
        
        {!token && !code && !preview && (
          <div className="stack" style={{ marginBottom: '1.5rem' }}>
            <label>
              Code d'invitation du club
              <input 
                placeholder="Ex: AB123" 
                value={inviteCodeInput} 
                onChange={(e) => setInviteCodeInput(e.target.value.toUpperCase())}
              />
            </label>
            <p className="small muted">Saisissez le code fourni par votre club pour continuer.</p>
          </div>
        )}

        {preview && <p className="success-banner">Club : <strong>{preview.club.name}</strong></p>}
        
        <form onSubmit={onSubmit} className="stack">
          <label>Prénom<input value={form.first_name} onChange={(e) => setForm({ ...form, first_name: e.target.value })} required disabled={!preview} /></label>
          <label>Nom<input value={form.last_name} onChange={(e) => setForm({ ...form, last_name: e.target.value })} required disabled={!preview} /></label>
          <label>Email<input type="email" value={form.email} onChange={(e) => setForm({ ...form, email: e.target.value })} required readOnly={!!token} disabled={!preview} /></label>
          <label>Mot de passe<input type="password" value={form.password} onChange={(e) => setForm({ ...form, password: e.target.value })} required minLength={8} disabled={!preview} /></label>
          <label>Profession<input value={form.profession} onChange={(e) => setForm({ ...form, profession: e.target.value })} disabled={!preview} /></label>
          {preview && preview.roles.length > 0 && (
            <label>
              Rôle
              <select value={roleId} onChange={(e) => setRoleId(e.target.value)} required>
                {preview.roles.map((r) => (
                  <option key={r.id} value={r.id}>{r.name}</option>
                ))}
              </select>
            </label>
          )}
          {error && <p className="error">{error}</p>}
          <button type="submit" className="btn-primary" disabled={loading || !preview}>S'inscrire</button>
        </form>
        <p className="muted center"><Link to="/login">Déjà inscrit ?</Link></p>
      </div>
    </div>
  )
}
