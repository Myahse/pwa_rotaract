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

export type ClubRole = {
  id: string
  name: string
  description?: string
}

export type UserCommissionAccess = {
  commission_id: string
  commission_name: string
  member_role: 'president' | 'secretary' | 'member'
}

export type ClubAccessSummary = {
  member_role: 'head' | 'member'
  roles: ClubRole[]
  permissions: string[]
  commissions: UserCommissionAccess[]
}

export type ClubRoleAssignment = {
  id: string
  club_id: string
  user_id: string
  club_role_id: string
  role?: ClubRole
}

export type ClubDuePayment = {
  id: string
  club_id: string
  user_id: string
  due_month: string
  amount_xof: number
  status: 'pending' | 'received'
  receipt_url?: string
  created_at: string
  user?: User
}

export type ClubMember = {
  id: string
  club_id: string
  user_id: string
  member_role: 'head' | 'member'
  joined_at: string
  user?: User
}

export type EmailInvite = {
  id: string
  club_id: string
  email: string
  role_id?: string
  expires_at: string
  used_at?: string
  created_at: string
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

export type CommissionMember = {
  id: string
  commission_id: string
  user_id: string
  member_role: 'president' | 'secretary' | 'member'
  user?: User
}

export type ClubMembership = {
  id: string
  club_id: string
  user_id: string
  member_role: 'head' | 'member'
  joined_at: string
  club?: Club
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
  created_at: string
}

export type ChatGroup = {
  id: string
  club_id: string
  commission_id?: string
  group_type: 'club' | 'commission'
  name: string
  created_at: string
}

export type ChatMessage = {
  id: string
  group_id: string
  user_id: string
  content: string
  created_at: string
  user?: User
}

export type BirthdayMember = {
  user_id: string
  first_name: string
  last_name: string
  avatar_url?: string
  message: string
}

export type BirthdayWidget = {
  is_my_birthday: boolean
  title: string
  message: string
  emoji?: string
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
  commission_name?: string
  first_name: string
  last_name: string
  photo_url?: string
  notes?: string
}

export type ClubMandateCommissionGroup = {
  commission_name: string
  president?: ClubMandateAssignment
  secretary?: ClubMandateAssignment
  members: ClubMandateAssignment[]
}

export type ClubMandate = {
  id: string
  name: string
  started_at: string
  ended_at?: string
  is_current: boolean
  bureau: ClubMandateAssignment[]
  commissions: ClubMandateCommissionGroup[]
  members: ClubMandateAssignment[]
}
