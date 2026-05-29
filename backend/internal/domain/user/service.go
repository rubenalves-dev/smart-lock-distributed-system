package user

import (
	"context"
	"fmt"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) RegisterTagIfNotExists(ctx context.Context, rfidUID string) (*User, error) {
	existing, err := s.repo.FindByUID(ctx, rfidUID)
	if err != nil {
		return nil, fmt.Errorf("failed to search user UID: %w", err)
	}
	if existing != nil {
		return existing, nil
	}

	u := &User{
		RfidUID: rfidUID,
	}
	if err := s.repo.Create(ctx, u); err != nil {
		return nil, fmt.Errorf("failed to register user UID: %w", err)
	}

	return u, nil
}

func (s *Service) UpdateUser(ctx context.Context, rfidUID string, name, email *string, isAccepted, isBlocked *bool) (*User, error) {
	u, err := s.repo.FindByUID(ctx, rfidUID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}
	if u == nil {
		u = &User{
			RfidUID: rfidUID,
		}
		if name != nil {
			u.Name = name
		}
		if email != nil {
			u.Email = email
		}
		if isAccepted != nil {
			u.IsAccepted = *isAccepted
		}
		if isBlocked != nil {
			u.IsBlocked = *isBlocked
		}
		if err := s.repo.Create(ctx, u); err != nil {
			return nil, fmt.Errorf("failed to register and update user: %w", err)
		}
		return u, nil
	}

	if name != nil {
		u.Name = name
	}
	if email != nil {
		u.Email = email
	}
	if isAccepted != nil {
		u.IsAccepted = *isAccepted
	}
	if isBlocked != nil {
		u.IsBlocked = *isBlocked
	}
	if err := s.repo.Update(ctx, u); err != nil {
		return nil, fmt.Errorf("failed to update user details: %w", err)
	}

	return u, nil
}

func (s *Service) GetAllUsers(ctx context.Context) ([]User, error) {
	return s.repo.ListAll(ctx)
}

func (s *Service) GetUserByUID(ctx context.Context, rfidUID string) (*User, error) {
	return s.repo.FindByUID(ctx, rfidUID)
}

func (s *Service) ProcessAuthentication(ctx context.Context, rfidUID string) (*User, error) {
	u, err := s.repo.ProcessMFAAuthentication(ctx, rfidUID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user for authentication: %w", err)
	}
	if u == nil {
		return nil, fmt.Errorf("authentication failed: unknown RFID UID %s", rfidUID)
	}
	return u, nil
}
