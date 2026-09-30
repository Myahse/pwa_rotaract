import { Users } from 'lucide-react'

function initialsFromName(name: string): string {
  const parts = name.trim().split(/\s+/).filter(Boolean)
  if (parts.length === 0) return '?'
  if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase()
  return `${parts[0][0] ?? ''}${parts[1][0] ?? ''}`.toUpperCase()
}

type Props = {
  name: string
  isGroup?: boolean
  size?: 'md' | 'lg'
}

export function ChatAvatar({ name, isGroup, size = 'md' }: Props) {
  return (
    <span className={`chat-avatar chat-avatar--${size}${isGroup ? ' chat-avatar--group' : ''}`} aria-hidden="true">
      {isGroup ? <Users className="chat-avatar-icon" /> : initialsFromName(name)}
    </span>
  )
}
