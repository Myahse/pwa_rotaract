import { Link } from 'react-router-dom'
import { useEffect, useState } from 'react'
import { BrandLogo } from '../components/BrandLogo'
import { RevealSection } from '../components/RevealSection'
import { apiRequest } from '../api/client'
import type { FeaturedPostulant, SiteGalleryImage } from '../api/types'
import { useI18n } from '../i18n'
import { formatPostulantPeriod } from '../lib/postulantPeriod'

export function HomePage() {
  const { t, lang } = useI18n()
  const [gallery, setGallery] = useState<SiteGalleryImage[]>([])
  const [postulant, setPostulant] = useState<FeaturedPostulant | null>(null)

  useEffect(() => {
    apiRequest<SiteGalleryImage[]>('/gallery')
      .then((data) => setGallery(data ?? []))
      .catch(() => setGallery([]))
  }, [])

  useEffect(() => {
    apiRequest<FeaturedPostulant>('/featured-postulant')
      .then((data) => setPostulant(data))
      .catch(() => setPostulant(null))
  }, [])

  const postulantName = postulant
    ? `${postulant.first_name} ${postulant.last_name}`.trim()
    : ''
  const postulantQuote = postulant
    ? lang === 'fr'
      ? postulant.quote_fr || postulant.quote_en
      : postulant.quote_en || postulant.quote_fr
    : ''
  return (
    <div className="home-scroll">
      <RevealSection id="hero" className="home-section home-section--hero" immediate>
        <div className="home-section-inner home-hero">
          <BrandLogo hero onDark />
          <p className="home-eyebrow">{t.heroEyebrow}</p>
          <h1>{t.heroTitle}</h1>
          <p className="lede">{t.heroLede}</p>
          <div className="cta-row center">
            <Link to="/events" className="btn-primary btn-on-hero">
              {t.heroEventsCta}
            </Link>
            <a href="#about" className="btn-outline btn-on-hero-outline">
              {t.heroAboutCta}
            </a>
          </div>
          <a className="home-scroll-hint" href="#about" aria-label={t.scrollHint}>
            <span>{t.scrollHint}</span>
            <svg viewBox="0 0 24 24" width="20" height="20" aria-hidden="true">
              <path
                d="M12 16.5l-6-6 1.4-1.4 4.6 4.6 4.6-4.6 1.4 1.4-6 6z"
                fill="currentColor"
              />
            </svg>
          </a>
        </div>
      </RevealSection>

      <RevealSection id="about" className="home-section">
        <div className="home-section-inner">
          <p className="home-eyebrow">{t.aboutKicker}</p>
          <h2>{t.aboutTitle}</h2>
          <p className="home-body">{t.aboutBody}</p>
        </div>
      </RevealSection>

      <RevealSection id="mission" className="home-section home-section--alt">
        <div className="home-section-inner">
          <p className="home-eyebrow">{t.missionKicker}</p>
          <h2>{t.missionTitle}</h2>
          <p className="mission-motto">{t.missionQuote}</p>
          <p className="home-body">{t.missionBody}</p>
          <ul className="mission-list">
            {t.pillars.map((p) => (
              <li key={p.title} className="mission-item">
                <h3>{p.title}</h3>
                <p>{p.text}</p>
              </li>
            ))}
          </ul>
        </div>
      </RevealSection>

      <RevealSection id="gallery" className="home-section">
        <div className="home-section-inner home-section-inner--wide">
          <p className="home-eyebrow">{t.galleryKicker}</p>
          <h2>{t.galleryTitle}</h2>
          <p className="home-body">{t.galleryBody}</p>
          {gallery.length === 0 ? (
            <p className="home-empty">{t.galleryEmpty}</p>
          ) : (
            <div className="home-gallery">
              {gallery.map((item, index) => {
                const caption = lang === 'fr' ? item.caption_fr || item.caption_en : item.caption_en || item.caption_fr
                return (
                  <figure key={item.id} className={`home-gallery-item home-gallery-item--${(index % 6) + 1}`}>
                    <img src={item.url} alt={caption || t.galleryTitle} loading="lazy" decoding="async" />
                    {caption ? <figcaption>{caption}</figcaption> : null}
                  </figure>
                )
              })}
            </div>
          )}
        </div>
      </RevealSection>

      <RevealSection id="postulant" className="home-section home-section--postulant">
        <div className="home-section-inner">
          <p className="home-eyebrow">{t.postulantKicker}</p>
          <h2>{t.postulantTitle}</h2>
          {postulant ? (
            <p className="home-postulant-period">
              {t.postulantPeriodHint.replace(
                '{period}',
                formatPostulantPeriod(postulant.period_year, postulant.period_month, lang),
              )}
            </p>
          ) : null}
          {postulant ? (
            <div className="home-postulant">
              {postulant.flyer_url ? (
                <img
                  src={`${postulant.flyer_url}${postulant.flyer_url.includes('?') ? '&' : '?'}v=${encodeURIComponent(postulant.updated_at ?? '')}`}
                  alt={postulantName || t.postulantTitle}
                  className="home-postulant-flyer"
                />
              ) : null}
              <div className="home-postulant-details">
                {postulantName ? <p className="home-postulant-name">{postulantName}</p> : null}
                {postulant.home_club ? (
                  <p className="home-postulant-meta">{postulant.home_club}</p>
                ) : null}
                {postulantQuote ? <p className="home-postulant-quote">{postulantQuote}</p> : null}
                <div className="home-postulant-facts">
                  <p>
                    <span>{t.postulantVisitsLabel}</span>
                    {postulant.visit_count}
                  </p>
                  <p>
                    <span>{t.postulantClubsLabel}</span>
                    {postulant.clubs_visited?.length
                      ? postulant.clubs_visited.join(' · ')
                      : t.postulantNoClubs}
                  </p>
                </div>
              </div>
            </div>
          ) : (
            <p className="home-empty">{t.postulantEmpty}</p>
          )}
        </div>
      </RevealSection>
    </div>
  )
}
