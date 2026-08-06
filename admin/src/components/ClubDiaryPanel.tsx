import { useEffect, useState, type FormEvent } from 'react'
import { apiRequest } from '../api/client'
import type {
  ClubDiaryResponse,
  ClubMandate,
  ClubMandateRole,
  Commission,
} from '../api/types'
import { DiaryTreeView, ROLE_LABELS } from './DiaryTreeView'

function toISODate(value: string) {
  return value ? `${value}T00:00:00.000Z` : undefined
}

type Props = {
  clubId: string
  token: string
}

const emptyMandate = { name: '', started_at: '', ended_at: '', is_current: true, notes: '' }
const emptyAssignment = {
  role: 'president' as ClubMandateRole,
  first_name: '',
  last_name: '',
  commission_name: '',
  notes: '',
}

export function ClubDiaryPanel({ clubId, token }: Props) {
  const [diary, setDiary] = useState<ClubDiaryResponse | null>(null)
  const [commissions, setCommissions] = useState<Commission[]>([])
  const [parrainForm, setParrainForm] = useState({ first_name: '', last_name: '', organization: '', notes: '' })
  const [mandateForm, setMandateForm] = useState(emptyMandate)
  const [selectedMandate, setSelectedMandate] = useState('')
  const [assignmentForm, setAssignmentForm] = useState(emptyAssignment)
  const [error, setError] = useState('')
  const [message, setMessage] = useState('')
  const [loading, setLoading] = useState(false)

  async function load() {
    const [data, comms] = await Promise.all([
      apiRequest<ClubDiaryResponse>(`/clubs/${clubId}/diary`, {}, token),
      apiRequest<Commission[]>(`/clubs/${clubId}/commissions`, {}, token),
    ])
    setDiary(data)
    setCommissions(comms)
    if (!selectedMandate && data.mandates.length > 0) {
      setSelectedMandate(data.mandates[0].id)
    }
  }

  useEffect(() => {
    load().catch((err) => setError(err instanceof Error ? err.message : 'Erreur'))
  }, [clubId, token])

  async function addParrain(e: FormEvent) {
    e.preventDefault()
    setLoading(true)
    setError('')
    try {
      await apiRequest(`/clubs/${clubId}/diary`, {
        method: 'POST',
        body: JSON.stringify({
          entry_type: 'parrain',
          first_name: parrainForm.first_name,
          last_name: parrainForm.last_name,
          organization: parrainForm.organization || undefined,
          notes: parrainForm.notes || undefined,
        }),
      }, token)
      setParrainForm({ first_name: '', last_name: '', organization: '', notes: '' })
      setMessage('Parrain ajouté')
      await load()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erreur')
    } finally {
      setLoading(false)
    }
  }

  async function addMandate(e: FormEvent) {
    e.preventDefault()
    setLoading(true)
    setError('')
    try {
      await apiRequest(`/clubs/${clubId}/mandates`, {
        method: 'POST',
        body: JSON.stringify({
          name: mandateForm.name,
          started_at: toISODate(mandateForm.started_at),
          ended_at: toISODate(mandateForm.ended_at),
          is_current: mandateForm.is_current,
          notes: mandateForm.notes || undefined,
        }),
      }, token)
      setMandateForm(emptyMandate)
      setMessage('Mandat créé')
      await load()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erreur')
    } finally {
      setLoading(false)
    }
  }

  async function addAssignment(e: FormEvent) {
    e.preventDefault()
    if (!selectedMandate) return
    setLoading(true)
    setError('')
    try {
      const isCommission = assignmentForm.role.startsWith('commission_')
      await apiRequest(`/clubs/${clubId}/mandates/${selectedMandate}/assignments`, {
        method: 'POST',
        body: JSON.stringify({
          role: assignmentForm.role,
          first_name: assignmentForm.first_name,
          last_name: assignmentForm.last_name,
          commission_name: isCommission ? assignmentForm.commission_name : undefined,
          notes: assignmentForm.notes || undefined,
        }),
      }, token)
      setAssignmentForm({ ...emptyAssignment, role: assignmentForm.role })
      setMessage('Rôle ajouté au mandat')
      await load()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erreur')
    } finally {
      setLoading(false)
    }
  }

  async function deleteMandate(m: ClubMandate) {
    if (!confirm(`Supprimer le mandat ${m.name} ?`)) return
    await apiRequest(`/clubs/${clubId}/mandates/${m.id}`, { method: 'DELETE' }, token)
    await load()
  }

  return (
    <section className="card stack">
      <div>
        <h2>Carnet du club</h2>
        <p className="muted">Organisé par mandat : parrain, bureau et commissions de chaque génération.</p>
      </div>

      {diary && <DiaryTreeView nodes={diary.tree} />}

      <form onSubmit={addParrain} className="stack nested">
        <h3>Parrain</h3>
        <div className="grid-2">
          <label>Prénom<input value={parrainForm.first_name} onChange={(e) => setParrainForm({ ...parrainForm, first_name: e.target.value })} required /></label>
          <label>Nom<input value={parrainForm.last_name} onChange={(e) => setParrainForm({ ...parrainForm, last_name: e.target.value })} required /></label>
          <label className="span-2">Organisation / club parrain<input value={parrainForm.organization} onChange={(e) => setParrainForm({ ...parrainForm, organization: e.target.value })} /></label>
        </div>
        <button type="submit" className="btn-primary" disabled={loading}>Ajouter le parrain</button>
      </form>

      <form onSubmit={addMandate} className="stack nested">
        <h3>Nouveau mandat</h3>
        <div className="grid-2">
          <label>Nom (ex. 2024-2025)<input value={mandateForm.name} onChange={(e) => setMandateForm({ ...mandateForm, name: e.target.value })} required /></label>
          <label className="checkbox"><input type="checkbox" checked={mandateForm.is_current} onChange={(e) => setMandateForm({ ...mandateForm, is_current: e.target.checked })} />Mandat actuel</label>
          <label>Début<input type="date" value={mandateForm.started_at} onChange={(e) => setMandateForm({ ...mandateForm, started_at: e.target.value })} required /></label>
          <label>Fin<input type="date" value={mandateForm.ended_at} onChange={(e) => setMandateForm({ ...mandateForm, ended_at: e.target.value })} /></label>
        </div>
        <button type="submit" className="btn-primary" disabled={loading}>Créer le mandat</button>
      </form>

      {diary && diary.mandates.length > 0 && (
        <div className="stack nested">
          <h3>Composition par mandat</h3>
          {diary.mandates.map((m) => (
            <div key={m.id} className="mandate-card stack">
              <div className="row-between">
                <strong>{m.name} {m.is_current && <span className="badge">actuel</span>}</strong>
                <button type="button" className="btn-ghost btn-small" onClick={() => deleteMandate(m)}>Supprimer</button>
              </div>
              {m.bureau.length > 0 && (
                <div>
                  <h4>Bureau</h4>
                  <ul className="list">{m.bureau.map((a) => (
                    <li key={a.id}>{ROLE_LABELS[a.role]} — {a.first_name} {a.last_name}</li>
                  ))}</ul>
                </div>
              )}
              {m.commissions.map((c) => (
                <div key={c.commission_name}>
                  <h4>{c.commission_name}</h4>
                  <ul className="list">
                    {c.president && <li>Président — {c.president.first_name} {c.president.last_name}</li>}
                    {c.secretary && <li>Secrétaire — {c.secretary.first_name} {c.secretary.last_name}</li>}
                    {c.members.map((a) => (
                      <li key={a.id}>Membre — {a.first_name} {a.last_name}</li>
                    ))}
                  </ul>
                </div>
              ))}
            </div>
          ))}

          <form onSubmit={addAssignment} className="stack">
            <h4>Ajouter un rôle au mandat</h4>
            <div className="grid-2">
              <label>Mandat
                <select value={selectedMandate} onChange={(e) => setSelectedMandate(e.target.value)} required>
                  {diary.mandates.map((m) => (
                    <option key={m.id} value={m.id}>{m.name}</option>
                  ))}
                </select>
              </label>
              <label>Rôle
                <select value={assignmentForm.role} onChange={(e) => setAssignmentForm({ ...assignmentForm, role: e.target.value as ClubMandateRole })}>
                  {Object.entries(ROLE_LABELS).map(([value, label]) => (
                    <option key={value} value={value}>{label}</option>
                  ))}
                </select>
              </label>
              <label>Prénom<input value={assignmentForm.first_name} onChange={(e) => setAssignmentForm({ ...assignmentForm, first_name: e.target.value })} required /></label>
              <label>Nom<input value={assignmentForm.last_name} onChange={(e) => setAssignmentForm({ ...assignmentForm, last_name: e.target.value })} required /></label>
              {assignmentForm.role.startsWith('commission_') && (
                <label className="span-2">Commission
                  <input list="commission-names" value={assignmentForm.commission_name} onChange={(e) => setAssignmentForm({ ...assignmentForm, commission_name: e.target.value })} required />
                  <datalist id="commission-names">
                    {commissions.map((c) => <option key={c.id} value={c.name} />)}
                  </datalist>
                </label>
              )}
            </div>
            <button type="submit" className="btn-primary" disabled={loading}>Ajouter au mandat</button>
          </form>
        </div>
      )}

      {error && <p className="error">{error}</p>}
      {message && <p className="success">{message}</p>}
    </section>
  )
}
