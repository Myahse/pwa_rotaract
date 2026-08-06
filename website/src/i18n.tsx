import { createContext, useContext, useEffect, useMemo, useState, type ReactNode } from 'react'

export type Lang = 'fr' | 'en'

type Messages = (typeof messages)['fr'] | (typeof messages)['en']

type I18nContextValue = {
  lang: Lang
  setLang: (lang: Lang) => void
  t: Messages
}

const STORAGE_KEY = 'rotaract_website_lang'

const messages = {
  fr: {
    brand: 'Rotaract CIV',
    navAbout: 'À propos',
    navMission: 'Mission',
    navContact: 'Contact',
    heroTitle: 'Rotaract Côte d\'Ivoire',
    heroLede:
      'Des jeunes leaders engagés pour le service, le leadership et l\'amitié à travers tout le pays.',
    aboutTitle: 'Qui sommes-nous ?',
    aboutBody:
      'Rotaract est un programme de Rotary International qui rassemble des jeunes adultes pour développer leurs compétences de leadership, mener des projets de service et tisser des liens d\'amitié. En Côte d\'Ivoire, les clubs Rotaract œuvrent au quotidien pour leurs communautés.',
    missionTitle: 'Notre mission',
    missionBody:
      'Servir les communautés, former la prochaine génération de leaders et promouvoir la compréhension mutuelle — avec pour devise « Service avant soi ».',
    pillarsTitle: 'Nos piliers',
    pillars: [
      {
        title: 'Service',
        text: 'Actions concrètes au profit des communautés locales et nationales.',
      },
      {
        title: 'Leadership',
        text: 'Développement des compétences pour diriger avec responsabilité et éthique.',
      },
      {
        title: 'Amitié',
        text: 'Un réseau de jeunes engagés uni par des valeurs partagées.',
      },
    ],
    factsTitle: 'En bref',
    facts: [
      { label: 'Organisation', value: 'Rotaract · Rotary International' },
      { label: 'Pays', value: 'Côte d\'Ivoire' },
      { label: 'Devise', value: 'Service avant soi' },
      { label: 'Public', value: 'Jeunes adultes engagés' },
    ],
    contactTitle: 'Contact',
    contactBody:
      'Pour rejoindre un club ou en savoir plus sur Rotaract en Côte d\'Ivoire, contactez l\'équipe nationale.',
    contactEmail: 'contact@rotaract-civ.local',
    footer: 'Rotaract Côte d\'Ivoire · Service avant soi',
    langFr: 'FR',
    langEn: 'EN',
    langLabel: 'Langue',
  },
  en: {
    brand: 'Rotaract CIV',
    navAbout: 'About',
    navMission: 'Mission',
    navContact: 'Contact',
    heroTitle: 'Rotaract Côte d\'Ivoire',
    heroLede:
      'Young leaders dedicated to service, leadership and fellowship across the country.',
    aboutTitle: 'Who we are',
    aboutBody:
      'Rotaract is a Rotary International program that brings young adults together to build leadership skills, run service projects and form lasting friendships. In Côte d\'Ivoire, Rotaract clubs work every day for their communities.',
    missionTitle: 'Our mission',
    missionBody:
      'Serve communities, develop the next generation of leaders and foster mutual understanding — guided by the motto “Service Above Self”.',
    pillarsTitle: 'Our pillars',
    pillars: [
      {
        title: 'Service',
        text: 'Hands-on projects that benefit local and national communities.',
      },
      {
        title: 'Leadership',
        text: 'Building skills to lead with responsibility and integrity.',
      },
      {
        title: 'Fellowship',
        text: 'A network of engaged young people united by shared values.',
      },
    ],
    factsTitle: 'At a glance',
    facts: [
      { label: 'Organization', value: 'Rotaract · Rotary International' },
      { label: 'Country', value: 'Côte d\'Ivoire' },
      { label: 'Motto', value: 'Service Above Self' },
      { label: 'Audience', value: 'Engaged young adults' },
    ],
    contactTitle: 'Contact',
    contactBody:
      'To join a club or learn more about Rotaract in Côte d\'Ivoire, reach out to the national team.',
    contactEmail: 'contact@rotaract-civ.local',
    footer: 'Rotaract Côte d\'Ivoire · Service Above Self',
    langFr: 'FR',
    langEn: 'EN',
    langLabel: 'Language',
  },
} as const

const I18nContext = createContext<I18nContextValue | null>(null)

function detectLang(): Lang {
  try {
    const saved = localStorage.getItem(STORAGE_KEY)
    if (saved === 'fr' || saved === 'en') return saved
  } catch {
    /* ignore */
  }
  const nav = typeof navigator !== 'undefined' ? navigator.language.toLowerCase() : 'fr'
  return nav.startsWith('fr') ? 'fr' : 'en'
}

export function I18nProvider({ children }: { children: ReactNode }) {
  const [lang, setLangState] = useState<Lang>(() => detectLang())

  useEffect(() => {
    document.documentElement.lang = lang
    try {
      localStorage.setItem(STORAGE_KEY, lang)
    } catch {
      /* ignore */
    }
  }, [lang])

  const value = useMemo(
    () => ({
      lang,
      setLang: setLangState,
      t: messages[lang],
    }),
    [lang],
  )

  return <I18nContext.Provider value={value}>{children}</I18nContext.Provider>
}

export function useI18n() {
  const ctx = useContext(I18nContext)
  if (!ctx) throw new Error('useI18n must be used within I18nProvider')
  return ctx
}
