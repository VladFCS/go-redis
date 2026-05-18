package main

import (
	"context"
	"database/sql"
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
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"
	"github.com/vladfc/go-redis/internal/reservation"
	"github.com/vladfc/go-redis/internal/room"
)

type Config struct {
	HTTPAddr      string
	PostgresDSN   string
	RedisAddr     string
	RedisPassword string
	RedisDB       int
}

func main() {
	cfg := loadConfig()

	postgresDB, err := connectPostgres(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := postgresDB.Close(); err != nil {
			log.Println("postgres close error:", err)
		}
	}()

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

	roomRepository := room.NewPostgreSQLRepository(postgresDB)
	roomCache := room.NewRedisRoomCache(redisClient)
	roomService := room.NewRoomService(roomRepository, roomCache)
	roomHandler := room.NewHandler(roomService)

	reservationRepository := reservation.NewRedisReservationRepository(redisClient)
	reservationService := reservation.NewReservationService(reservationRepository, roomService)
	reservationHandler := reservation.NewHandler(reservationService)

	router.Mount("/", reservationHandler.Routes())
	router.Mount("/", roomHandler.Routes())

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

func connectPostgres(cfg Config) (*sql.DB, error) {
	db, err := sql.Open("pgx", cfg.PostgresDSN)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return db, nil
}

func loadConfig() Config {
	return Config{
		HTTPAddr:      envOrDefault("HTTP_ADDR", ":8080"),
		PostgresDSN:   envOrDefault("POSTGRES_DSN", "postgres://postgres:postgres@localhost:5432/go_redis?sslmode=disable"),
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
	fmt.Println("PostgreSQL connected")
	fmt.Printf("Redis connected: %s (db=%d)\n", cfg.RedisAddr, cfg.RedisDB)
	fmt.Println("Available routes:")
	fmt.Println("GET /")
	fmt.Println("GET /healthz")
	fmt.Println("GET /rooms/")
	fmt.Println("POST /rooms/")
	fmt.Println("GET /rooms/{id}")
	fmt.Println("GET /reservations/{id}")
	fmt.Println("POST /reservations/")
	fmt.Println("POST /reservations/{id}/confirm")
}
