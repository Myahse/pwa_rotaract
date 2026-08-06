import { LangSwitcher } from '../components/LangSwitcher'
import { useI18n } from '../i18n'

export function HomePage() {
  const { t } = useI18n()

  return (
    <div className="vitrine">
      <header className="site-header">
        <div className="brand-row">
          <span className="brand-dot" aria-hidden />
          <strong>{t.brand}</strong>
        </div>
        <nav className="site-nav" aria-label="Primary">
          <a href="#about">{t.navAbout}</a>
          <a href="#mission">{t.navMission}</a>
          <a href="#contact">{t.navContact}</a>
        </nav>
        <LangSwitcher />
      </header>

      <main>
        <section className="vitrine-hero">
          <span className="brand-dot lg" aria-hidden />
          <h1>{t.heroTitle}</h1>
          <p className="lede">{t.heroLede}</p>
        </section>

        <section id="about" className="section">
          <h2>{t.aboutTitle}</h2>
          <p>{t.aboutBody}</p>
        </section>

        <section id="mission" className="section">
          <h2>{t.missionTitle}</h2>
          <p>{t.missionBody}</p>
        </section>

        <section className="section">
          <h2>{t.pillarsTitle}</h2>
          <div className="pillar-grid">
            {t.pillars.map((p) => (
              <article key={p.title} className="pillar">
                <h3>{p.title}</h3>
                <p>{p.text}</p>
              </article>
            ))}
          </div>
        </section>

        <section className="section facts">
          <h2>{t.factsTitle}</h2>
          <dl className="fact-list">
            {t.facts.map((f) => (
              <div key={f.label} className="fact">
                <dt>{f.label}</dt>
                <dd>{f.value}</dd>
              </div>
            ))}
          </dl>
        </section>

        <section id="contact" className="section">
          <h2>{t.contactTitle}</h2>
          <p>{t.contactBody}</p>
          <a className="contact-mail" href={`mailto:${t.contactEmail}`}>
            {t.contactEmail}
          </a>
        </section>
      </main>

      <footer className="vitrine-foot">{t.footer}</footer>
    </div>
  )
}
