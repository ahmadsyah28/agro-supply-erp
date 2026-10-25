package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"agro-supply-erp/internal/database"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Repository interface {
	Create(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id int64) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	List(ctx context.Context) ([]*User, error)
	Update(ctx context.Context, u *User) error
	UpdateStatus(ctx context.Context, id int64, isActive bool) error
}

type pgRepository struct {
	db *database.DB
}

func NewRepository(db *database.DB) Repository {
	return &pgRepository{db: db}
}

func (r *pgRepository) Create(ctx context.Context, u *User) error {
	query := `
		INSERT INTO users (username, email, password_hash, full_name, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`
	now := time.Now().UTC()
	u.CreatedAt = now
	u.UpdatedAt = now

	err := r.db.Pool.QueryRow(ctx, query,
		u.Username,
		u.Email,
		u.PasswordHash,
		u.FullName,
		string(u.Role),
		u.IsActive,
		u.CreatedAt,
		u.UpdatedAt,
	).Scan(&u.ID)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if pgErr.ConstraintName == "users_username_key" || pgErr.ConstraintName == "idx_users_username" {
				return ErrUsernameExists
			}
			return ErrEmailExists
		}
		return fmt.Errorf("gagal membuat user: %w", err)
	}

	return nil
}

func (r *pgRepository) GetByID(ctx context.Context, id int64) (*User, error) {
	query := `
		SELECT id, username, email, password_hash, full_name, role, is_active, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	var u User
	var roleStr string
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&u.ID,
		&u.Username,
		&u.Email,
		&u.PasswordHash,
		&u.FullName,
		&roleStr,
		&u.IsActive,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("gagal mengambil user by id: %w", err)
	}
	u.Role = Role(roleStr)
	return &u, nil
}

func (r *pgRepository) GetByUsername(ctx context.Context, username string) (*User, error) {
	query := `
		SELECT id, username, email, password_hash, full_name, role, is_active, created_at, updated_at
		FROM users
		WHERE username = $1
	`
	var u User
	var roleStr string
	err := r.db.Pool.QueryRow(ctx, query, username).Scan(
		&u.ID,
		&u.Username,
		&u.Email,
		&u.PasswordHash,
		&u.FullName,
		&roleStr,
		&u.IsActive,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("gagal mengambil user by username: %w", err)
	}
	u.Role = Role(roleStr)
	return &u, nil
}

func (r *pgRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, username, email, password_hash, full_name, role, is_active, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	var u User
	var roleStr string
	err := r.db.Pool.QueryRow(ctx, query, email).Scan(
		&u.ID,
		&u.Username,
		&u.Email,
		&u.PasswordHash,
		&u.FullName,
		&roleStr,
		&u.IsActive,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("gagal mengambil user by email: %w", err)
	}
	u.Role = Role(roleStr)
	return &u, nil
}

func (r *pgRepository) List(ctx context.Context) ([]*User, error) {
	query := `
		SELECT id, username, email, password_hash, full_name, role, is_active, created_at, updated_at
		FROM users
		ORDER BY id ASC
	`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar user: %w", err)
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		var u User
		var roleStr string
		if err := rows.Scan(
			&u.ID,
			&u.Username,
			&u.Email,
			&u.PasswordHash,
			&u.FullName,
			&roleStr,
			&u.IsActive,
			&u.CreatedAt,
			&u.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("gagal scan user: %w", err)
		}
		u.Role = Role(roleStr)
		users = append(users, &u)
	}

	return users, nil
}

func (r *pgRepository) Update(ctx context.Context, u *User) error {
	query := `
		UPDATE users
		SET full_name = $1, role = $2, is_active = $3, updated_at = $4
		WHERE id = $5
	`
	now := time.Now().UTC()
	u.UpdatedAt = now

	res, err := r.db.Pool.Exec(ctx, query,
		u.FullName,
		string(u.Role),
		u.IsActive,
		u.UpdatedAt,
		u.ID,
	)
	if err != nil {
		return fmt.Errorf("gagal update user: %w", err)
	}
	if res.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *pgRepository) UpdateStatus(ctx context.Context, id int64, isActive bool) error {
	query := `
		UPDATE users
		SET is_active = $1, updated_at = $2
		WHERE id = $3
	`
	res, err := r.db.Pool.Exec(ctx, query, isActive, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("gagal update status user: %w", err)
	}
	if res.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}
