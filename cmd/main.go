package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"
)

type Config struct {
	HTTPAddr      string
	RedisAddr     string
	RedisPassword string
	RedisDB       int
}

func main() {
	cfg := loadConfig()

	redisClient, err := connectRedis(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Println("redis close error:", err)
		}
	}()

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(10 * time.Second))

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"service":"go-redis","status":"starting-point"}`))
	})

	router.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	logStartup(cfg)

	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- server.ListenAndServe()
	}()

	shutdownCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serverErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	case <-shutdownCtx.Done():
		log.Println("shutdown signal received")
	}

	gracefulCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(gracefulCtx); err != nil {
		log.Fatal(err)
	}
}

func connectRedis(cfg Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return client, nil
}

func loadConfig() Config {
	return Config{
		HTTPAddr:      envOrDefault("HTTP_ADDR", ":8080"),
		RedisAddr:     envOrDefault("REDIS_ADDR", "localhost:6379"),
		RedisPassword: envOrDefault("REDIS_PASSWORD", ""),
		RedisDB:       envIntOrDefault("REDIS_DB", 0),
	}
}

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func envIntOrDefault(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func logStartup(cfg Config) {
	fmt.Printf("HTTP server listening on http://localhost%s\n", cfg.HTTPAddr)
	fmt.Printf("Redis connected: %s (db=%d)\n", cfg.RedisAddr, cfg.RedisDB)
	fmt.Println("Available routes:")
	fmt.Println("GET /")
	fmt.Println("GET /healthz")
}
