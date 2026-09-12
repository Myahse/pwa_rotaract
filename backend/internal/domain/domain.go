package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID  `json:"id"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	GoogleSub    *string    `json:"-"`
	FirstName    string     `json:"first_name"`
	LastName     string     `json:"last_name"`
	Phone        *string    `json:"phone,omitempty"`
	BirthDate    *time.Time `json:"birth_date,omitempty"`
	Profession   *string    `json:"profession,omitempty"`
	MemberSince  *time.Time `json:"member_since,omitempty"`
	AvatarPath   *string    `json:"-"`
	AvatarURL    *string    `json:"avatar_url,omitempty"`
	IsAdmin      bool       `json:"is_admin"`
	IsActive     bool       `json:"is_active"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func (u User) FullName() string {
	return u.FirstName + " " + u.LastName
}

type Club struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Slug        string     `json:"slug"`
	InviteCode  string     `json:"invite_code"`
	Description *string    `json:"description,omitempty"`
	Country     *string    `json:"country,omitempty"`
	City        *string    `json:"city,omitempty"`
	Commune     *string    `json:"commune,omitempty"`
	LogoPath    *string    `json:"-"`
	LogoURL     *string    `json:"logo_url,omitempty"`
	FoundedAt   *time.Time `json:"founded_at,omitempty"`
	IsActive    bool       `json:"is_active"`
	CreatedBy   *uuid.UUID `json:"created_by,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type ClubMemberRole string

const (
	ClubMemberRoleHead   ClubMemberRole = "head"
	ClubMemberRoleMember ClubMemberRole = "member"
)

type ClubMembership struct {
	ID         uuid.UUID      `json:"id"`
	ClubID     uuid.UUID      `json:"club_id"`
	UserID     uuid.UUID      `json:"user_id"`
	MemberRole ClubMemberRole `json:"member_role"`
	JoinedAt   time.Time      `json:"joined_at"`
	User       *User          `json:"user,omitempty"`
	Club       *Club          `json:"club,omitempty"`
}

type Permission struct {
	ID          uuid.UUID `json:"id"`
	Key         string    `json:"key"`
	Description string    `json:"description"`
}

type ClubRole struct {
	ID          uuid.UUID    `json:"id"`
	ClubID      uuid.UUID    `json:"club_id"`
	Name        string       `json:"name"`
	Description *string      `json:"description,omitempty"`
	IsSystem    bool         `json:"is_system"`
	CreatedAt   time.Time    `json:"created_at"`
	Permissions []Permission `json:"permissions,omitempty"`
}

type ClubMemberRoleAssignment struct {
	ID         uuid.UUID `json:"id"`
	ClubID     uuid.UUID `json:"club_id"`
	UserID     uuid.UUID `json:"user_id"`
	ClubRoleID uuid.UUID `json:"club_role_id"`
	AssignedBy *uuid.UUID `json:"assigned_by,omitempty"`
	AssignedAt time.Time `json:"assigned_at"`
	Role       *ClubRole `json:"role,omitempty"`
}

type CommissionMemberRole string

const (
	CommissionMemberRolePresident CommissionMemberRole = "president"
	CommissionMemberRoleSecretary CommissionMemberRole = "secretary"
	CommissionMemberRoleMember    CommissionMemberRole = "member"
)

type Commission struct {
	ID          uuid.UUID  `json:"id"`
	ClubID      uuid.UUID  `json:"club_id"`
	Code        *string    `json:"code,omitempty"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	IsSystem    bool       `json:"is_system"`
	CreatedBy   *uuid.UUID `json:"created_by,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

type CommissionMembership struct {
	ID           uuid.UUID            `json:"id"`
	CommissionID uuid.UUID            `json:"commission_id"`
	UserID       uuid.UUID            `json:"user_id"`
	MemberRole   CommissionMemberRole `json:"member_role"`
	JoinedAt     time.Time            `json:"joined_at"`
	User         *User                `json:"user,omitempty"`
}

type ChatGroupType string

const (
	ChatGroupTypeClub       ChatGroupType = "club"
	ChatGroupTypeCommission ChatGroupType = "commission"
	ChatGroupTypeCustom     ChatGroupType = "custom"
)

type ChatGroup struct {
	ID           uuid.UUID      `json:"id"`
	ClubID       uuid.UUID      `json:"club_id"`
	CommissionID *uuid.UUID     `json:"commission_id,omitempty"`
	GroupType    ChatGroupType  `json:"group_type"`
	Name         string         `json:"name"`
	CreatedBy    *uuid.UUID     `json:"created_by,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
}

