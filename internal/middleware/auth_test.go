package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"agro-supply-erp/internal/auth"
)

type mockAuthService struct {
	validToken string
	claims     *auth.JWTClaims
}

func (m *mockAuthService) Login(ctx context.Context, req auth.LoginRequest) (*auth.AuthResponse, error) {
	return nil, nil
}
func (m *mockAuthService) Register(ctx context.Context, req auth.RegisterRequest) (*auth.UserDTO, error) {
	return nil, nil
}
func (m *mockAuthService) GetProfile(ctx context.Context, userID int64) (*auth.UserDTO, error) {
	return nil, nil
}
func (m *mockAuthService) ListUsers(ctx context.Context) ([]auth.UserDTO, error) {
	return nil, nil
}
func (m *mockAuthService) UpdateUserStatus(ctx context.Context, id int64, isActive bool) error {
	return nil
}
func (m *mockAuthService) ValidateToken(tokenStr string) (*auth.JWTClaims, error) {
	if tokenStr == m.validToken {
		return m.claims, nil
	}
	return nil, auth.ErrUnauthorized
}

func TestMiddleware_Authenticate(t *testing.T) {
	mockSvc := &mockAuthService{
		validToken: "good-token-123",
		claims: &auth.JWTClaims{
			UserID:   1,
			Username: "admin",
			Role:     auth.RoleAdmin,
		},
	}

	authMw := Authenticate(mockSvc)
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := auth.GetClaims(r.Context())
		if !ok || claims.Username != "admin" {
			t.Errorf("failed to get expected claims from context")
		}
		w.WriteHeader(http.StatusOK)
	})

	// 1. Missing Authorization header
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	authMw(dummyHandler).ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for missing header, got %d", rr.Code)
	}

	// 2. Invalid Token
	req = httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer bad-token")
	rr = httptest.NewRecorder()
	authMw(dummyHandler).ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for bad token, got %d", rr.Code)
	}

	// 3. Valid Token
	req = httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer good-token-123")
	rr = httptest.NewRecorder()
	authMw(dummyHandler).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 for valid token, got %d", rr.Code)
	}
}

func TestMiddleware_RequireRole(t *testing.T) {
	purchasingClaims := &auth.JWTClaims{
		UserID:   2,
		Username: "purchasing_user",
		Role:     auth.RolePurchasing,
	}

	adminClaims := &auth.JWTClaims{
		UserID:   1,
		Username: "admin",
		Role:     auth.RoleAdmin,
	}

	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	managerOnly := RequireRole(auth.RoleManager)

	// 1. User with Purchasing role tries to access Manager-only route -> 403
	req := httptest.NewRequest("POST", "/approve-pr", nil)
	req = req.WithContext(auth.SetClaimsToContext(req.Context(), purchasingClaims))
	rr := httptest.NewRecorder()
	managerOnly(dummyHandler).ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for purchasing accessing manager route, got %d", rr.Code)
	}

	// 2. User with Admin role tries to access Manager-only route -> 200 OK (Admin Bypass)
	req = httptest.NewRequest("POST", "/approve-pr", nil)
	req = req.WithContext(auth.SetClaimsToContext(req.Context(), adminClaims))
	rr = httptest.NewRecorder()
	managerOnly(dummyHandler).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK for admin bypass, got %d", rr.Code)
	}
}
