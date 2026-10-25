package auth

import (
	"context"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

type mockUserRepo struct {
	users  map[string]*User
	usersByID map[int64]*User
	lastID int64
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users:     make(map[string]*User),
		usersByID: make(map[int64]*User),
	}
}

func (m *mockUserRepo) Create(ctx context.Context, u *User) error {
	if _, ok := m.users[u.Username]; ok {
		return ErrUsernameExists
	}
	m.lastID++
	u.ID = m.lastID
	m.users[u.Username] = u
	m.usersByID[u.ID] = u
	return nil
}

func (m *mockUserRepo) GetByID(ctx context.Context, id int64) (*User, error) {
	u, ok := m.usersByID[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (m *mockUserRepo) GetByUsername(ctx context.Context, username string) (*User, error) {
	u, ok := m.users[username]
	if !ok {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*User, error) {
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, ErrUserNotFound
}

func (m *mockUserRepo) List(ctx context.Context) ([]*User, error) {
	var list []*User
	for _, u := range m.usersByID {
		list = append(list, u)
	}
	return list, nil
}

func (m *mockUserRepo) Update(ctx context.Context, u *User) error {
	m.users[u.Username] = u
	m.usersByID[u.ID] = u
	return nil
}

func (m *mockUserRepo) UpdateStatus(ctx context.Context, id int64, isActive bool) error {
	u, ok := m.usersByID[id]
	if !ok {
		return ErrUserNotFound
	}
	u.IsActive = isActive
	return nil
}

func TestAuthService_Login_Success(t *testing.T) {
	repo := newMockUserRepo()
	hashed, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	repo.Create(context.Background(), &User{
		Username:     "admin",
		Email:        "admin@test.com",
		PasswordHash: string(hashed),
		FullName:     "Admin Test",
		Role:         RoleAdmin,
		IsActive:     true,
	})

	svc := NewService(repo, "secret-key-12345")
	resp, err := svc.Login(context.Background(), LoginRequest{
		Username: "admin",
		Password: "password123",
	})

	if err != nil {
		t.Fatalf("expected login success, got error: %v", err)
	}

	if resp.Token == "" {
		t.Errorf("expected non-empty token")
	}

	if resp.User.Role != RoleAdmin {
		t.Errorf("expected role ADMIN, got %v", resp.User.Role)
	}

	// Test ValidateToken
	claims, err := svc.ValidateToken(resp.Token)
	if err != nil {
		t.Fatalf("expected valid token, got error: %v", err)
	}

	if claims.Username != "admin" || claims.Role != RoleAdmin {
		t.Errorf("claims mismatch: %+v", claims)
	}
}

func TestAuthService_Login_InvalidPassword(t *testing.T) {
	repo := newMockUserRepo()
	hashed, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	repo.Create(context.Background(), &User{
		Username:     "admin",
		Email:        "admin@test.com",
		PasswordHash: string(hashed),
		FullName:     "Admin Test",
		Role:         RoleAdmin,
		IsActive:     true,
	})

	svc := NewService(repo, "secret-key-12345")
	_, err := svc.Login(context.Background(), LoginRequest{
		Username: "admin",
		Password: "wrongpassword",
	})

	if err != ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got: %v", err)
	}
}

func TestAuthService_Login_InactiveUser(t *testing.T) {
	repo := newMockUserRepo()
	hashed, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	repo.Create(context.Background(), &User{
		Username:     "budi",
		Email:        "budi@test.com",
		PasswordHash: string(hashed),
		FullName:     "Budi Inactive",
		Role:         RolePurchasing,
		IsActive:     false,
	})

	svc := NewService(repo, "secret-key-12345")
	_, err := svc.Login(context.Background(), LoginRequest{
		Username: "budi",
		Password: "password123",
	})

	if err != ErrUserInactive {
		t.Fatalf("expected ErrUserInactive, got: %v", err)
	}
}

func TestAuthService_Register_Validation(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewService(repo, "secret-key-12345")

	// Invalid role
	_, err := svc.Register(context.Background(), RegisterRequest{
		Username: "test",
		Email:    "test@test.com",
		Password: "password123",
		FullName: "Test User",
		Role:     "HACKER",
	})

	if err != ErrInvalidRole {
		t.Fatalf("expected ErrInvalidRole, got: %v", err)
	}

	// Valid registration
	user, err := svc.Register(context.Background(), RegisterRequest{
		Username: "warehouse1",
		Email:    "wh@test.com",
		Password: "password123",
		FullName: "Joko Gudang",
		Role:     RoleWarehouse,
	})

	if err != nil {
		t.Fatalf("expected registration success, got error: %v", err)
	}

	if user.Role != RoleWarehouse || user.Username != "warehouse1" {
		t.Errorf("unexpected registered user: %+v", user)
	}
}
