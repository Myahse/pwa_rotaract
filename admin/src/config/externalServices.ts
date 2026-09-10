export type ExternalService = {
  id: string
  label: string
  short: string
  description: string
  href: string
}

const organizerUrl = import.meta.env.VITE_TOMBOLA_ORGANIZER_URL || 'https://organisateurs.rotaractiugb.com'
const campaignUrl = import.meta.env.VITE_TOMBOLA_CAMPAIGN_URL || 'https://campagnes.rotaractiugb.com'
const monitorUrl = import.meta.env.VITE_TOMBOLA_MONITOR_URL || 'https://monitor.rotaractiugb.com'

export const tombolaServices: ExternalService[] = [
  {
    id: 'organizer',
    label: 'Organisateurs',
    short: 'Org.',
    description: 'Espace organisateurs tombola (tickets, paiements, tirages)',
    href: organizerUrl,
  },
  {
    id: 'campaign',
    label: 'Campagnes e-mail',
    short: 'E-mail',
    description: 'Création et envoi des campagnes e-mail Brevo',
    href: campaignUrl,
  },
  {
    id: 'monitor',
    label: 'Monitoring',
    short: 'Monitor',
    description: 'Salle de surveillance des examens en direct',
    href: monitorUrl,
  },
]
