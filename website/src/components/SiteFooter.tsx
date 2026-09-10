import { Link } from 'react-router-dom'
import { CLUB_CAMPUS, CLUB_EMAIL, CLUB_NAME } from '../lib/club'
import { useI18n } from '../i18n'

export function SiteFooter() {
  const { t } = useI18n()
  const year = new Date().getFullYear()

  return (
    <footer className="site-footer hide-mobile">
      <div className="site-footer-inner">
        <div className="site-footer-brand">
          <p className="site-footer-name">{CLUB_NAME}</p>
          <p className="site-footer-motto">{t.footerMotto}</p>
          <p className="site-footer-campus">{CLUB_CAMPUS}</p>
        </div>

        <nav className="site-footer-nav" aria-label={t.footerNavLabel}>
          <p className="site-footer-label">{t.footerNavLabel}</p>
          <Link to="/">{t.navHome}</Link>
          <Link to="/#about">{t.navAbout}</Link>
          <Link to="/events">{t.navEvents}</Link>
          <Link to="/donate">{t.navDonate}</Link>
          <Link to="/join">{t.navJoin}</Link>
        </nav>

        <div className="site-footer-contact">
          <p className="site-footer-label">{t.footerContactLabel}</p>
          <a href={`mailto:${CLUB_EMAIL}`}>{CLUB_EMAIL}</a>
        </div>

        <p className="site-footer-copy">
          © {year} {CLUB_NAME}. {t.footerRights}
        </p>
      </div>
    </footer>
  )
}
