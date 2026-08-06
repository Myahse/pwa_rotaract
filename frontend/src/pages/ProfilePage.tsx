import { useState, type FormEvent } from 'react'
import { apiRequest } from '../api/client'
import { useAuth } from '../auth/AuthContext'

export function ProfilePage() {
  const { user, token, refresh } = useAuth()
  const [form, setForm] = useState({
    first_name: user?.first_name ?? '',
    last_name: user?.last_name ?? '',
    phone: user?.phone ?? '',
    profession: user?.profession ?? '',
  })
  const [message, setMessage] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    if (!token) return
    setLoading(true)
    setError('')
    setMessage('')
    try {
      await apiRequest('/users/me', {
        method: 'PATCH',
        body: JSON.stringify(form),
      }, token)
      await refresh()
      setMessage('Profil mis à jour')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Mise à jour impossible')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="stack gap-lg">
      <h2>Mon profil</h2>
      {user?.avatar_url && (
        <img src={user.avatar_url} alt="" className="avatar" />
      )}
      <form onSubmit={onSubmit} className="stack card">
        <label>Prénom<input value={form.first_name} onChange={(e) => setForm({ ...form, first_name: e.target.value })} required /></label>
        <label>Nom<input value={form.last_name} onChange={(e) => setForm({ ...form, last_name: e.target.value })} required /></label>
        <label>Téléphone<input value={form.phone} onChange={(e) => setForm({ ...form, phone: e.target.value })} /></label>
        <label>Profession<input value={form.profession} onChange={(e) => setForm({ ...form, profession: e.target.value })} /></label>
        <p className="muted">Email : {user?.email}</p>
        {error && <p className="error">{error}</p>}
        {message && <p className="success">{message}</p>}
        <button type="submit" className="btn-primary" disabled={loading}>Enregistrer</button>
      </form>
    </div>
  )
}
