package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

var _ Repository = (*sqlRepository)(nil)

type Repository interface {
	Create(ctx context.Context, u *User) error
	Update(ctx context.Context, u *User) error
	FindByUID(ctx context.Context, rfidUID string) (*User, error)
	ListAll(ctx context.Context) ([]User, error)
	ProcessMFAAuthentication(ctx context.Context, rfidUID string) (*User, error)
}

type sqlRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &sqlRepository{db: db}
}

func (r *sqlRepository) Create(ctx context.Context, u *User) error {
	query := `
		INSERT INTO users (rfid_uid, name, email, is_accepted, is_blocked, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, created_at, updated_at`
	err := r.db.QueryRowContext(ctx, query, u.RfidUID, u.Name, u.Email, u.IsAccepted, u.IsBlocked).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *sqlRepository) Update(ctx context.Context, u *User) error {
	query := `
		UPDATE users
		SET name = $1, email = $2, is_accepted = $3, is_blocked = $4, updated_at = NOW()
		WHERE rfid_uid = $5
		RETURNING id, created_at, updated_at`
	err := r.db.QueryRowContext(ctx, query, u.Name, u.Email, u.IsAccepted, u.IsBlocked, u.RfidUID).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("user not found: %s", u.RfidUID)
		}
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

func (r *sqlRepository) FindByUID(ctx context.Context, rfidUID string) (*User, error) {
	query := `
		SELECT id, rfid_uid, name, email, is_accepted, is_blocked, created_at, updated_at
		FROM users
		WHERE rfid_uid = $1`
	var u User
	err := r.db.QueryRowContext(ctx, query, rfidUID).Scan(&u.ID, &u.RfidUID, &u.Name, &u.Email, &u.IsAccepted, &u.IsBlocked, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find user by UID: %w", err)
	}
	return &u, nil
}

func (r *sqlRepository) ListAll(ctx context.Context) ([]User, error) {
	query := `
		SELECT id, rfid_uid, name, email, is_accepted, is_blocked, created_at, updated_at
		FROM users
		ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	users := []User{}
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.RfidUID, &u.Name, &u.Email, &u.IsAccepted, &u.IsBlocked, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan user row: %w", err)
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}
	return users, nil
}

func (r *sqlRepository) ProcessMFAAuthentication(ctx context.Context, rfidUID string) (*User, error) {
	u, err := r.FindByUID(ctx, rfidUID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user for MFA: %w", err)
	}
	if u == nil {
		return nil, fmt.Errorf("user not found for MFA: %s", rfidUID)
	}

	return u, nil
}
