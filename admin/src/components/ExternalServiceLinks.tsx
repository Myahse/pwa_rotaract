import { tombolaServices } from '../config/externalServices'
import { IconCampaign, IconExternal, IconMonitor, IconOrganizer } from './Icons'

const icons = {
  organizer: IconOrganizer,
  campaign: IconCampaign,
  monitor: IconMonitor,
} as const

type Props = {
  variant?: 'sidebar' | 'grid'
}

export function ExternalServiceLinks({ variant = 'sidebar' }: Props) {
  if (variant === 'grid') {
    return (
      <div className="quick-grid">
        {tombolaServices.map((service) => {
          const Icon = icons[service.id as keyof typeof icons]
          return (
            <a
              key={service.id}
              href={service.href}
              target="_blank"
              rel="noopener noreferrer"
              className="quick-link external-link"
            >
              <span className="quick-icon"><Icon /></span>
              <span>
                {service.label}
                <small>{service.description}</small>
              </span>
              <IconExternal className="external-link-icon" />
            </a>
          )
        })}
      </div>
    )
  }

  return (
    <div className="external-nav">
      <p className="external-nav-label">Tombola IUGB</p>
      {tombolaServices.map((service) => {
        const Icon = icons[service.id as keyof typeof icons]
        return (
          <a
            key={service.id}
            href={service.href}
            target="_blank"
            rel="noopener noreferrer"
            className="external-nav-link"
          >
            <Icon />
            <span>{service.label}</span>
            <IconExternal className="external-link-icon" />
          </a>
        )
      })}
    </div>
  )
}
