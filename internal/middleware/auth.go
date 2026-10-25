package middleware

import (
	"net/http"
	"strings"

	"agro-supply-erp/internal/auth"
	"agro-supply-erp/internal/response"
)

// Wrap menggabungkan handler function dengan serangkaian middleware
func Wrap(h http.HandlerFunc, middlewares ...func(http.Handler) http.Handler) http.HandlerFunc {
	var handler http.Handler = h
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler.ServeHTTP
}

// Authenticate memvalidasi header Authorization: Bearer <token>
func Authenticate(authService auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Token autentikasi tidak ditemukan. Silakan login terlebih dahulu.", nil)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Format header Authorization tidak valid. Gunakan format 'Bearer <token>'.", nil)
				return
			}

			tokenStr := strings.TrimSpace(parts[1])
			claims, err := authService.ValidateToken(tokenStr)
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Sesi Anda tidak valid atau telah kedaluwarsa. Silakan login kembali.", nil)
				return
			}

			ctx := auth.SetClaimsToContext(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole membatasi akses endpoint hanya untuk role tertentu
// Peran ADMIN secara otomatis diizinkan untuk semua endpoint yang diproteksi
func RequireRole(allowedRoles ...auth.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := auth.GetClaims(r.Context())
			if !ok || claims == nil {
				response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Pengguna tidak terotentikasi", nil)
				return
			}

			// Admin memiliki bypass akses penuh
			if claims.Role == auth.RoleAdmin {
				next.ServeHTTP(w, r)
				return
			}

			// Cek apakah role user ada dalam daftar yang diizinkan
			hasRole := false
			for _, role := range allowedRoles {
				if claims.Role == role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				response.Error(w, http.StatusForbidden, "FORBIDDEN", "Akses ditolak: Peran Anda tidak memiliki izin untuk melakukan tindakan ini.", nil)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
