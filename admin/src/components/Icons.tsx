import type { SVGProps } from 'react'

type IconProps = SVGProps<SVGSVGElement>

const base = {
  fill: 'none' as const,
  stroke: 'currentColor',
  strokeWidth: 1.8,
  strokeLinecap: 'round' as const,
  strokeLinejoin: 'round' as const,
}

export function IconDashboard(props: IconProps) {
  return (
    <svg viewBox="0 0 24 24" aria-hidden {...base} {...props}>
      <rect x="3" y="3" width="8" height="8" rx="2" />
      <rect x="13" y="3" width="8" height="5" rx="2" />
      <rect x="13" y="10" width="8" height="11" rx="2" />
      <rect x="3" y="13" width="8" height="8" rx="2" />
    </svg>
  )
}

export function IconClubs(props: IconProps) {
  return (
    <svg viewBox="0 0 24 24" aria-hidden {...base} {...props}>
      <path d="M12 3l8 4.5v9L12 21l-8-4.5v-9L12 3z" />
      <path d="M12 12l8-4.5M12 12v9M12 12L4 7.5" />
    </svg>
  )
}

export function IconAccess(props: IconProps) {
  return (
    <svg viewBox="0 0 24 24" aria-hidden {...base} {...props}>
      <path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" />
      <circle cx="9" cy="7" r="4" />
      <path d="M19 8v6M22 11h-6" />
    </svg>
  )
}

export function IconCard(props: IconProps) {
  return (
    <svg viewBox="0 0 24 24" aria-hidden {...base} {...props}>
      <rect x="3" y="5" width="18" height="14" rx="2" />
      <path d="M3 10h18" />
      <path d="M7 15h4" />
    </svg>
  )
}

export function IconDonate(props: IconProps) {
  return (
    <svg viewBox="0 0 24 24" aria-hidden {...base} {...props}>
      <path d="M12 21s-7-4.4-7-9.1A3.7 3.7 0 0 1 12 9a3.7 3.7 0 0 1 7 2.9c0 4.7-7 9.1-7 9.1Z" />
    </svg>
  )
}

export function IconRegister(props: IconProps) {
  return (
    <svg viewBox="0 0 24 24" aria-hidden {...base} {...props}>
      <path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20" />
      <path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z" />
      <path d="M8 7h8M8 11h6" />
    </svg>
  )
}

export function IconWebsite(props: IconProps) {
  return (
    <svg viewBox="0 0 24 24" aria-hidden {...base} {...props}>
      <circle cx="12" cy="12" r="9" />
      <path d="M3 12h18M12 3a14 14 0 0 1 0 18M12 3a14 14 0 0 0 0 18" />
    </svg>
  )
}

export function IconCalendar(props: IconProps) {
  return (
    <svg viewBox="0 0 24 24" aria-hidden {...base} {...props}>
      <rect x="3" y="5" width="18" height="16" rx="2" />
      <path d="M8 3v4M16 3v4M3 10h18" />
    </svg>
  )
}

export function IconPlus(props: IconProps) {
  return (
    <svg viewBox="0 0 24 24" aria-hidden {...base} {...props}>
      <path d="M12 5v14M5 12h14" />
    </svg>
  )
}

export function IconArrowRight(props: IconProps) {
  return (
    <svg viewBox="0 0 24 24" aria-hidden {...base} {...props}>
      <path d="M5 12h14M13 6l6 6-6 6" />
    </svg>
  )
}

export function IconExternal(props: IconProps) {
  return (
    <svg viewBox="0 0 24 24" aria-hidden {...base} {...props}>
      <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6" />
      <path d="M15 3h6v6M10 14 21 3" />
    </svg>
  )
}

export function IconOrganizer(props: IconProps) {
  return (
    <svg viewBox="0 0 24 24" aria-hidden {...base} {...props}>
      <path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" />
      <circle cx="9" cy="7" r="4" />
      <path d="M22 21v-2a4 4 0 0 0-3-3.87M16 3.13a4 4 0 0 1 0 7.75" />
    </svg>
  )
}

export function IconCampaign(props: IconProps) {
  return (
    <svg viewBox="0 0 24 24" aria-hidden {...base} {...props}>
      <path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z" />
      <path d="m22 6-10 7L2 6" />
    </svg>
  )
}

export function IconMonitor(props: IconProps) {
  return (
    <svg viewBox="0 0 24 24" aria-hidden {...base} {...props}>
      <rect x="2" y="3" width="20" height="14" rx="2" />
      <path d="M8 21h8M12 17v4" />
      <path d="M7 10h2M7 13h6" />
    </svg>
  )
}
