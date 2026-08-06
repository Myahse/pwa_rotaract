import { useEffect, useState, type FormEvent } from 'react'
import { apiRequest } from '../api/client'
import type { ClubDiaryResponse, ClubMandateRole, Commission } from '../api/types'
import { DiaryTreeView, ROLE_LABELS } from './DiaryTreeView'

function toISODate(value: string) {
  return value ? `${value}T00:00:00.000Z` : undefined
}

type Props = { clubId: string; token: string }

export function ClubDiaryPanel({ clubId, token }: Props) {
  const [diary, setDiary] = useState<ClubDiaryResponse | null>(null)
  const [commissions, setCommissions] = useState<Commission[]>([])
  const [selectedMandate, setSelectedMandate] = useState('')
  const [mandateForm, setMandateForm] = useState({ name: '', started_at: '', ended_at: '', is_current: true })
  const [assignmentForm, setAssignmentForm] = useState({
    role: 'president' as ClubMandateRole,
    first_name: '',
    last_name: '',
    commission_name: '',
  })
  const [error, setError] = useState('')

  async function load() {
    const [data, comms] = await Promise.all([
      apiRequest<ClubDiaryResponse>(`/clubs/${clubId}/diary`, {}, token),
      apiRequest<Commission[]>(`/clubs/${clubId}/commissions`, {}, token),
    ])
    setDiary(data)
    setCommissions(comms)
    if (!selectedMandate && data.mandates.length > 0) setSelectedMandate(data.mandates[0].id)
  }

  useEffect(() => {
    load().catch((err) => setError(err instanceof Error ? err.message : 'Erreur'))
  }, [clubId, token])

  async function addMandate(e: FormEvent) {
    e.preventDefault()
    await apiRequest(`/clubs/${clubId}/mandates`, {
      method: 'POST',
      body: JSON.stringify({
        name: mandateForm.name,
        started_at: toISODate(mandateForm.started_at),
        ended_at: toISODate(mandateForm.ended_at),
        is_current: mandateForm.is_current,
      }),
    }, token)
    setMandateForm({ name: '', started_at: '', ended_at: '', is_current: true })
    await load()
  }

  async function addAssignment(e: FormEvent) {
    e.preventDefault()
    if (!selectedMandate) return
    const isCommission = assignmentForm.role.startsWith('commission_')
    await apiRequest(`/clubs/${clubId}/mandates/${selectedMandate}/assignments`, {
      method: 'POST',
      body: JSON.stringify({
        role: assignmentForm.role,
        first_name: assignmentForm.first_name,
        last_name: assignmentForm.last_name,
        commission_name: isCommission ? assignmentForm.commission_name : undefined,
      }),
    }, token)
    setAssignmentForm({ role: assignmentForm.role, first_name: '', last_name: '', commission_name: '' })
    await load()
  }

  return (
    <section className="card stack">
      <h3>Carnet du club (mandats)</h3>
      {diary && <DiaryTreeView nodes={diary.tree} />}
      {error && <p className="error">{error}</p>}

      <form onSubmit={addMandate} className="stack nested">
        <h4>Nouveau mandat</h4>
        <div className="grid-2">
          <label>Nom<input value={mandateForm.name} onChange={(e) => setMandateForm({ ...mandateForm, name: e.target.value })} required /></label>
          <label className="checkbox"><input type="checkbox" checked={mandateForm.is_current} onChange={(e) => setMandateForm({ ...mandateForm, is_current: e.target.checked })} />Actuel</label>
          <label>Début<input type="date" value={mandateForm.started_at} onChange={(e) => setMandateForm({ ...mandateForm, started_at: e.target.value })} required /></label>
          <label>Fin<input type="date" value={mandateForm.ended_at} onChange={(e) => setMandateForm({ ...mandateForm, ended_at: e.target.value })} /></label>
        </div>
        <button type="submit" className="btn-primary">Créer mandat</button>
      </form>

      {diary && diary.mandates.length > 0 && (
        <form onSubmit={addAssignment} className="stack nested">
          <h4>Ajouter un rôle</h4>
          <div className="grid-2">
            <label>Mandat<select value={selectedMandate} onChange={(e) => setSelectedMandate(e.target.value)}>
              {diary.mandates.map((m) => <option key={m.id} value={m.id}>{m.name}</option>)}
            </select></label>
            <label>Rôle<select value={assignmentForm.role} onChange={(e) => setAssignmentForm({ ...assignmentForm, role: e.target.value as ClubMandateRole })}>
              {Object.entries(ROLE_LABELS).map(([v, l]) => <option key={v} value={v}>{l}</option>)}
            </select></label>
            <label>Prénom<input value={assignmentForm.first_name} onChange={(e) => setAssignmentForm({ ...assignmentForm, first_name: e.target.value })} required /></label>
            <label>Nom<input value={assignmentForm.last_name} onChange={(e) => setAssignmentForm({ ...assignmentForm, last_name: e.target.value })} required /></label>
            {assignmentForm.role.startsWith('commission_') && (
              <label className="span-2">Commission
                <input list="commissions-list" value={assignmentForm.commission_name} onChange={(e) => setAssignmentForm({ ...assignmentForm, commission_name: e.target.value })} required />
                <datalist id="commissions-list">{commissions.map((c) => <option key={c.id} value={c.name} />)}</datalist>
              </label>
            )}
          </div>
          <button type="submit" className="btn-primary">Ajouter</button>
        </form>
      )}
    </section>
  )
}
