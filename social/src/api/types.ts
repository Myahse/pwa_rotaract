export type User = {
  id: string
  email: string
  first_name: string
  last_name: string
  avatar_url?: string
  is_admin: boolean
  is_active: boolean
}

export type LoginResponse = {
  access_token: string
  expires_at: string
  user: User
}

export type ClubMembership = {
  club_id: string
  member_role: string
  club?: { id: string; name: string }
}

export type SocialAuthor = {
  id: string
  first_name: string
  last_name: string
  avatar_url?: string
}

export type SocialClubTag = {
  id: string
  name: string
}

export type SocialMedia = {
  id: string
  post_id: string
  kind: 'image' | 'video'
  url: string
  sort_order: number
}

export type SocialPost = {
  id: string
  author_id: string
  club_id?: string
  reposted_post_id?: string
  quote_body?: string
  original?: SocialPost
  body: string
  created_at: string
  updated_at: string
  author?: SocialAuthor
  club?: SocialClubTag
  media?: SocialMedia[]
  comment_count: number
  reaction_count: number
  reacted_by_me: boolean
  author_followed_by_me: boolean
}

export type SocialComment = {
  id: string
  post_id: string
  author_id: string
  parent_id?: string
  body: string
  created_at: string
  author?: SocialAuthor
  replies?: SocialComment[]
}

export type FriendshipStatus = 'none' | 'pending' | 'accepted' | 'declined'

export type FollowProfile = {
  id: string
  first_name: string
  last_name: string
  avatar_url?: string
  followed_by_me: boolean
  followers_count: number
  friendship_status?: FriendshipStatus
}

export type FriendProfile = {
  id: string
  first_name: string
  last_name: string
  avatar_url?: string
  friendship_status: FriendshipStatus
  incoming_request: boolean
  friends_since?: string
}

export type SocialUserProfile = {
  id: string
  first_name: string
  last_name: string
  avatar_url?: string
  followers_count: number
  following_count: number
  friends_count: number
  followed_by_me: boolean
  friendship_status: FriendshipStatus
  incoming_request: boolean
  is_me: boolean
  clubs?: SocialClubTag[]
}

export type SocialMessage = {
  id: string
  conversation_id: string
  sender_id: string
  body: string
  shared_comment_id?: string
  shared_post_id?: string
  created_at: string
  sender?: SocialAuthor
  shared_comment?: SocialComment
  shared_post?: SocialPost
}

export type SocialConversation = {
  id: string
  updated_at: string
  peer?: SocialAuthor
  last_message?: SocialMessage
  unread_count: number
}

export type SocialGroupPrivacy = 'open' | 'approval'
export type SocialGroupRole = 'admin' | 'member'
export type SocialGroupJoinStatus = 'none' | 'pending' | 'member' | 'rejected'

export type SocialGroup = {
  id: string
  name: string
  description: string
  privacy: SocialGroupPrivacy
  created_by: string
  created_at: string
  updated_at: string
  member_count: number
  my_role?: SocialGroupRole
  join_status: SocialGroupJoinStatus
  pending_count?: number
  creator?: SocialAuthor
}

export type SocialGroupMember = {
  user_id: string
  role: SocialGroupRole
  joined_at: string
  user?: SocialAuthor
}

export type SocialGroupJoinRequest = {
  id: string
  group_id: string
  user_id: string
  message: string
  status: string
  created_at: string
  user?: SocialAuthor
}

export type SocialGroupMessage = {
  id: string
  group_id: string
  sender_id: string
  body: string
  created_at: string
  sender?: SocialAuthor
}

export type ApiError = {
  error: string
  message: string
}
