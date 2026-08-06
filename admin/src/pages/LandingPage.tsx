import { Link } from 'react-router-dom'

export function LandingPage() {
  return (
    <div className="landing">
      <section className="landing-hero">
        <span className="brand-dot lg" aria-hidden />
        <h1>Rotaract CIV</h1>
        <p className="lede">
          Console d&apos;administration pour piloter les clubs, les demandes et la vie de la plateforme.
        </p>
        <div className="landing-cta">
          <Link to="/login" className="btn-primary">Se connecter</Link>
        </div>
      </section>
      <footer className="landing-foot">
        Plateforme officielle · Accès réservé aux administrateurs
      </footer>
    </div>
  )
}
