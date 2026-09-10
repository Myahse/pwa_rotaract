export type IconName = 'home' | 'about' | 'events' | 'donate' | 'contact'

const PATHS: Record<IconName, string> = {
  home: 'M4.2 10.6 12 4.2l7.8 6.4V20a1 1 0 0 1-1 1h-4.4v-6.2H9.6V21H5.2a1 1 0 0 1-1-1Z',
  about:
    'M12 12.1A3.6 3.6 0 1 0 12 5a3.6 3.6 0 0 0 0 7.1ZM5.4 19.8c.7-3.2 3.3-5 6.6-5s5.9 1.8 6.6 5',
  events:
    'M7.2 4.8v2.4M16.8 4.8v2.4M4.8 9.2h14.4M6.2 6.2h11.6A1.4 1.4 0 0 1 19.2 7.6v11a1.4 1.4 0 0 1-1.4 1.4H6.2A1.4 1.4 0 0 1 4.8 18.6v-11A1.4 1.4 0 0 1 6.2 6.2Zm3.2 7.2h.1M12 13.4h.1M14.6 13.4h.1M9.4 16.2h.1M12 16.2h.1',
  donate: 'M12 20.2s-7.2-4.4-7.2-9.1A3.7 3.7 0 0 1 12 8.2a3.7 3.7 0 0 1 7.2 2.9c0 4.7-7.2 9.1-7.2 9.1Z',
  contact:
    'M4.8 7.2A1.6 1.6 0 0 1 6.4 5.6h11.2A1.6 1.6 0 0 1 19.2 7.2v9.6a1.6 1.6 0 0 1-1.6 1.6H6.4A1.6 1.6 0 0 1 4.8 16.8Zm0 0 7.2 5.4 7.2-5.4',
}

export function NavIcon({ name }: { name: IconName }) {
  return (
    <svg className="nav-icon" viewBox="0 0 24 24" width="20" height="20" aria-hidden>
      <path
        fill="none"
        stroke="currentColor"
        strokeWidth="1.75"
        strokeLinecap="round"
        strokeLinejoin="round"
        d={PATHS[name]}
      />
    </svg>
  )
}
