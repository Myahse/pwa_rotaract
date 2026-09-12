export type User = {
  id: string
  email: string
  first_name: string
  last_name: string
  phone?: string
  birth_date?: string
  profession?: string
  member_since?: string
  avatar_url?: string
  is_admin: boolean
  is_active: boolean
  created_at: string
  updated_at: string
}

export type Club = {
  id: string
  name: string
  slug: string
  invite_code: string
  description?: string
  country?: string
  city?: string
  commune?: string
  logo_url?: string
  founded_at?: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export type CreateClubResult = {
  club: Club
  president: User
}

export type ClubMember = {
  id: string
  club_id: string
  user_id: string
  member_role: 'head' | 'member'
  joined_at: string
  user?: User
}

export type ClubDiaryEntryType = 'parrain' | 'president' | 'member'

export type ClubDiaryEntry = {
  id: string
  club_id: string
  parent_id?: string
  entry_type: ClubDiaryEntryType
  user_id?: string
  first_name: string
  last_name: string
  organization?: string
  photo_url?: string
  started_at?: string
  ended_at?: string
  notes?: string
  sort_order: number
  created_at: string
  updated_at: string
}

export type ClubDiaryTreeNode = ClubDiaryEntry & {
  children: ClubDiaryTreeNode[]
}

export type ClubDiaryResponse = {
  entries: ClubDiaryEntry[]
  tree: ClubDiaryTreeNode[]
  parrains: ClubDiaryEntry[]
  mandates: ClubMandate[]
}

export type ClubMandateRole =
  | 'president'
  | 'vice_president'
  | 'secretary'
  | 'treasurer'
  | 'commission_president'
  | 'commission_secretary'
  | 'commission_member'
  | 'member'

export type ClubMandateAssignment = {
  id: string
  mandate_id: string
  role: ClubMandateRole
  commission_id?: string
  commission_name?: string
  user_id?: string
  first_name: string
  last_name: string
  photo_url?: string
  notes?: string
  sort_order: number
  created_at: string
  updated_at: string
}

export type ClubMandateCommissionGroup = {
  commission_id?: string
  commission_name: string
  president?: ClubMandateAssignment
  secretary?: ClubMandateAssignment
  members: ClubMandateAssignment[]
}

export type ClubMandate = {
  id: string
  club_id: string
  name: string
  started_at: string
  ended_at?: string
  is_current: boolean
  notes?: string
  sort_order: number
  bureau: ClubMandateAssignment[]
  commissions: ClubMandateCommissionGroup[]
  members: ClubMandateAssignment[]
  created_at: string
  updated_at: string
}

export type Commission = {
  id: string
  club_id: string
  code?: string
  name: string
  description?: string
  is_system: boolean
  created_at: string
}

export type LoginResponse = {
  access_token: string
  expires_at: string
  user: User
}

export type ApiError = {
  error: string
  message: string
}

export type AccessRequest = {
  id: string
  club_id?: string
  club_name: string
  email: string
  first_name: string
  last_name: string
  status: string
  known_member?: boolean
  existing_user_id?: string
  created_at: string
}

export type Donation = {
  id: string
  user_id?: string
  name: string
  email: string
  amount_xof: number
  status: 'pending' | 'received'
  known_member: boolean
  receipt_mime?: string
  receipt_url?: string
  created_at: string
}

export type PublicEventImage = {
  id: string
  url: string
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

export type SiteGalleryImage = {
  id: string
  published: boolean
  caption_fr: string
  caption_en: string
  url?: string
  sort_order: number
  created_at: string
  updated_at: string
}

export type FeaturedPostulant = {
  id: string
  period_year: number
  period_month: number
  published: boolean
  first_name: string
  last_name: string
  home_club: string
  quote_fr: string
  quote_en: string
  visit_count: number
  clubs_visited: string[]
  flyer_url?: string
  updated_at: string
}

export type ClubRegistrationStatus = 'pending' | 'approved' | 'rejected'

export type ClubRegistrationRequest = {
  id: string
  club_name: string
  contact_email: string
  contact_first_name: string
  contact_last_name: string
  phone?: string
  description?: string
  country: string
  city: string
  commune: string
  founded_at?: string
  message?: string
  status: ClubRegistrationStatus
  reviewed_by?: string
  review_note?: string
  reviewed_at?: string
  created_club_id?: string
  access_token_expires_at?: string
  approved_slug?: string
  created_at: string
}

export type ApproveClubRegistrationResult = {
  request: ClubRegistrationRequest
  message: string
  expires_at: string
}

export type MemberCard = {
  id: string
  club_id: string
  user_id: string
  card_number: string
  issued_at: string
  sent_at?: string
  sent_by?: string
  user?: User
  club?: Club
}

export type MemberCardListResponse = {
  items: MemberCard[]
  pending_count: number
}
