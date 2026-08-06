package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rotaract-civ/backend/internal/domain"
)

type UpdateProfileParams struct {
	FirstName   *string
	LastName    *string
	Phone       *string
	BirthDate   *time.Time
	Profession  *string
	MemberSince *time.Time
}

const userSelectColumns = `
	id, email, password_hash, first_name, last_name, phone,
	birth_date, profession, member_since, avatar_path,
	is_admin, is_active, created_at, updated_at
`

const userSelectColumnsAliased = `
	u.id, u.email, u.password_hash, u.first_name, u.last_name, u.phone,
	u.birth_date, u.profession, u.member_since, u.avatar_path,
	u.is_admin, u.is_active, u.created_at, u.updated_at
`

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (
			email, password_hash, google_sub, first_name, last_name, phone,
			birth_date, profession, member_since, is_admin, is_active
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at
	`
	err := r.pool.QueryRow(ctx, query,
		user.Email, user.PasswordHash, user.GoogleSub, user.FirstName, user.LastName, user.Phone,
		user.BirthDate, user.Profession, user.MemberSince, user.IsAdmin, user.IsActive,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *UserRepository) GetByGoogleSub(ctx context.Context, googleSub string) (*domain.User, error) {
	query := `SELECT ` + userSelectColumns + ` FROM users WHERE google_sub = $1`
	return r.scanUser(r.pool.QueryRow(ctx, query, strings.TrimSpace(googleSub)))
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT ` + userSelectColumns + ` FROM users WHERE email = $1`
	return r.scanUser(r.pool.QueryRow(ctx, query, email))
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `SELECT ` + userSelectColumns + ` FROM users WHERE id = $1`
	return r.scanUser(r.pool.QueryRow(ctx, query, id))
}

func (r *UserRepository) UpdateProfile(ctx context.Context, userID uuid.UUID, input UpdateProfileParams) (*domain.User, error) {
	setParts := make([]string, 0, 6)
	args := make([]any, 0, 7)
	argIndex := 1

	if input.FirstName != nil {
		setParts = append(setParts, fmt.Sprintf("first_name = $%d", argIndex))
		args = append(args, *input.FirstName)
		argIndex++
	}
	if input.LastName != nil {
		setParts = append(setParts, fmt.Sprintf("last_name = $%d", argIndex))
		args = append(args, *input.LastName)
		argIndex++
	}
	if input.Phone != nil {
		setParts = append(setParts, fmt.Sprintf("phone = $%d", argIndex))
		args = append(args, *input.Phone)
		argIndex++
	}
	if input.BirthDate != nil {
		setParts = append(setParts, fmt.Sprintf("birth_date = $%d", argIndex))
		args = append(args, *input.BirthDate)
		argIndex++
	}
	if input.Profession != nil {
		setParts = append(setParts, fmt.Sprintf("profession = $%d", argIndex))
		args = append(args, *input.Profession)
		argIndex++
	}
	if input.MemberSince != nil {
		setParts = append(setParts, fmt.Sprintf("member_since = $%d", argIndex))
		args = append(args, *input.MemberSince)
		argIndex++
	}

	if len(setParts) == 0 {
		return r.GetByID(ctx, userID)
	}

	setParts = append(setParts, "updated_at = NOW()")
	args = append(args, userID)

	query := fmt.Sprintf(`
		UPDATE users SET %s
		WHERE id = $%d
		RETURNING %s
	`, strings.Join(setParts, ", "), argIndex, userSelectColumns)

	return r.scanUser(r.pool.QueryRow(ctx, query, args...))
}

func (r *UserRepository) UpdateAvatarPath(ctx context.Context, userID uuid.UUID, path string) (*domain.User, error) {
	var avatarPath any
	if strings.TrimSpace(path) == "" {
		avatarPath = nil
	} else {
		avatarPath = path
	}

	query := `
		UPDATE users SET avatar_path = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING ` + userSelectColumns

	return r.scanUser(r.pool.QueryRow(ctx, query, userID, avatarPath))
}

func (r *UserRepository) CountAdmins(ctx context.Context) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE is_admin = TRUE`).Scan(&count)
	return count, err
}

func (r *UserRepository) scanUser(row pgx.Row) (*domain.User, error) {
	var user domain.User
	err := row.Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName, &user.Phone,
		&user.BirthDate, &user.Profession, &user.MemberSince, &user.AvatarPath,
		&user.IsAdmin, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan user: %w", err)
	}
	return &user, nil
}

func ScanUser(row pgx.Row) (*domain.User, error) {
	repo := &UserRepository{}
	return repo.scanUser(row)
}
