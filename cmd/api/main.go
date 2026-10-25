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

	"agro-supply-erp/internal/auth"
	"agro-supply-erp/internal/config"
	"agro-supply-erp/internal/database"
	"agro-supply-erp/internal/health"
	"agro-supply-erp/internal/inventory"
	"agro-supply-erp/internal/middleware"
	"agro-supply-erp/internal/procurement"
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
		// Auth Module
		authRepo := auth.NewRepository(db)
		authService := auth.NewService(authRepo, cfg.JWTSecret)
		authHandler := auth.NewHandler(authService)

		authMw := middleware.Authenticate(authService)
		requireAdmin := middleware.RequireRole(auth.RoleAdmin)
		requirePurchasing := middleware.RequireRole(auth.RolePurchasing)
		requireManager := middleware.RequireRole(auth.RoleManager)
		requireWarehouse := middleware.RequireRole(auth.RoleWarehouse)

		// 4.1 Auth Routes
		mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)
		mux.HandleFunc("GET /api/v1/auth/me", middleware.Wrap(authHandler.GetMe, authMw))
		mux.HandleFunc("POST /api/v1/auth/register", middleware.Wrap(authHandler.Register, authMw, requireAdmin))
		mux.HandleFunc("GET /api/v1/users", middleware.Wrap(authHandler.ListUsers, authMw, requireAdmin))
		mux.HandleFunc("PUT /api/v1/users/{id}/status", middleware.Wrap(authHandler.UpdateUserStatus, authMw, requireAdmin))

		// 4.2 Product Module
		productRepo := product.NewRepository(db)
		productService := product.NewService(productRepo)
		productHandler := product.NewHandler(productService)

		mux.HandleFunc("GET /api/v1/products", middleware.Wrap(productHandler.List, authMw))
		mux.HandleFunc("GET /api/v1/products/{id}", middleware.Wrap(productHandler.GetByID, authMw))
		mux.HandleFunc("POST /api/v1/products", middleware.Wrap(productHandler.Create, authMw, requirePurchasing))
		mux.HandleFunc("PUT /api/v1/products/{id}", middleware.Wrap(productHandler.Update, authMw, requirePurchasing))
		mux.HandleFunc("DELETE /api/v1/products/{id}", middleware.Wrap(productHandler.Delete, authMw, requirePurchasing))

		// 4.3 Supplier Module
		supplierRepo := supplier.NewRepository(db)
		supplierService := supplier.NewService(supplierRepo)
		supplierHandler := supplier.NewHandler(supplierService)

		mux.HandleFunc("GET /api/v1/suppliers", middleware.Wrap(supplierHandler.List, authMw))
		mux.HandleFunc("GET /api/v1/suppliers/{id}", middleware.Wrap(supplierHandler.GetByID, authMw))
		mux.HandleFunc("POST /api/v1/suppliers", middleware.Wrap(supplierHandler.Create, authMw, requirePurchasing))
		mux.HandleFunc("PUT /api/v1/suppliers/{id}", middleware.Wrap(supplierHandler.Update, authMw, requirePurchasing))
		mux.HandleFunc("DELETE /api/v1/suppliers/{id}", middleware.Wrap(supplierHandler.Delete, authMw, requirePurchasing))

		// 4.4 Procurement Module (PR & PO)
		procurementRepo := procurement.NewRepository(db)
		procurementService := procurement.NewService(procurementRepo, productRepo, supplierRepo)
		procurementHandler := procurement.NewHandler(procurementService)

		// Purchase Requests
		mux.HandleFunc("GET /api/v1/purchase-requests", middleware.Wrap(procurementHandler.ListPR, authMw))
		mux.HandleFunc("GET /api/v1/purchase-requests/{id}", middleware.Wrap(procurementHandler.GetPRByID, authMw))
		mux.HandleFunc("POST /api/v1/purchase-requests", middleware.Wrap(procurementHandler.CreatePR, authMw, requirePurchasing))
		mux.HandleFunc("POST /api/v1/purchase-requests/{id}/submit", middleware.Wrap(procurementHandler.SubmitPR, authMw, requirePurchasing))
		mux.HandleFunc("POST /api/v1/purchase-requests/{id}/approve", middleware.Wrap(procurementHandler.ApprovePR, authMw, requireManager))
		mux.HandleFunc("POST /api/v1/purchase-requests/{id}/reject", middleware.Wrap(procurementHandler.RejectPR, authMw, requireManager))

		// Purchase Orders
		mux.HandleFunc("GET /api/v1/purchase-orders", middleware.Wrap(procurementHandler.ListPO, authMw))
		mux.HandleFunc("GET /api/v1/purchase-orders/{id}", middleware.Wrap(procurementHandler.GetPOByID, authMw))
		mux.HandleFunc("POST /api/v1/purchase-orders", middleware.Wrap(procurementHandler.CreatePO, authMw, requirePurchasing))
		mux.HandleFunc("POST /api/v1/purchase-orders/{id}/approve", middleware.Wrap(procurementHandler.ApprovePO, authMw, requireManager))
		mux.HandleFunc("POST /api/v1/purchase-orders/{id}/cancel", middleware.Wrap(procurementHandler.CancelPO, authMw, requireManager))

		// 4.5 Inventory Module (Goods Receipt & Stock Ledger)
		inventoryRepo := inventory.NewRepository(db)
		inventoryService := inventory.NewService(inventoryRepo, procurementRepo)
		inventoryHandler := inventory.NewHandler(inventoryService)

		mux.HandleFunc("GET /api/v1/inventory/stock", middleware.Wrap(inventoryHandler.GetStockSummary, authMw))
		mux.HandleFunc("GET /api/v1/inventory/ledger", middleware.Wrap(inventoryHandler.ListTransactions, authMw))
		mux.HandleFunc("GET /api/v1/goods-receipts", middleware.Wrap(inventoryHandler.ListGoodsReceipts, authMw))
		mux.HandleFunc("GET /api/v1/goods-receipts/{id}", middleware.Wrap(inventoryHandler.GetGoodsReceiptByID, authMw))
		mux.HandleFunc("POST /api/v1/goods-receipts", middleware.Wrap(inventoryHandler.CreateGoodsReceipt, authMw, requireWarehouse))
	}

	// Middleware chain: Recovery -> CORS -> Logger -> Mux
	handler := middleware.Recovery(middleware.CORS(middleware.Logger(mux)))

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
