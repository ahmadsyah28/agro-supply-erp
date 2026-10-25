package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Login(ctx context.Context, req LoginRequest) (*AuthResponse, error)
	Register(ctx context.Context, req RegisterRequest) (*UserDTO, error)
	GetProfile(ctx context.Context, userID int64) (*UserDTO, error)
	ListUsers(ctx context.Context) ([]UserDTO, error)
	UpdateUserStatus(ctx context.Context, id int64, isActive bool) error
	ValidateToken(tokenStr string) (*JWTClaims, error)
}

type service struct {
	repo      Repository
	jwtSecret []byte
}

func NewService(repo Repository, jwtSecret string) Service {
	return &service{
		repo:      repo,
		jwtSecret: []byte(jwtSecret),
	}
}

func (s *service) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	req.Username = strings.TrimSpace(req.Username)
	req.Password = strings.TrimSpace(req.Password)

	if req.Username == "" || req.Password == "" {
		return nil, errors.New("username dan password wajib diisi")
	}

	user, err := s.repo.GetByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if !user.IsActive {
		return nil, ErrUserInactive
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Buat JWT Token
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		FullName: user.FullName,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "agro-supply-erp",
			Subject:   fmt.Sprintf("%d", user.ID),
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat token autentikasi: %w", err)
	}

	dto := user.ToDTO()
	return &AuthResponse{
		Token: tokenString,
		User:  dto,
	}, nil
}

func (s *service) Register(ctx context.Context, req RegisterRequest) (*UserDTO, error) {
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.FullName = strings.TrimSpace(req.FullName)
	req.Role = Role(strings.ToUpper(strings.TrimSpace(string(req.Role))))

	if req.Username == "" {
		return nil, errors.New("username wajib diisi")
	}
	if req.Email == "" || !strings.Contains(req.Email, "@") {
		return nil, errors.New("format email tidak valid")
	}
	if len(req.Password) < 6 {
		return nil, errors.New("password minimal harus 6 karakter")
	}
	if req.FullName == "" {
		return nil, errors.New("nama lengkap wajib diisi")
	}
	if !req.Role.IsValid() {
		return nil, ErrInvalidRole
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("gagal mengenkripsi password: %w", err)
	}

	user := &User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		FullName:     req.FullName,
		Role:         req.Role,
		IsActive:     true,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	dto := user.ToDTO()
	return &dto, nil
}

func (s *service) GetProfile(ctx context.Context, userID int64) (*UserDTO, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	dto := user.ToDTO()
	return &dto, nil
}

func (s *service) ListUsers(ctx context.Context) ([]UserDTO, error) {
	users, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	dtos := make([]UserDTO, len(users))
	for i, u := range users {
		dtos[i] = u.ToDTO()
	}
	return dtos, nil
}

func (s *service) UpdateUserStatus(ctx context.Context, id int64, isActive bool) error {
	return s.repo.UpdateStatus(ctx, id, isActive)
}

func (s *service) ValidateToken(tokenStr string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("metode signing tidak valid: %v", t.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("token tidak valid atau telah kedaluwarsa")
	}

	return claims, nil
}
