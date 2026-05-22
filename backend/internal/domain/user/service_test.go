package user

import (
	"context"
	"testing"
)

type mockRepository struct {
	users map[string]*User
}

func (m *mockRepository) Create(ctx context.Context, u *User) error {
	m.users[u.RfidUID] = u
	u.ID = len(m.users)
	return nil
}

func (m *mockRepository) Update(ctx context.Context, u *User) error {
	m.users[u.RfidUID] = u
	return nil
}

func (m *mockRepository) FindByUID(ctx context.Context, rfidUID string) (*User, error) {
	u, ok := m.users[rfidUID]
	if !ok {
		return nil, nil
	}
	return u, nil
}

func (m *mockRepository) ListAll(ctx context.Context) ([]User, error) {
	var list []User
	for _, u := range m.users {
		list = append(list, *u)
	}
	return list, nil
}

func TestRegisterTagIfNotExists(t *testing.T) {
	repo := &mockRepository{users: make(map[string]*User)}
	svc := NewService(repo)

	// Register new tag
	u1, err := svc.RegisterTagIfNotExists(context.Background(), "AA:BB:CC:DD")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u1.RfidUID != "AA:BB:CC:DD" {
		t.Errorf("expected UID AA:BB:CC:DD, got %s", u1.RfidUID)
	}

	// Register same tag again, should return existing
	u2, err := svc.RegisterTagIfNotExists(context.Background(), "AA:BB:CC:DD")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u1.ID != u2.ID {
		t.Errorf("expected same user ID, got %d and %d", u1.ID, u2.ID)
	}
}

func TestUpdateUser(t *testing.T) {
	repo := &mockRepository{users: make(map[string]*User)}
	svc := NewService(repo)

	// First, update a non-existent user (should register it)
	name := "Alice"
	email := "alice@example.com"
	u1, err := svc.UpdateUser(context.Background(), "11:22:33:44", &name, &email)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u1.Name == nil || *u1.Name != "Alice" {
		t.Errorf("expected name Alice, got %v", u1.Name)
	}

	// Next, update details for the same user
	newName := "Bob"
	u2, err := svc.UpdateUser(context.Background(), "11:22:33:44", &newName, &email)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u2.Name == nil || *u2.Name != "Bob" {
		t.Errorf("expected updated name Bob, got %v", u2.Name)
	}
}
