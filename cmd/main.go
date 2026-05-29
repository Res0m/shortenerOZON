package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"shortener/internal/config"
	"shortener/internal/handlers/getter"
	"shortener/internal/handlers/save"
	"shortener/internal/repository"
	"shortener/internal/repository/db"
	"shortener/internal/repository/memory"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

func main() {
	// Загрузка конфига
	cfg := config.MustLoad()

	// Выбор места хранения, либо memmory(внутренная память), либо postgres(База Данных PostgreSQL)
	var repo repository.Repository

	switch cfg.Settings.StorageType {
	case "memory", "":
		log.Println("Using in-memory storage")
		repo = memory.New()

	case "postgres":
		log.Println("Using PostgreSQL storage")

		postgresDB, err := db.NewDb(cfg)
		if err != nil {
			log.Fatalf("failed to connect to postgres: %v", err)
		}
		repo = postgresDB

	default:
		log.Fatalf("unknown storage type: %q", cfg.Settings.StorageType)
	}

	defer repo.Close()

	// Роутер chi + регистрация хендлеров
	r := chi.NewRouter()
	validate := validator.New()

	r.Post("/", save.New(repo, cfg.Settings.LengthShortUrl, validate))
	r.Get("/{hash}", getter.New(repo))

	// Создание сервера
	srv := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      r,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	// Graceful shutdown (опционально)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-stop
		log.Println("Shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("Shutdown error: %v", err)
		}
	}()

	log.Printf("Server starting on :8081")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
	log.Println("Server stopped")
}
