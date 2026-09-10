export type ApiError = {
  error?: string
  message?: string
}

export type AccessRequest = {
  id: string
  club_name: string
  email: string
  known_member: boolean
  status: string
}

export type Donation = {
  id: string
  name: string
  email: string
  amount_xof: number
  status: string
  known_member: boolean
  receipt_mime?: string
  receipt_url?: string
  created_at: string
}

export type PublicEventImage = {
  id: string
  url: string
}

export type SiteGalleryImage = {
  id: string
  caption_fr: string
  caption_en: string
  url: string
  sort_order: number
}

export type FeaturedPostulant = {
  id: string
  period_year: number
  period_month: number
  first_name: string
  last_name: string
  home_club: string
  quote_fr: string
  quote_en: string
  visit_count: number
  clubs_visited: string[]
  flyer_url?: string
  updated_at?: string
}

export type PublicEvent = {
  id: string
  published: boolean
  starts_at: string
  city: string
  venue_fr: string
  venue_en: string
  title_fr: string
  title_en: string
  summary_fr: string
  summary_en: string
  body_fr: string
  body_en: string
  cta_label_fr: string
  cta_label_en: string
  cta_url: string
  flyer_url?: string
  images: PublicEventImage[]
  status: 'upcoming' | 'past'
  created_at: string
  updated_at: string
}
