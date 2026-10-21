package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"agro-supply-erp/internal/config"
	"agro-supply-erp/internal/database"
	"agro-supply-erp/internal/health"
	"agro-supply-erp/internal/middleware"
	"agro-supply-erp/internal/product"
	"agro-supply-erp/internal/supplier"
	"agro-supply-erp/migrations"
)

func main() {
	log.Println("=== Starting PT AgroSupply ERP Backend ===")

	// 1. Load Config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("[FATAL] Failed to load configuration: %v", err)
	}

	log.Printf("[INFO] Environment: %s", cfg.AppEnv)
	log.Printf("[INFO] HTTP Addr: %s", cfg.HTTPAddr)

	// 2. Initialize Database Connection
	var db *database.DB
	if cfg.DatabaseURL == "" {
		log.Println("[WARN] DATABASE_URL is not set. Please set DATABASE_URL in your .env file.")
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		dbInstance, err := database.NewPool(ctx, cfg.DatabaseURL)
		cancel()

		if err != nil {
			log.Printf("[WARN] Failed to connect to database: %v", err)
		} else {
			db = dbInstance
			defer db.Close()
			log.Println("[INFO] Successfully connected to PostgreSQL database")

			// 3. Run Migrations
			migrationRunner := database.NewMigrationRunner(db, migrations.FS)
			migCtx, migCancel := context.WithTimeout(context.Background(), 30*time.Second)
			if err := migrationRunner.RunMigrations(migCtx); err != nil {
				log.Printf("[ERROR] Migration failure: %v", err)
			} else {
				log.Println("[INFO] All database migrations are up to date")
			}
			migCancel()
		}
	}

	// 4. Setup Routes & Router
	mux := http.NewServeMux()

	healthHandler := health.NewHandler(db)
	mux.HandleFunc("/health/live", healthHandler.Live)
	mux.HandleFunc("/health/ready", healthHandler.Ready)

	// Register Business Modules
	if db != nil {
		// Product Module
		productRepo := product.NewRepository(db)
		productService := product.NewService(productRepo)
		productHandler := product.NewHandler(productService)
		productHandler.RegisterRoutes(mux)

		// Supplier Module
		supplierRepo := supplier.NewRepository(db)
		supplierService := supplier.NewService(supplierRepo)
		supplierHandler := supplier.NewHandler(supplierService)
		supplierHandler.RegisterRoutes(mux)
	}

	// Middleware chain: CORS -> Logger -> Mux
	handler := middleware.CORS(middleware.Logger(mux))

	// 5. Server Configuration
	server := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 6. Start Server in goroutine
	go func() {
		log.Printf("[INFO] Server listening on %s", cfg.HTTPAddr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("[FATAL] HTTP server error: %v", err)
		}
	}()

	// 7. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("[INFO] Shutting down server gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("[ERROR] Server forced to shutdown: %v", err)
	}

	log.Println("[INFO] Server stopped gracefully")
}
