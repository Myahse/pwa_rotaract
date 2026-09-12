package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rotaract-civ/backend/internal/clubname"
	"github.com/rotaract-civ/backend/internal/domain"
)

const clubSelectColumns = `
	id, name, slug, invite_code, description, country, city, commune, logo_path, founded_at,
	is_active, created_by, created_at, updated_at
`

const clubSelectColumnsAliased = `
	c.id, c.name, c.slug, c.invite_code, c.description, c.country, c.city, c.commune,
	c.logo_path, c.founded_at, c.is_active, c.created_by, c.created_at, c.updated_at
`

type ClubRepository struct {
	pool *pgxpool.Pool
}

func NewClubRepository(pool *pgxpool.Pool) *ClubRepository {
	return &ClubRepository{pool: pool}
}

func (r *ClubRepository) Create(ctx context.Context, club *domain.Club) error {
	query := `
		INSERT INTO clubs (name, slug, invite_code, description, country, city, commune, founded_at, logo_path, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, is_active, created_at, updated_at
	`
	err := r.pool.QueryRow(ctx, query,
		club.Name, club.Slug, club.InviteCode, club.Description,
		club.Country, club.City, club.Commune,
		club.FoundedAt, club.LogoPath, club.CreatedBy,
	).Scan(&club.ID, &club.IsActive, &club.CreatedAt, &club.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create club: %w", err)
	}
	return nil
}

func (r *ClubRepository) List(ctx context.Context) ([]domain.Club, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+clubSelectColumns+`
		FROM clubs ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list clubs: %w", err)
	}
	defer rows.Close()

	return r.scanClubs(rows)
}

func (r *ClubRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Club, error) {
	query := `SELECT ` + clubSelectColumns + ` FROM clubs WHERE id = $1`
	return r.scanClub(r.pool.QueryRow(ctx, query, id))
}

func (r *ClubRepository) GetByInviteCode(ctx context.Context, code string) (*domain.Club, error) {
	query := `SELECT ` + clubSelectColumns + ` FROM clubs WHERE invite_code = $1`
	return r.scanClub(r.pool.QueryRow(ctx, query, strings.ToUpper(strings.TrimSpace(code))))
}

func (r *ClubRepository) FindByName(ctx context.Context, name string) (*domain.Club, error) {
	trimmed := strings.TrimSpace(name)
	query := `
		SELECT ` + clubSelectColumns + `
		FROM clubs
		WHERE LOWER(name) = LOWER($1) AND is_active = TRUE
		LIMIT 1
	`
	club, err := r.scanClub(r.pool.QueryRow(ctx, query, trimmed))
	if err == nil {
		return club, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return nil, err
	}

	key := clubname.NormalizeKey(trimmed)
	if key == "" {
		return nil, ErrNotFound
	}

	rows, err := r.pool.Query(ctx, `
		SELECT `+clubSelectColumns+`
		FROM clubs
		WHERE is_active = TRUE
	`)
	if err != nil {
		return nil, fmt.Errorf("list clubs for name match: %w", err)
	}
	defer rows.Close()

	clubs, err := r.scanClubs(rows)
	if err != nil {
		return nil, err
	}
	for i := range clubs {
		if clubname.NormalizeKey(clubs[i].Name) == key {
			return &clubs[i], nil
		}
	}
	return nil, ErrNotFound
}

func (r *ClubRepository) scanClub(row pgx.Row) (*domain.Club, error) {
	var club domain.Club
	err := row.Scan(
		&club.ID, &club.Name, &club.Slug, &club.InviteCode, &club.Description,
		&club.Country, &club.City, &club.Commune,
		&club.LogoPath, &club.FoundedAt,
		&club.IsActive, &club.CreatedBy, &club.CreatedAt, &club.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan club: %w", err)
	}
	return &club, nil
}

func (r *ClubRepository) scanClubs(rows pgx.Rows) ([]domain.Club, error) {
	clubs := make([]domain.Club, 0)
	for rows.Next() {
		var club domain.Club
		if err := rows.Scan(
			&club.ID, &club.Name, &club.Slug, &club.InviteCode, &club.Description,
			&club.Country, &club.City, &club.Commune,
			&club.LogoPath, &club.FoundedAt,
			&club.IsActive, &club.CreatedBy, &club.CreatedAt, &club.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan club: %w", err)
		}
		clubs = append(clubs, club)
	}
	return clubs, rows.Err()
}

func (r *ClubRepository) Update(ctx context.Context, club *domain.Club) (*domain.Club, error) {
	err := r.pool.QueryRow(ctx, `
		UPDATE clubs
		SET name = $2, slug = $3, description = $4, country = $5, city = $6, commune = $7,
		    founded_at = $8, is_active = $9, updated_at = NOW()
		WHERE id = $1
		RETURNING `+clubSelectColumns+`
	`, club.ID, club.Name, club.Slug, club.Description, club.Country, club.City, club.Commune,
		club.FoundedAt, club.IsActive).Scan(
		&club.ID, &club.Name, &club.Slug, &club.InviteCode, &club.Description,
		&club.Country, &club.City, &club.Commune,
		&club.LogoPath, &club.FoundedAt,
		&club.IsActive, &club.CreatedBy, &club.CreatedAt, &club.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update club: %w", err)
	}
	return club, nil
}

func (r *ClubRepository) UpdateLogoPath(ctx context.Context, clubID uuid.UUID, path string) (*domain.Club, error) {
	var logoPath *string
	if path != "" {
		logoPath = &path
	}
	club, err := r.scanClub(r.pool.QueryRow(ctx, `
		UPDATE clubs SET logo_path = $2, updated_at = NOW() WHERE id = $1
		RETURNING `+clubSelectColumns+`
	`, clubID, logoPath))
	if err != nil {
		return nil, err
	}
	return club, nil
}

func (r *ClubRepository) UnassignRole(ctx context.Context, clubID, userID, roleID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `
		DELETE FROM club_member_role_assignments
		WHERE club_id = $1 AND user_id = $2 AND club_role_id = $3
	`, clubID, userID, roleID)
	if err != nil {
		return fmt.Errorf("unassign role: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *ClubRepository) AddMembership(ctx context.Context, membership *domain.ClubMembership) error {
	query := `
		INSERT INTO club_memberships (club_id, user_id, member_role)
		VALUES ($1, $2, $3)
		RETURNING id, joined_at
	`
	err := r.pool.QueryRow(ctx, query, membership.ClubID, membership.UserID, membership.MemberRole).
		Scan(&membership.ID, &membership.JoinedAt)
	if err != nil {
		return fmt.Errorf("add club membership: %w", err)
	}
	return nil
}

func (r *ClubRepository) GetMembership(ctx context.Context, clubID, userID uuid.UUID) (*domain.ClubMembership, error) {
	query := `
		SELECT id, club_id, user_id, member_role, joined_at
		FROM club_memberships WHERE club_id = $1 AND user_id = $2
	`
	var membership domain.ClubMembership
	err := r.pool.QueryRow(ctx, query, clubID, userID).Scan(
		&membership.ID, &membership.ClubID, &membership.UserID, &membership.MemberRole, &membership.JoinedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get membership: %w", err)
	}
	return &membership, nil
}

func (r *ClubRepository) IsHead(ctx context.Context, clubID, userID uuid.UUID) (bool, error) {
	membership, err := r.GetMembership(ctx, clubID, userID)
	if errors.Is(err, ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return membership.MemberRole == domain.ClubMemberRoleHead, nil
}

func (r *ClubRepository) ListMembers(ctx context.Context, clubID uuid.UUID) ([]domain.ClubMembership, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT cm.id, cm.club_id, cm.user_id, cm.member_role, cm.joined_at,
		       `+userSelectColumnsAliased+`
		FROM club_memberships cm
		JOIN users u ON u.id = cm.user_id
		WHERE cm.club_id = $1
		ORDER BY cm.joined_at ASC
	`, clubID)
	if err != nil {
		return nil, fmt.Errorf("list members: %w", err)
	}
	defer rows.Close()

	members := make([]domain.ClubMembership, 0)
	for rows.Next() {
		var membership domain.ClubMembership
		var user domain.User
		if err := rows.Scan(
			&membership.ID, &membership.ClubID, &membership.UserID, &membership.MemberRole, &membership.JoinedAt,
			&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName, &user.Phone,
			&user.BirthDate, &user.Profession, &user.MemberSince, &user.AvatarPath,
			&user.IsAdmin, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan member: %w", err)
		}
		membership.User = &user
		members = append(members, membership)
	}
	return members, rows.Err()
}

