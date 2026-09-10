import { useState } from 'react'
import { CLUB_NAME } from '../lib/club'

export function BrandLogo({ hero = false, onDark = false }: { hero?: boolean; onDark?: boolean }) {
  const className = hero ? 'brand-logo hero' : 'brand-logo'
  const [useWhite, setUseWhite] = useState(true)
  const logoSrc = onDark ? '/logo-white.png' : '/logo.png'
  return (
    <picture>
      {useWhite && !onDark ? <source srcSet="/logo-white.png" media="(prefers-color-scheme: dark)" /> : null}
      <img
        src={logoSrc}
        alt={CLUB_NAME}
        className={className}
        width={751}
        height={284}
        decoding="async"
        fetchPriority="high"
        onError={() => setUseWhite(false)}
      />
    </picture>
  )
}
