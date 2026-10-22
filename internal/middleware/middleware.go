package middleware

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"agro-supply-erp/internal/response"
)

// Logger mencatat setiap request HTTP masuk tanpa membocorkan data sensitif
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rw, r)

		// Sanitasi path/query agar parameter sensitif tidak tercatat di log
		cleanPath := sanitizeURL(r.URL.String())
		log.Printf("[%s] %s %d %v", r.Method, cleanPath, rw.statusCode, time.Since(start))
	})
}

// Recovery menangani unhandled panic di seluruh handler agar server tidak crash
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				// Cetak trace panic ke terminal server (hanya untuk developer)
				log.Printf("[PANIC RECOVERED] %v\nStack: %s", rec, string(debug.Stack()))

				// Kembalikan respon 500 yang aman ke client (tanpa membocorkan detail internal)
				response.Error(
					w,
					http.StatusInternalServerError,
					"INTERNAL_SERVER_ERROR",
					"Terjadi kesalahan internal pada server. Silakan hubungi administrator.",
					nil,
				)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// CORS menambahkan header Cross-Origin Resource Sharing
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// sanitizeURL menyamarkan nilai query sensitif seperti token, password, atau secret
func sanitizeURL(rawURL string) string {
	sensitiveKeys := []string{"password", "token", "secret", "key", "pass"}
	for _, key := range sensitiveKeys {
		if strings.Contains(strings.ToLower(rawURL), key+"=") {
			return fmt.Sprintf("%s [SANITIZED]", strings.Split(rawURL, "?")[0])
		}
	}
	return rawURL
}
