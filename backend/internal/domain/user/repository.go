package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type Repository interface {
	Create(ctx context.Context, u *User) error
	Update(ctx context.Context, u *User) error
	FindByUID(ctx context.Context, rfidUID string) (*User, error)
	ListAll(ctx context.Context) ([]User, error)
}

type sqlRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &sqlRepository{db: db}
}

func (r *sqlRepository) Create(ctx context.Context, u *User) error {
	query := `
		INSERT INTO users (rfid_uid, name, email, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id, created_at, updated_at`
	err := r.db.QueryRowContext(ctx, query, u.RfidUID, u.Name, u.Email).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *sqlRepository) Update(ctx context.Context, u *User) error {
	query := `
		UPDATE users
		SET name = $1, email = $2, updated_at = NOW()
		WHERE rfid_uid = $3
		RETURNING id, created_at, updated_at`
	err := r.db.QueryRowContext(ctx, query, u.Name, u.Email, u.RfidUID).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
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
		SELECT id, rfid_uid, name, email, created_at, updated_at
		FROM users
		WHERE rfid_uid = $1`
	var u User
	err := r.db.QueryRowContext(ctx, query, rfidUID).Scan(&u.ID, &u.RfidUID, &u.Name, &u.Email, &u.CreatedAt, &u.UpdatedAt)
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
		SELECT id, rfid_uid, name, email, created_at, updated_at
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
		if err := rows.Scan(&u.ID, &u.RfidUID, &u.Name, &u.Email, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan user row: %w", err)
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}
	return users, nil
}