type ChatGroupMember struct {
	ID       uuid.UUID `json:"id"`
	GroupID  uuid.UUID `json:"group_id"`
	UserID   uuid.UUID `json:"user_id"`
	JoinedAt time.Time `json:"joined_at"`
	User     *User     `json:"user,omitempty"`
}

type ChatMessage struct {
	ID        uuid.UUID `json:"id"`
	GroupID   uuid.UUID `json:"group_id"`
	UserID    uuid.UUID `json:"user_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	User      *User     `json:"user,omitempty"`
}

type AccessRequestStatus string

const (
	AccessRequestStatusPending  AccessRequestStatus = "pending"
	AccessRequestStatusApproved AccessRequestStatus = "approved"
	AccessRequestStatusRejected AccessRequestStatus = "rejected"
)

type AccessRequest struct {
	ID              uuid.UUID           `json:"id"`
	ClubID          *uuid.UUID          `json:"club_id,omitempty"`
	ClubName        string              `json:"club_name"`
	RequestedRoleID *uuid.UUID          `json:"requested_role_id,omitempty"`
	Email           string              `json:"email"`
	FirstName       string              `json:"first_name"`
	LastName        string              `json:"last_name"`
	Phone           *string             `json:"phone,omitempty"`
	BirthDate       *time.Time          `json:"birth_date,omitempty"`
	Profession      *string             `json:"profession,omitempty"`
	MemberSince     *time.Time          `json:"member_since,omitempty"`
	ExistingUserID  *uuid.UUID          `json:"existing_user_id,omitempty"`
	KnownMember     bool                `json:"known_member"`
	Status          AccessRequestStatus `json:"status"`
	ReviewedBy      *uuid.UUID          `json:"reviewed_by,omitempty"`
	ReviewNote      *string             `json:"review_note,omitempty"`
	ReviewedAt      *time.Time          `json:"reviewed_at,omitempty"`
	CreatedAt       time.Time           `json:"created_at"`
	RequestedRole   *ClubRole           `json:"requested_role,omitempty"`
}

type ClubRegistrationStatus string

const (
	ClubRegistrationStatusPending  ClubRegistrationStatus = "pending"
	ClubRegistrationStatusApproved ClubRegistrationStatus = "approved"
	ClubRegistrationStatusRejected ClubRegistrationStatus = "rejected"
)

// ClubRegistrationRequest is a public application to create a new club.
// On approval, an access link is emailed; the contact finishes with Google Sign-In.
type ClubRegistrationRequest struct {
	ID                   uuid.UUID              `json:"id"`
	ClubName             string                 `json:"club_name"`
	ContactEmail         string                 `json:"contact_email"`
	ContactFirstName     string                 `json:"contact_first_name"`
	ContactLastName      string                 `json:"contact_last_name"`
	Phone                *string                `json:"phone,omitempty"`
	Description          *string                `json:"description,omitempty"`
	Country              string                 `json:"country"`
	City                 string                 `json:"city"`
	Commune              string                 `json:"commune"`
	FoundedAt            *time.Time             `json:"founded_at,omitempty"`
	Message              *string                `json:"message,omitempty"`
	Status               ClubRegistrationStatus `json:"status"`
	ReviewedBy           *uuid.UUID             `json:"reviewed_by,omitempty"`
	ReviewNote           *string                `json:"review_note,omitempty"`
	ReviewedAt           *time.Time             `json:"reviewed_at,omitempty"`
	CreatedClubID        *uuid.UUID             `json:"created_club_id,omitempty"`
	AccessToken          string                 `json:"-"`
	AccessTokenExpiresAt *time.Time             `json:"access_token_expires_at,omitempty"`
	AccessTokenUsedAt    *time.Time             `json:"access_token_used_at,omitempty"`
	ApprovedSlug         *string                `json:"approved_slug,omitempty"`
	CreatedAt            time.Time              `json:"created_at"`
}

type ClubRegistrationAccessPreview struct {
	ClubName         string    `json:"club_name"`
	ContactEmail     string    `json:"contact_email"`
	ContactFirstName string    `json:"contact_first_name"`
	ContactLastName  string    `json:"contact_last_name"`
	ExpiresAt        time.Time `json:"expires_at"`
}

type PushSubscription struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Endpoint  string    `json:"endpoint"`
	P256dh    string    `json:"p256dh"`
	Auth      string    `json:"auth"`
	UserAgent *string   `json:"user_agent,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type BirthdayWidgetMember struct {
	UserID    uuid.UUID `json:"user_id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	AvatarURL *string   `json:"avatar_url,omitempty"`
	Message   string    `json:"message"`
}

type BirthdayWidget struct {
	Date          string                 `json:"date"`
	IsMyBirthday  bool                   `json:"is_my_birthday"`
	Title         string                 `json:"title"`
	Message       string                 `json:"message"`
	Emoji         string                 `json:"emoji"`
	Theme         BirthdayWidgetTheme    `json:"theme"`
	ClubBirthdays []BirthdayWidgetMember `json:"club_birthdays,omitempty"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

type BirthdayWidgetTheme struct {
	Primary string `json:"primary"`
	Accent  string `json:"accent"`
	Gradient string `json:"gradient"`
}

type EmailInvite struct {
	ID        uuid.UUID  `json:"id"`
	ClubID    uuid.UUID  `json:"club_id"`
	Email     string     `json:"email"`
	RoleID    *uuid.UUID `json:"role_id,omitempty"`
	Token     string     `json:"-"`
	InvitedBy uuid.UUID  `json:"invited_by"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

type EmailInvitePreview struct {
	Club      Club       `json:"club"`
	Email     string     `json:"email"`
	Role      *ClubRole  `json:"role,omitempty"`
	Roles     []ClubRole `json:"roles"`
	ExpiresAt time.Time  `json:"expires_at"`
}

type ClubDiaryEntryType string

const (
	ClubDiaryEntryParrain   ClubDiaryEntryType = "parrain"
	ClubDiaryEntryPresident ClubDiaryEntryType = "president"
	ClubDiaryEntryMember    ClubDiaryEntryType = "member"
)

type ClubDiaryEntry struct {
	ID           uuid.UUID          `json:"id"`
	ClubID       uuid.UUID          `json:"club_id"`
	ParentID     *uuid.UUID         `json:"parent_id,omitempty"`
	EntryType    ClubDiaryEntryType `json:"entry_type"`
	UserID       *uuid.UUID         `json:"user_id,omitempty"`
	FirstName    string             `json:"first_name"`
	LastName     string             `json:"last_name"`
	Organization *string            `json:"organization,omitempty"`
	PhotoPath    *string            `json:"-"`
	PhotoURL     *string            `json:"photo_url,omitempty"`
	StartedAt    *time.Time         `json:"started_at,omitempty"`
	EndedAt      *time.Time         `json:"ended_at,omitempty"`
	Notes        *string            `json:"notes,omitempty"`
	SortOrder    int                `json:"sort_order"`
	CreatedBy    *uuid.UUID         `json:"created_by,omitempty"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
	User         *User              `json:"user,omitempty"`
}

type ClubDiaryTreeNode struct {
	ClubDiaryEntry
	Children []ClubDiaryTreeNode `json:"children"`
}

type ClubDiaryResponse struct {
	Entries  []ClubDiaryEntry    `json:"entries"`
	Tree     []ClubDiaryTreeNode `json:"tree"`
	Parrains []ClubDiaryEntry    `json:"parrains"`
	Mandates []ClubMandate       `json:"mandates"`
}

type ClubMandateRole string

const (
	ClubMandateRolePresident             ClubMandateRole = "president"
	ClubMandateRolePresidentElect        ClubMandateRole = "president_elect"
	ClubMandateRoleImmediatePastPresident ClubMandateRole = "immediate_past_president"
	ClubMandateRoleVicePresident         ClubMandateRole = "vice_president"
	ClubMandateRoleSecretary             ClubMandateRole = "secretary"
	ClubMandateRoleAssistantSecretary    ClubMandateRole = "assistant_secretary"
	ClubMandateRoleTreasurer             ClubMandateRole = "treasurer"
	ClubMandateRoleAssistantTreasurer    ClubMandateRole = "assistant_treasurer"
	ClubMandateRoleProtocol               ClubMandateRole = "protocol"
	ClubMandateRoleAssistantProtocol     ClubMandateRole = "assistant_protocol"
	ClubMandateRoleCommissionPresident   ClubMandateRole = "commission_president"
	ClubMandateRoleCommissionSecretary   ClubMandateRole = "commission_secretary"
	ClubMandateRoleCommissionMember      ClubMandateRole = "commission_member"
	ClubMandateRoleMember                ClubMandateRole = "member"
)

func (r ClubMandateRole) IsExecutiveBureau() bool {
	switch r {
	case ClubMandateRolePresident, ClubMandateRolePresidentElect, ClubMandateRoleImmediatePastPresident,
		ClubMandateRoleVicePresident, ClubMandateRoleSecretary, ClubMandateRoleAssistantSecretary,
		ClubMandateRoleTreasurer, ClubMandateRoleAssistantTreasurer,
		ClubMandateRoleProtocol, ClubMandateRoleAssistantProtocol:
		return true
	default:
		return false
	}
}

func (r ClubMandateRole) IsPresidentTier() bool {
	switch r {
	case ClubMandateRolePresident, ClubMandateRolePresidentElect,
		ClubMandateRoleImmediatePastPresident, ClubMandateRoleVicePresident:
		return true
	default:
		return false
	}
}

type ClubMandate struct {
	ID        uuid.UUID                `json:"id"`
	ClubID    uuid.UUID                `json:"club_id"`
	Name      string                   `json:"name"`
	StartedAt time.Time                `json:"started_at"`
	EndedAt   *time.Time               `json:"ended_at,omitempty"`
	IsCurrent bool                     `json:"is_current"`
	Notes     *string                  `json:"notes,omitempty"`
	SortOrder int                      `json:"sort_order"`
	CreatedBy *uuid.UUID               `json:"created_by,omitempty"`
	CreatedAt time.Time                `json:"created_at"`
	UpdatedAt time.Time                `json:"updated_at"`
	Bureau    []ClubMandateAssignment  `json:"bureau"`
	Commissions []ClubMandateCommissionGroup `json:"commissions"`
	Members   []ClubMandateAssignment  `json:"members"`
}

type ClubMandateAssignment struct {
	ID             uuid.UUID       `json:"id"`
	MandateID      uuid.UUID       `json:"mandate_id"`
	Role           ClubMandateRole `json:"role"`
	CommissionID   *uuid.UUID      `json:"commission_id,omitempty"`
	CommissionName *string         `json:"commission_name,omitempty"`
	UserID         *uuid.UUID      `json:"user_id,omitempty"`
	FirstName      string          `json:"first_name"`
	LastName       string          `json:"last_name"`
	PhotoPath      *string         `json:"-"`
	PhotoURL       *string         `json:"photo_url,omitempty"`
	Notes          *string         `json:"notes,omitempty"`
	SortOrder      int             `json:"sort_order"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	User           *User           `json:"user,omitempty"`
}

type ClubMandateCommissionGroup struct {
	CommissionID   *uuid.UUID              `json:"commission_id,omitempty"`
	CommissionName string                  `json:"commission_name"`
	President      *ClubMandateAssignment  `json:"president,omitempty"`
	Secretary      *ClubMandateAssignment  `json:"secretary,omitempty"`
	Members        []ClubMandateAssignment `json:"members"`
}

// Social feed (community platform — distinct from club ops chat)

type SocialAuthor struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	AvatarURL *string   `json:"avatar_url,omitempty"`
}

type SocialClubTag struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type SocialMediaKind string

const (
	SocialMediaImage SocialMediaKind = "image"
	SocialMediaVideo SocialMediaKind = "video"
)

type SocialPostMedia struct {
	ID        uuid.UUID       `json:"id"`
	PostID    uuid.UUID       `json:"post_id"`
	Kind      SocialMediaKind `json:"kind"`
	Path      string          `json:"-"`
	URL       string          `json:"url"`
	SortOrder int             `json:"sort_order"`
	CreatedAt time.Time       `json:"created_at"`
}

type SocialPost struct {
	ID            uuid.UUID         `json:"id"`
	AuthorID      uuid.UUID         `json:"author_id"`
	ClubID        *uuid.UUID        `json:"club_id,omitempty"`
	RepostedPostID *uuid.UUID       `json:"reposted_post_id,omitempty"`
	QuoteBody     string            `json:"quote_body,omitempty"`
	Body          string            `json:"body"`
	IsHidden      bool              `json:"is_hidden,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	Author        *SocialAuthor     `json:"author,omitempty"`
	Club          *SocialClubTag    `json:"club,omitempty"`
	Media         []SocialPostMedia `json:"media,omitempty"`
	CommentCount  int               `json:"comment_count"`
	ReactionCount int               `json:"reaction_count"`
	ReactedByMe   bool              `json:"reacted_by_me"`
	AuthorFollowedByMe bool         `json:"author_followed_by_me"`
	Original      *SocialPost      `json:"original,omitempty"`
}

type SocialComment struct {
	ID        uuid.UUID     `json:"id"`
	PostID    uuid.UUID     `json:"post_id"`
	AuthorID  uuid.UUID     `json:"author_id"`
	ParentID  *uuid.UUID    `json:"parent_id,omitempty"`
	Body      string        `json:"body"`
	CreatedAt time.Time     `json:"created_at"`
	Author    *SocialAuthor `json:"author,omitempty"`
	Replies   []SocialComment `json:"replies,omitempty"`
}

type SocialFollowProfile struct {
	ID               uuid.UUID        `json:"id"`
	FirstName        string           `json:"first_name"`
	LastName         string           `json:"last_name"`
	AvatarURL        *string          `json:"avatar_url,omitempty"`
	FollowedByMe     bool             `json:"followed_by_me"`
	FollowersCount   int              `json:"followers_count"`
	FriendshipStatus FriendshipStatus `json:"friendship_status,omitempty"`
}

type FriendshipStatus string

const (
	FriendshipNone     FriendshipStatus = "none"
	FriendshipPending  FriendshipStatus = "pending"
	FriendshipAccepted FriendshipStatus = "accepted"
	FriendshipDeclined FriendshipStatus = "declined"
)

type SocialFriendProfile struct {
	ID               uuid.UUID        `json:"id"`
	FirstName        string           `json:"first_name"`
	LastName         string           `json:"last_name"`
	AvatarURL        *string          `json:"avatar_url,omitempty"`
	FriendshipStatus FriendshipStatus `json:"friendship_status"`
	IncomingRequest  bool             `json:"incoming_request"`
	FriendsSince     *time.Time       `json:"friends_since,omitempty"`
}

type SocialUserProfile struct {
	ID               uuid.UUID        `json:"id"`
	FirstName        string           `json:"first_name"`
	LastName         string           `json:"last_name"`
	AvatarURL        *string          `json:"avatar_url,omitempty"`
	FollowersCount   int              `json:"followers_count"`
	FollowingCount   int              `json:"following_count"`
	FriendsCount     int              `json:"friends_count"`
	FollowedByMe     bool             `json:"followed_by_me"`
	FriendshipStatus FriendshipStatus `json:"friendship_status"`
	IncomingRequest  bool             `json:"incoming_request"`
	IsMe             bool             `json:"is_me"`
	Clubs            []SocialClubTag  `json:"clubs,omitempty"`
}

type SocialConversation struct {
	ID            uuid.UUID           `json:"id"`
	UpdatedAt     time.Time           `json:"updated_at"`
	Peer          *SocialAuthor       `json:"peer,omitempty"`
	LastMessage   *SocialMessage      `json:"last_message,omitempty"`
	UnreadCount   int                 `json:"unread_count"`
}

type SocialMessage struct {
	ID              uuid.UUID      `json:"id"`
	ConversationID  uuid.UUID      `json:"conversation_id"`
	SenderID        uuid.UUID      `json:"sender_id"`
	Body            string         `json:"body"`
	SharedCommentID *uuid.UUID     `json:"shared_comment_id,omitempty"`
	SharedPostID    *uuid.UUID     `json:"shared_post_id,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	Sender          *SocialAuthor  `json:"sender,omitempty"`
	SharedComment   *SocialComment `json:"shared_comment,omitempty"`
	SharedPost      *SocialPost    `json:"shared_post,omitempty"`
}

type SocialGroupPrivacy string

const (
	SocialGroupOpen     SocialGroupPrivacy = "open"
	SocialGroupApproval SocialGroupPrivacy = "approval"
)

type SocialGroupMemberRole string

const (
	SocialGroupRoleAdmin  SocialGroupMemberRole = "admin"
	SocialGroupRoleMember SocialGroupMemberRole = "member"
)

type SocialGroupJoinStatus string

const (
	SocialGroupJoinNone     SocialGroupJoinStatus = "none"
	SocialGroupJoinPending  SocialGroupJoinStatus = "pending"
	SocialGroupJoinMember   SocialGroupJoinStatus = "member"
	SocialGroupJoinRejected SocialGroupJoinStatus = "rejected"
)

type SocialGroup struct {
	ID           uuid.UUID          `json:"id"`
	Name         string             `json:"name"`
	Description  string             `json:"description"`
	Privacy      SocialGroupPrivacy `json:"privacy"`
	CreatedBy    uuid.UUID          `json:"created_by"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
	MemberCount  int                `json:"member_count"`
	MyRole       *SocialGroupMemberRole `json:"my_role,omitempty"`
	JoinStatus   SocialGroupJoinStatus  `json:"join_status"`
	PendingCount int                `json:"pending_count,omitempty"`
	Creator      *SocialAuthor      `json:"creator,omitempty"`
}

type SocialGroupMember struct {
	UserID   uuid.UUID             `json:"user_id"`
	Role     SocialGroupMemberRole `json:"role"`
	JoinedAt time.Time             `json:"joined_at"`
	User     *SocialAuthor         `json:"user,omitempty"`
}

type SocialGroupJoinRequest struct {
	ID        uuid.UUID  `json:"id"`
	GroupID   uuid.UUID  `json:"group_id"`
	UserID    uuid.UUID  `json:"user_id"`
	Message   string     `json:"message"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	User      *SocialAuthor `json:"user,omitempty"`
}

type SocialGroupMessage struct {
	ID        uuid.UUID     `json:"id"`
	GroupID   uuid.UUID     `json:"group_id"`
	SenderID  uuid.UUID     `json:"sender_id"`
	Body      string        `json:"body"`
	CreatedAt time.Time     `json:"created_at"`
	Sender    *SocialAuthor `json:"sender,omitempty"`
}

type DonationStatus string

const (
	DonationStatusPending  DonationStatus = "pending"
	DonationStatusReceived DonationStatus = "received"
)

type ClubDuePayment struct {
	ID          uuid.UUID      `json:"id"`
	ClubID      uuid.UUID      `json:"club_id"`
	UserID      uuid.UUID      `json:"user_id"`
	DueMonth    string         `json:"due_month"`
	AmountXOF   int            `json:"amount_xof"`
	Status      DonationStatus `json:"status"`
	ReceiptPath string         `json:"-"`
	ReceiptMime string         `json:"receipt_mime,omitempty"`
	ReceiptURL  *string        `json:"receipt_url,omitempty"`
	ReviewedBy  *uuid.UUID     `json:"reviewed_by,omitempty"`
	ReviewedAt  *time.Time     `json:"reviewed_at,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	User        *User          `json:"user,omitempty"`
}

type UserCommissionAccess struct {
	CommissionID   uuid.UUID            `json:"commission_id"`
	CommissionName string               `json:"commission_name"`
	MemberRole     CommissionMemberRole `json:"member_role"`
}

type ClubAccessSummary struct {
	MemberRole  ClubMemberRole         `json:"member_role"`
	Roles       []ClubRole             `json:"roles"`
	Permissions []string               `json:"permissions"`
	Commissions []UserCommissionAccess `json:"commissions"`
}

type Donation struct {
	ID          uuid.UUID      `json:"id"`
	UserID      *uuid.UUID     `json:"user_id,omitempty"`
	Name        string         `json:"name"`
	Email       string         `json:"email"`
	AmountXOF   int            `json:"amount_xof"`
	Status      DonationStatus `json:"status"`
	KnownMember bool           `json:"known_member"`
	ReceiptPath string         `json:"-"`
	ReceiptMime string         `json:"receipt_mime,omitempty"`
	ReceiptURL  *string        `json:"receipt_url,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
}

type PublicEventImage struct {
	ID  uuid.UUID `json:"id"`
	URL string    `json:"url"`
}

type PublicEvent struct {
	ID          uuid.UUID          `json:"id"`
	Published   bool               `json:"published"`
	StartsAt    time.Time          `json:"starts_at"`
	City        string             `json:"city"`
	VenueFr     string             `json:"venue_fr"`
	VenueEn     string             `json:"venue_en"`
	TitleFr     string             `json:"title_fr"`
	TitleEn     string             `json:"title_en"`
	SummaryFr   string             `json:"summary_fr"`
	SummaryEn   string             `json:"summary_en"`
	BodyFr      string             `json:"body_fr"`
	BodyEn      string             `json:"body_en"`
	CtaLabelFr  string             `json:"cta_label_fr"`
	CtaLabelEn  string             `json:"cta_label_en"`
	CtaURL      string             `json:"cta_url"`
	FlyerPath   *string            `json:"-"`
	FlyerURL    *string            `json:"flyer_url,omitempty"`
	Images      []PublicEventImage `json:"images"`
	Status      string             `json:"status"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

type SiteGalleryImage struct {
	ID        uuid.UUID `json:"id"`
	Published bool      `json:"published"`
	CaptionFr string    `json:"caption_fr"`
	CaptionEn string    `json:"caption_en"`
	Path      string    `json:"-"`
	URL       string    `json:"url,omitempty"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type FeaturedPostulant struct {
	ID           uuid.UUID `json:"id"`
	PeriodYear   int       `json:"period_year"`
	PeriodMonth  int       `json:"period_month"`
	Published    bool      `json:"published"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	HomeClub     string    `json:"home_club"`
	QuoteFr      string    `json:"quote_fr"`
	QuoteEn      string    `json:"quote_en"`
	VisitCount   int       `json:"visit_count"`
	ClubsVisited []string  `json:"clubs_visited"`
	FlyerPath    *string   `json:"-"`
	FlyerURL     *string   `json:"flyer_url,omitempty"`
	UpdatedAt    time.Time `json:"updated_at"`
}
