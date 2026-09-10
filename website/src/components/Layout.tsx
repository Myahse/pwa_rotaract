import { Link, NavLink, Outlet, useLocation } from 'react-router-dom'
import { useEffect } from 'react'
import { useI18n } from '../i18n'
import { LangSwitcher } from './LangSwitcher'
import { BrandLogo } from './BrandLogo'
import { NavIcon, type IconName } from './NavIcon'
import { SiteFooter } from './SiteFooter'
import { ScrollRail } from './ScrollRail'
import { useSmoothScroll } from '../hooks/useSmoothScroll'
import { scrollToSectionById, scrollToTop } from '../lib/smooth-scroll-api'

type NavItem = {
  to: string
  end?: boolean
  hash?: string
  labelKey: 'navHome' | 'navAbout' | 'navEvents' | 'navDonate' | 'navJoin'
  icon: IconName
}

const items: NavItem[] = [
  { to: '/', end: true, labelKey: 'navHome', icon: 'home' },
  { to: '/#about', hash: 'about', labelKey: 'navAbout', icon: 'about' },
  { to: '/events', labelKey: 'navEvents', icon: 'events' },
  { to: '/donate', labelKey: 'navDonate', icon: 'donate' },
  { to: '/join', labelKey: 'navJoin', icon: 'contact' },
]

export function Layout() {
  const { t } = useI18n()
  const location = useLocation()

  useSmoothScroll(true)

  useEffect(() => {
    if (location.pathname === '/' && location.hash) return
    scrollToTop(true)
  }, [location.pathname])

  useEffect(() => {
    if (location.pathname !== '/') return
    const id = location.hash.slice(1)
    const timer = window.setTimeout(() => {
      if (!id) {
        scrollToTop()
        return
      }
      scrollToSectionById(id)
    }, 60)
    return () => window.clearTimeout(timer)
  }, [location.pathname, location.hash])

  useEffect(() => {
    function onAnchorClick(event: MouseEvent) {
      const link = (event.target as Element | null)?.closest('a[href^="#"]') as HTMLAnchorElement | null
      if (!link) return

      const href = link.getAttribute('href')
      if (!href || href === '#') return

      const id = href.slice(1)
      if (!id || !document.getElementById(id)) return

      event.preventDefault()
      scrollToSectionById(id)
      window.history.replaceState(null, '', `#${id}`)
    }

    document.addEventListener('click', onAnchorClick)
    return () => document.removeEventListener('click', onAnchorClick)
  }, [])

  function itemClass(item: NavItem) {
    if (item.hash) {
      return location.pathname === '/' && location.hash === `#${item.hash}` ? 'active' : ''
    }
    if (item.end) {
      return location.pathname === '/' && !location.hash ? 'active' : ''
    }
    return location.pathname === item.to ? 'active' : ''
  }

  const navLinks = (id: string, withIcons = false) =>
    items.map((item) => (
      <Link key={`${id}-${item.to}`} to={item.to} className={itemClass(item)}>
        {withIcons ? <NavIcon name={item.icon} /> : null}
        <span>{t[item.labelKey]}</span>
      </Link>
    ))

  const fullScreenPage =
    location.pathname === '/' || location.pathname === '/events' || location.pathname === '/donate'

  return (
    <div className="app-shell">
      <ScrollRail onDark={fullScreenPage && location.pathname === '/'} />
      <header className="site-header">
        <NavLink to="/" className="brand-row" end>
          <BrandLogo />
        </NavLink>
        <nav className="site-nav hide-mobile" aria-label="Primary">
          {navLinks('top')}
        </nav>
        <div className="header-end">
          <Link className="header-auth hide-mobile" to="/join">
            {t.navJoin}
          </Link>
          <LangSwitcher />
        </div>
      </header>
      <main
        className={
          location.pathname === '/' || location.pathname === '/events' || location.pathname === '/donate'
            ? 'page page--home'
            : 'page'
        }
      >
        <div key={location.pathname} className="page-appear">
          <Outlet />
        </div>
      </main>
      <SiteFooter />
      <nav className="bottom-nav show-mobile nav-5" aria-label="Mobile">
        {navLinks('bottom', true)}
      </nav>
    </div>
  )
}
