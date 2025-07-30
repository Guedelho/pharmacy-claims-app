package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"pharmacyclaims/internal/core"
	"pharmacyclaims/internal/database"
	"pharmacyclaims/internal/handlers"
	"pharmacyclaims/internal/repository"
	"pharmacyclaims/internal/service"
)

func main() {
	cfg := core.LoadConfig()

	if err := database.WaitForConnection(cfg.Database, 10, 2*time.Second); err != nil {
		log.Fatalf("Database readiness check failed: %v", err)
	}

	if err := database.RunMigrations(cfg.Database, cfg.MigrationsDir); err != nil {
		log.Fatalf("Failed to apply migrations: %v", err)
	}

	db, err := database.NewConnection(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	repo := repository.NewPostgresRepository(db)

	fileLogger := core.NewLogger(cfg.LogDir)

	loaderService := service.NewLoaderService(repo, fileLogger)
	claimsService := service.NewClaimsService(repo, fileLogger)

	if err := loaderService.LoadPharmaciesFromData(cfg.DataDir); err != nil {
		log.Printf("Warning: Failed to load pharmacy data: %v", err)
	}

	if err := loaderService.LoadClaimsFromData(cfg.DataDir); err != nil {
		log.Printf("Warning: Failed to load claims data: %v", err)
	}

	if err := loaderService.LoadReversalsFromData(cfg.DataDir); err != nil {
		log.Printf("Warning: Failed to load reversals data: %v", err)
	}

	handler := handlers.NewHttpHandler(claimsService)

	router := handler.SetupRoutes()

	server := &http.Server{
		Addr:         ":" + strconv.Itoa(cfg.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Starting server on port %d", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Failed to gracefully shutdown server: %v", err)
	}

	log.Println("Server shutdown complete")
}
