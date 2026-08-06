import type { CSSProperties } from 'react'

type BlockProps = {
  className?: string
  style?: CSSProperties
}

function Block({ className = '', style }: BlockProps) {
  return <span className={`sk ${className}`.trim()} style={style} aria-hidden />
}

export function SkeletonPost({ lines = 2 }: { lines?: number }) {
  return (
    <div className="sk-post" aria-hidden>
      <div className="sk-post-head">
        <Block className="sk-circle" />
        <div className="sk-col">
          <Block className="sk-line short" />
          <Block className="sk-line tiny" />
        </div>
      </div>
      {Array.from({ length: lines }, (_, i) => (
        <Block key={i} className={`sk-line ${i === lines - 1 ? 'medium' : 'long'}`} />
      ))}
      <Block className="sk-media" />
      <div className="sk-actions">
        <Block className="sk-line tiny" />
        <Block className="sk-line tiny" />
        <Block className="sk-line tiny" />
      </div>
    </div>
  )
}

export function SkeletonFeed({ count = 3 }: { count?: number }) {
  return (
    <div className="sk-section" aria-busy="true" aria-label="Chargement du fil">
      {Array.from({ length: count }, (_, i) => (
        <SkeletonPost key={i} lines={i % 2 === 0 ? 2 : 3} />
      ))}
    </div>
  )
}

export function SkeletonComposer() {
  return (
    <div className="sk-composer" aria-hidden>
      <div className="sk-post-head">
        <Block className="sk-circle" />
        <Block className="sk-pill" />
      </div>
    </div>
  )
}

export function SkeletonListRow() {
  return (
    <li className="sk-list-row" aria-hidden>
      <div className="sk-post-head">
        <Block className="sk-circle" />
        <div className="sk-col">
          <Block className="sk-line short" />
          <Block className="sk-line tiny" />
        </div>
      </div>
      <Block className="sk-chip" />
    </li>
  )
}

export function SkeletonList({ count = 4 }: { count?: number }) {
  return (
    <ul className="suggest-list sk-section" aria-busy="true" aria-label="Chargement">
      {Array.from({ length: count }, (_, i) => (
        <SkeletonListRow key={i} />
      ))}
    </ul>
  )
}

export function SkeletonProfile() {
  return (
    <div className="sk-section" aria-busy="true" aria-label="Chargement du profil">
      <div className="sk-profile-hero">
        <Block className="sk-circle xl" />
        <div className="sk-col">
          <Block className="sk-line medium" />
          <Block className="sk-line short" />
          <div className="sk-actions">
            <Block className="sk-chip" />
            <Block className="sk-chip" />
          </div>
        </div>
      </div>
      <div className="sk-panel">
        <Block className="sk-line short" />
        <Block className="sk-line medium" />
        <Block className="sk-line tiny" />
      </div>
      <SkeletonFeed count={2} />
    </div>
  )
}

export function SkeletonSuggestions({ count = 3 }: { count?: number }) {
  return (
    <div className="suggest-card sk-section" aria-busy="true" aria-label="Chargement des suggestions">
      <Block className="sk-line short" />
      <Block className="sk-line tiny" />
      <ul className="suggest-list">
        {Array.from({ length: count }, (_, i) => (
          <SkeletonListRow key={i} />
        ))}
      </ul>
    </div>
  )
}

export function SkeletonMessages() {
  return (
    <div className="sk-section" aria-busy="true" aria-label="Chargement des messages">
      <SkeletonList count={5} />
    </div>
  )
}

export function SkeletonThread() {
  return (
    <div className="dm-thread sk-section" aria-busy="true" aria-label="Chargement de la conversation">
      <Block className="sk-bubble" />
      <Block className="sk-bubble mine" />
      <Block className="sk-bubble" />
      <Block className="sk-bubble mine" />
    </div>
  )
}
