import type { ClubDiaryEntryType, ClubDiaryTreeNode } from '../api/types'

const TYPE_LABELS: Record<ClubDiaryEntryType, string> = {
  parrain: 'Parrain',
  president: 'Président',
  member: 'Membre',
}

function nodeLabel(node: ClubDiaryTreeNode) {
  if (node.last_name === 'Mandat') return 'Mandat'
  if (node.last_name === 'Commission') return 'Commission'
  return TYPE_LABELS[node.entry_type]
}

function formatPeriod(started?: string, ended?: string) {
  if (!started && !ended) return null
  const start = started ? new Date(started).toLocaleDateString('fr-FR') : '?'
  const end = ended ? new Date(ended).toLocaleDateString('fr-FR') : '…'
  return `${start} → ${end}`
}

type Props = {
  nodes: ClubDiaryTreeNode[]
  depth?: number
}

export function DiaryTreeView({ nodes, depth = 0 }: Props) {
  if (nodes.length === 0 && depth === 0) {
    return <p className="muted">Aucune entrée dans le carnet du club.</p>
  }

  return (
    <ul className={`diary-tree depth-${depth}`}>
      {nodes.map((node) => (
        <li key={node.id} className="diary-node">
          <div className={`diary-card type-${node.entry_type} ${node.last_name === 'Mandat' ? 'type-mandate' : ''}`}>
            {node.photo_url && <img src={node.photo_url} alt="" className="diary-photo" />}
            <div>
              <div className="row-between">
                <strong>{node.first_name} {node.last_name !== 'Mandat' && node.last_name !== 'Commission' ? node.last_name : ''}</strong>
                <span className="badge">{nodeLabel(node)}</span>
              </div>
              {formatPeriod(node.started_at, node.ended_at) && (
                <p className="muted small">{formatPeriod(node.started_at, node.ended_at)}</p>
              )}
              {node.notes && <p className="diary-notes">{node.notes}</p>}
            </div>
          </div>
          {node.children.length > 0 && <DiaryTreeView nodes={node.children} depth={depth + 1} />}
        </li>
      ))}
    </ul>
  )
}

export const ROLE_LABELS = {
  president: 'Président',
  vice_president: 'Vice-président',
  secretary: 'Secrétaire',
  treasurer: 'Trésorier',
  commission_president: 'Prés. commission',
  commission_secretary: 'Sec. commission',
  commission_member: 'Membre commission',
  member: 'Membre',
} as const
