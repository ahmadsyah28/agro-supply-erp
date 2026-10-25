package auth

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Role string

const (
	RoleAdmin      Role = "ADMIN"
	RolePurchasing Role = "PURCHASING"
	RoleManager    Role = "MANAGER"
	RoleWarehouse  Role = "WAREHOUSE"
)

func (r Role) IsValid() bool {
	switch r {
	case RoleAdmin, RolePurchasing, RoleManager, RoleWarehouse:
		return true
	default:
		return false
	}
}

var (
	ErrInvalidCredentials = errors.New("username atau password salah")
	ErrUserNotFound       = errors.New("pengguna tidak ditemukan")
	ErrUsernameExists     = errors.New("username sudah digunakan")
	ErrEmailExists        = errors.New("email sudah terdaftar")
	ErrUserInactive       = errors.New("akun pengguna telah dinonaktifkan")
	ErrInvalidRole        = errors.New("peran (role) tidak valid")
	ErrUnauthorized       = errors.New("autentikasi diperlukan")
	ErrForbidden          = errors.New("akses ditolak: Anda tidak memiliki izin untuk tindakan ini")
)

type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	FullName     string    `json:"full_name"`
	Role         Role      `json:"role"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type UserDTO struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
	Role      Role      `json:"role"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (u *User) ToDTO() UserDTO {
	return UserDTO{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		FullName:  u.FullName,
		Role:      u.Role,
		IsActive:  u.IsActive,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
	Role     Role   `json:"role"`
}

type AuthResponse struct {
	Token string  `json:"token"`
	User  UserDTO `json:"user"`
}

type JWTClaims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Role     Role   `json:"role"`
	FullName string `json:"full_name"`
	jwt.RegisteredClaims
}

type contextKey string

const (
	ClaimsKey contextKey = "user_claims"
)

// SetClaimsToContext menyisipkan JWTClaims ke dalam context request
func SetClaimsToContext(ctx context.Context, claims *JWTClaims) context.Context {
	return context.WithValue(ctx, ClaimsKey, claims)
}

// GetClaims mengambil JWTClaims dari context request
func GetClaims(ctx context.Context) (*JWTClaims, bool) {
	claims, ok := ctx.Value(ClaimsKey).(*JWTClaims)
	return claims, ok
}