func (r *ClubRepository) ListMembershipsByUser(ctx context.Context, userID uuid.UUID) ([]domain.ClubMembership, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT cm.id, cm.club_id, cm.user_id, cm.member_role, cm.joined_at,
		       `+clubSelectColumnsAliased+`
		FROM club_memberships cm
		JOIN clubs c ON c.id = cm.club_id
		WHERE cm.user_id = $1
		ORDER BY cm.joined_at ASC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list user memberships: %w", err)
	}
	defer rows.Close()

	memberships := make([]domain.ClubMembership, 0)
	for rows.Next() {
		var membership domain.ClubMembership
		var club domain.Club
		if err := rows.Scan(
			&membership.ID, &membership.ClubID, &membership.UserID, &membership.MemberRole, &membership.JoinedAt,
			&club.ID, &club.Name, &club.Slug, &club.InviteCode, &club.Description, &club.City,
			&club.IsActive, &club.CreatedBy, &club.CreatedAt, &club.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan user membership: %w", err)
		}
		membership.Club = &club
		memberships = append(memberships, membership)
	}
	return memberships, rows.Err()
}

func (r *ClubRepository) CreateRole(ctx context.Context, role *domain.ClubRole, permissionKeys []string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO club_roles (club_id, name, description, is_system)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`, role.ClubID, role.Name, role.Description, role.IsSystem).Scan(&role.ID, &role.CreatedAt)
	if err != nil {
		return fmt.Errorf("create club role: %w", err)
	}

	for _, key := range permissionKeys {
		_, err = tx.Exec(ctx, `
			INSERT INTO club_role_permissions (club_role_id, permission_id)
			SELECT $1, id FROM permissions WHERE key = $2
		`, role.ID, key)
		if err != nil {
			return fmt.Errorf("assign permission %s: %w", key, err)
		}
	}

	return tx.Commit(ctx)
}

func (r *ClubRepository) GetRoleByID(ctx context.Context, clubID, roleID uuid.UUID) (*domain.ClubRole, error) {
	var role domain.ClubRole
	err := r.pool.QueryRow(ctx, `
		SELECT id, club_id, name, description, is_system, created_at
		FROM club_roles WHERE id = $1 AND club_id = $2
	`, roleID, clubID).Scan(&role.ID, &role.ClubID, &role.Name, &role.Description, &role.IsSystem, &role.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get role: %w", err)
	}
	return &role, nil
}

func (r *ClubRepository) ListRoles(ctx context.Context, clubID uuid.UUID) ([]domain.ClubRole, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, club_id, name, description, is_system, created_at
		FROM club_roles WHERE club_id = $1 ORDER BY created_at ASC
	`, clubID)
	if err != nil {
		return nil, fmt.Errorf("list roles: %w", err)
	}
	defer rows.Close()

	roles := make([]domain.ClubRole, 0)
	for rows.Next() {
		var role domain.ClubRole
		if err := rows.Scan(&role.ID, &role.ClubID, &role.Name, &role.Description, &role.IsSystem, &role.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan role: %w", err)
		}
		roles = append(roles, role)
	}
	return roles, rows.Err()
}

