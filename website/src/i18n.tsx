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
    brand: 'Rotaract IUGB',
    navHome: 'Accueil',
    navAbout: 'À propos',
    navMission: 'Mission',
    navEvents: 'Événements',
    navDonate: 'Dons',
    navContact: 'Contact',
    navJoin: 'Rejoindre',
    heroEyebrow: 'Rotaract IUGB Club',
    heroTitle: 'Servir. Diriger. Grandir ensemble.',
    heroLede:
      'Le club Rotaract de l\'IUGB : service, leadership et amitié au campus et au-delà.',
    heroEventsCta: 'Voir les événements',
    heroAboutCta: 'Qui sommes-nous',
    scrollHint: 'Défiler',
    aboutKicker: 'Le club',
    missionKicker: 'Ce qu\'on fait',
    galleryKicker: 'En images',
    galleryTitle: 'Galerie',
    galleryBody:
      'Moments forts, actions de service et vie de club — un aperçu de l\'énergie du Rotaract IUGB.',
    galleryEmpty: 'La galerie sera bientôt disponible.',
    postulantKicker: 'Portrait',
    postulantTitle: 'Meilleur postulant',
    postulantPeriodHint: 'Mois de {period}',
    postulantVisitsLabel: 'Visites',
    postulantClubsLabel: 'Clubs visités',
    postulantNoClubs: '—',
    postulantEmpty: 'Le meilleur postulant du mois dernier sera publié ici.',
    homeEventsTitle: 'Prochains événements',
    homeEventsCta: 'Tous les événements',
    eventsKicker: 'Agenda',
    eventsTitle: 'Événements',
    eventsLede:
      'Forums, journées de service et activités du Rotaract IUGB Club.',
    eventsFeatured: 'À la une',
    eventsUpcoming: 'À venir',
    eventsPast: 'Passés',
    eventsEmpty: 'Aucun événement à afficher pour le moment.',
    eventsLoading: 'Chargement des événements…',
    eventsLoadError: 'Impossible de charger les événements.',
    eventUpcoming: 'À venir',
    eventPast: 'Passé',
    eventBack: 'Tous les événements',
    eventGallery: 'Photos',
    eventNotFound: 'Cet événement n’est plus publié.',
    donateKicker: 'Rotaract IUGB Club',
    donateTitle: 'Faire un don',
    donateLede:
      'Soutenez les actions du club. Les dons passent par Doaty Délice, partenaire du Rotaract IUGB Club.',
    donateHowTitle: 'Comment donner',
    donateWaveCta: 'Donner via Wave',
    donateWave: 'Wave',
    donateWaveText: 'Ouvrez Wave et choisissez le montant que vous souhaitez offrir.',
    donateCash: 'Espèces',
    donateCashText: 'Vous pouvez aussi remettre un don en espèces à un organisateur du club.',
    donateRefTitle: 'Envoyer le reçu de paiement',
    donateRefHelp:
      'Après le paiement Wave, prenez une photo ou un PDF du reçu, indiquez le montant, et envoyez-le au club.',
    donateName: 'Nom',
    donateEmailLabel: 'E-mail',
    donateAmount: 'Montant donné (F CFA)',
    donateSend: 'Envoyer le reçu',
    donateSending: 'Envoi…',
    donateSent: 'Reçu envoyé. Le club le verra dans la gestion.',
    donateSentKnown: 'Reçu envoyé et lié à votre compte club / tombola.',
    donateInvalid: 'Vérifiez le nom, le montant et le fichier du reçu.',
    receiptFileLabel: 'Photo ou document',
    receiptPickHint: 'Appuyez pour choisir une photo ou un PDF',
    receiptInvalidType: 'Format non accepté. Utilisez JPG, PNG, WebP ou PDF.',
    receiptTooLarge: 'Fichier trop lourd (5 Mo max pour une image, 10 Mo pour un PDF).',
    homeDonateTitle: 'Soutenir le club',
    homeDonateText: 'Vous pouvez aussi faire un don via Wave, au profit du club.',
    homeDonateCta: 'Faire un don',
    aboutTitle: 'Qui sommes-nous ?',
    aboutBody:
      'Le Rotaract IUGB Club rassemble les étudiants et jeunes diplômés de l\'IUGB autour du service, du leadership et de l\'amitié. Nous menons des actions concrètes sur le campus et dans nos communautés, dans l\'esprit de Rotary International.',
    missionTitle: 'Au cœur du club',
    missionQuote: '« Service avant soi »',
    missionBody:
      'On sort du campus pour agir, on apprend en menant des projets, et on construit une vraie équipe. C\'est le quotidien du Rotaract IUGB à Grand-Bassam.',
    pillarsTitle: 'Nos piliers',
    pillars: [
      {
        title: 'Service',
        text: 'Dons, caravanes, actions solidaires… Nos projets touchent le campus et les quartiers alentour.',
      },
      {
        title: 'Leadership',
        text: 'Bureau, commissions, organisation d\'événements : chacun y trouve sa place et monte en compétence.',
      },
      {
        title: 'Amitié',
        text: 'Afterworks, moments entre membres, entraide au quotidien — le club, c\'est aussi ça.',
      },
    ],
    contactTitle: 'Contact',
    contactBody:
      'Pour rejoindre le club ou en savoir plus sur nos activités, contactez l\'équipe du Rotaract IUGB.',
    contactEmail: 'rotaractiugb@gmail.com',
    footerMotto: 'Service avant soi',
    footerNavLabel: 'Navigation',
    footerContactLabel: 'Contact',
    footerRights: 'Tous droits réservés.',
    joinKicker: 'Adhésion',
    joinTitle: 'Rejoindre le Rotaract IUGB',
    joinLede:
      'Demandez à intégrer le Rotaract IUGB Club. Si votre e-mail existe déjà (tombola ou gestion club), nous rattacherons votre compte.',
    joinClubFixed: 'Rotaract IUGB Club',
    joinFirstName: 'Prénom',
    joinLastName: 'Nom',
    joinEmail: 'E-mail',
    joinClub: 'Nom du club',
    joinProfession: 'Profession (optionnel)',
    joinPassword: 'Mot de passe (nouveaux comptes)',
    joinPasswordHint:
      'Laissez vide si vous avez déjà un compte tombola ou club. Sinon, 8 caractères minimum.',
    joinSubmit: 'Envoyer la demande',
    joinSending: 'Envoi…',
    joinSentTitle: 'Demande envoyée',
    joinSentBody: 'Le bureau du Rotaract IUGB examinera votre demande.',
    joinSentKnown:
      'Votre e-mail est déjà dans la base club (tombola). Après validation, vous rejoindrez le club avec ce compte. Utilisez « mot de passe oublié » pour vous connecter si besoin.',
    joinLoginCta: 'Ouvrir la gestion de club',
    joinAlreadyMember: 'Cet e-mail est déjà membre de ce club. Connectez-vous à la gestion de club.',
    joinPending: 'Une demande est déjà en attente pour cet e-mail.',
    joinPasswordRequired: 'Un mot de passe d’au moins 8 caractères est requis pour un nouveau compte.',
    joinError: 'Envoi impossible. Réessayez.',
    langFr: 'FR',
    langEn: 'EN',
    langLabel: 'Langue',
  },
  en: {
    brand: 'Rotaract IUGB',
    navHome: 'Home',
    navAbout: 'About',
    navMission: 'Mission',
    navEvents: 'Events',
    navDonate: 'Donate',
    navContact: 'Contact',
    navJoin: 'Join',
    heroEyebrow: 'Rotaract IUGB Club',
    heroTitle: 'Serve. Lead. Grow together.',
    heroLede:
      'The Rotaract club at IUGB — service, leadership and fellowship on campus and beyond.',
    heroEventsCta: 'See events',
    heroAboutCta: 'Who we are',
    scrollHint: 'Scroll',
    aboutKicker: 'The club',
    missionKicker: 'What we do',
    galleryKicker: 'In pictures',
    galleryTitle: 'Gallery',
    galleryBody:
      'Highlights from service projects, training and club life at Rotaract IUGB.',
    galleryEmpty: 'The gallery will be available soon.',
    postulantKicker: 'Portrait',
    postulantTitle: 'Best postulant',
    postulantPeriodHint: '{period}',
    postulantVisitsLabel: 'Visits',
    postulantClubsLabel: 'Clubs visited',
    postulantNoClubs: '—',
    postulantEmpty: 'Last month’s best postulant will appear here.',
    homeEventsTitle: 'Upcoming events',
    homeEventsCta: 'All events',
    eventsKicker: 'Calendar',
    eventsTitle: 'Events',
    eventsLede: 'Forums, days of service and activities from Rotaract IUGB Club.',
    eventsFeatured: 'Featured',
    eventsUpcoming: 'Upcoming',
    eventsPast: 'Past',
    eventsEmpty: 'No events to show right now.',
    eventsLoading: 'Loading events…',
    eventsLoadError: 'Could not load events.',
    eventUpcoming: 'Upcoming',
    eventPast: 'Past',
    eventBack: 'All events',
    eventGallery: 'Photos',
    eventNotFound: 'This event is no longer published.',
    donateKicker: 'Rotaract IUGB Club',
    donateTitle: 'Donate',
    donateLede:
      'Support the club’s work. Donations go through Doaty Délice, a Rotaract IUGB Club partner.',
    donateHowTitle: 'How to give',
    donateWaveCta: 'Donate with Wave',
    donateWave: 'Wave',
    donateWaveText: 'Open Wave and choose any amount you wish to give.',
    donateCash: 'Cash',
    donateCashText: 'You can also give cash to a club organizer.',
    donateRefTitle: 'Send payment receipt',
    donateRefHelp:
      'After paying with Wave, take a photo or PDF of the receipt, enter the amount, and send it to the club.',
    donateName: 'Name',
    donateEmailLabel: 'Email',
    donateAmount: 'Amount given (XOF)',
    donateSend: 'Send receipt',
    donateSending: 'Sending…',
    donateSent: 'Receipt sent. The club will see it in club management.',
    donateSentKnown: 'Receipt sent and linked to your club / tombola account.',
    donateInvalid: 'Check the name, amount, and receipt file.',
    receiptFileLabel: 'Photo or document',
    receiptPickHint: 'Tap to choose a photo or PDF',
    receiptInvalidType: 'Unsupported format. Use JPG, PNG, WebP, or PDF.',
    receiptTooLarge: 'File is too large (5 MB max for images, 10 MB for PDF).',
    homeDonateTitle: 'Support the club',
    homeDonateText: 'You can also donate via Wave to support the club.',
    homeDonateCta: 'Donate',
    aboutTitle: 'Who we are',
    aboutBody:
      'Rotaract IUGB Club brings IUGB students and young graduates together around service, leadership and fellowship. We run hands-on projects on campus and in our communities, in the spirit of Rotary International.',
    missionTitle: 'What drives us',
    missionQuote: '“Service Above Self”',
    missionBody:
      'We step off campus to take action, learn by running projects, and build a real team. That’s everyday life at Rotaract IUGB in Grand-Bassam.',
    pillarsTitle: 'Our pillars',
    pillars: [
      {
        title: 'Service',
        text: 'Donations, outreach drives, community work… Our projects reach the campus and nearby neighbourhoods.',
      },
      {
        title: 'Leadership',
        text: 'Board roles, commissions, event planning — everyone finds a place to grow and take responsibility.',
      },
      {
        title: 'Fellowship',
        text: 'Afterworks, time together, day-to-day support — the club is also about people.',
      },
    ],
    contactTitle: 'Contact',
    contactBody:
      'To join the club or learn more about our activities, contact the Rotaract IUGB team.',
    contactEmail: 'rotaractiugb@gmail.com',
    footerMotto: 'Service Above Self',
    footerNavLabel: 'Navigation',
    footerContactLabel: 'Contact',
    footerRights: 'All rights reserved.',
    joinKicker: 'Membership',
    joinTitle: 'Join Rotaract IUGB',
    joinLede:
      'Apply to join Rotaract IUGB Club. If your email already exists (tombola or club app), we will link your existing account.',
    joinClubFixed: 'Rotaract IUGB Club',
    joinFirstName: 'First name',
    joinLastName: 'Last name',
    joinEmail: 'Email',
    joinClub: 'Club name',
    joinProfession: 'Profession (optional)',
    joinPassword: 'Password (new accounts)',
    joinPasswordHint:
      'Leave empty if you already have a tombola or club account. Otherwise use at least 8 characters.',
    joinSubmit: 'Submit request',
    joinSending: 'Sending…',
    joinSentTitle: 'Request sent',
    joinSentBody: 'The Rotaract IUGB board will review your request.',
    joinSentKnown:
      'Your email is already in the club database (tombola). After approval you will join with that account. Use “forgot password” to sign in if needed.',
    joinLoginCta: 'Open club management',
    joinAlreadyMember: 'This email is already a member of this club. Sign in to club management.',
    joinPending: 'A request is already pending for this email.',
    joinPasswordRequired: 'A password of at least 8 characters is required for a new account.',
    joinError: 'Could not send. Please try again.',
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