func (r *ClubRepository) ListJoinableRoles(ctx context.Context, clubID uuid.UUID) ([]domain.ClubRole, error) {
	roles, err := r.ListRoles(ctx, clubID)
	if err != nil {
		return nil, err
	}

	joinable := make([]domain.ClubRole, 0, len(roles))
	for _, role := range roles {
		if role.Name == "Responsable de club" {
			continue
		}
		joinable = append(joinable, role)
	}
	return joinable, nil
}

func (r *ClubRepository) AssignRole(ctx context.Context, assignment *domain.ClubMemberRoleAssignment) error {
	query := `
		INSERT INTO club_member_role_assignments (club_id, user_id, club_role_id, assigned_by)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (club_id, user_id, club_role_id) DO NOTHING
		RETURNING id, assigned_at
	`
	err := r.pool.QueryRow(ctx, query, assignment.ClubID, assignment.UserID, assignment.ClubRoleID, assignment.AssignedBy).
		Scan(&assignment.ID, &assignment.AssignedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("assign role: %w", err)
	}
	return nil
}

func (r *ClubRepository) UserHasPermission(ctx context.Context, clubID, userID uuid.UUID, permissionKey string) (bool, error) {
	if head, err := r.IsHead(ctx, clubID, userID); err != nil {
		return false, err
	} else if head {
		return true, nil
	}

	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM club_member_role_assignments cmra
			JOIN club_role_permissions crp ON crp.club_role_id = cmra.club_role_id
			JOIN permissions p ON p.id = crp.permission_id
			WHERE cmra.club_id = $1 AND cmra.user_id = $2 AND p.key = $3
		)
	`, clubID, userID, permissionKey).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check permission: %w", err)
	}

	if !exists && permissionKey == "commission.groups.create" {
		return r.userIsAnyCommissionPresident(ctx, clubID, userID)
	}

	return exists, nil
}

func (r *ClubRepository) userIsAnyCommissionPresident(ctx context.Context, clubID, userID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM commission_memberships cm
			JOIN commissions c ON c.id = cm.commission_id
			WHERE c.club_id = $1 AND cm.user_id = $2 AND cm.member_role = 'president'
		)
	`, clubID, userID).Scan(&exists)
	return exists, err
}

func (r *ClubRepository) SeedHeadRole(ctx context.Context, clubID uuid.UUID) (*domain.ClubRole, error) {
	role := &domain.ClubRole{
		ClubID:      clubID,
		Name:        "Responsable de club",
		Description: strPtr("Rôle système pour le responsable du club"),
		IsSystem:    true,
	}
	keys := []string{
		"club.members.read",
		"club.members.assign_role",
		"club.roles.manage",
		"club.commissions.create",
		"club.commissions.read",
		"club.commissions.manage_members",
		"club.requests.review",
		"club.invites.send",
		"club.diary.read",
		"club.diary.manage",
		"chat.groups.read",
	}
	if err := r.CreateRole(ctx, role, keys); err != nil {
		return nil, err
	}
	return role, nil
}

func (r *ClubRepository) SeedMemberRole(ctx context.Context, clubID uuid.UUID) (*domain.ClubRole, error) {
	role := &domain.ClubRole{
		ClubID:      clubID,
		Name:        "Membre",
		Description: strPtr("Rôle membre par défaut"),
		IsSystem:    true,
	}
	if err := r.CreateRole(ctx, role, []string{"club.members.read", "club.diary.read", "chat.groups.read"}); err != nil {
		return nil, err
	}
	return role, nil
}

func strPtr(v string) *string {
	return &v
}
