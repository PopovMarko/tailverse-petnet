package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	announcement_repository "github.com/PopovMarko/tailverse-petnet/internal/announcement/repository"
	announcement_service "github.com/PopovMarko/tailverse-petnet/internal/announcement/service"
	announcement_transport_http "github.com/PopovMarko/tailverse-petnet/internal/announcement/transport/http"
	auth_repository "github.com/PopovMarko/tailverse-petnet/internal/auth/repository"
	auth_service "github.com/PopovMarko/tailverse-petnet/internal/auth/service"
	auth_transport_http "github.com/PopovMarko/tailverse-petnet/internal/auth/transport/http"
	core_auth "github.com/PopovMarko/tailverse-petnet/internal/core/auth"
	"github.com/PopovMarko/tailverse-petnet/internal/core/logger"
	core_postgres "github.com/PopovMarko/tailverse-petnet/internal/core/repository/postgres"
	core_redis "github.com/PopovMarko/tailverse-petnet/internal/core/repository/redis"
	core_server_http "github.com/PopovMarko/tailverse-petnet/internal/core/server/http"
	"github.com/PopovMarko/tailverse-petnet/internal/core/transport/http/middleware"
	core_http_response "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http/response"
	feed_repository "github.com/PopovMarko/tailverse-petnet/internal/feed/repository"
	feed_service "github.com/PopovMarko/tailverse-petnet/internal/feed/service"
	feed_transport_http "github.com/PopovMarko/tailverse-petnet/internal/feed/transport/http"
	owner_repository "github.com/PopovMarko/tailverse-petnet/internal/owner/repository"
	owner_service "github.com/PopovMarko/tailverse-petnet/internal/owner/service"
	owner_transport_http "github.com/PopovMarko/tailverse-petnet/internal/owner/transport/http"
	pet_repository "github.com/PopovMarko/tailverse-petnet/internal/pet/repository"
	pet_service "github.com/PopovMarko/tailverse-petnet/internal/pet/service"
	"github.com/PopovMarko/tailverse-petnet/internal/pet/transport/http"
	presence_repository "github.com/PopovMarko/tailverse-petnet/internal/presence/repository"
	presence_service "github.com/PopovMarko/tailverse-petnet/internal/presence/service"
	services_repository "github.com/PopovMarko/tailverse-petnet/internal/services/repository"
	services_service "github.com/PopovMarko/tailverse-petnet/internal/services/service"
	services_transport_http "github.com/PopovMarko/tailverse-petnet/internal/services/transport/http"
	upload_repository "github.com/PopovMarko/tailverse-petnet/internal/upload/repository"
	upload_service "github.com/PopovMarko/tailverse-petnet/internal/upload/service"
	upload_transport_http "github.com/PopovMarko/tailverse-petnet/internal/upload/transport/http"
	walkspot_repository "github.com/PopovMarko/tailverse-petnet/internal/walkspot/repository"
	walkspot_service "github.com/PopovMarko/tailverse-petnet/internal/walkspot/service"
	walkspot_transport_http "github.com/PopovMarko/tailverse-petnet/internal/walkspot/transport/http"
	"github.com/PopovMarko/tailverse-petnet/internal/ws"
	"github.com/PopovMarko/tailverse-petnet/migrations"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func main() {
	migrateOnly := flag.Bool("migrate-only", false, "apply database migrations and exit")
	flag.Parse()

	loggerConfig := core_logger.NewLoggerConfigMust()
	logger, err := core_logger.NewLogger(loggerConfig)
	if err != nil {
		log.Fatalf("Failed to configure logger: %s", err.Error())
	}
	defer logger.Close()
	defer logger.Sync()
	logger.Debug("Initialized logger")

	if err := run(logger, *migrateOnly); err != nil {
		logger.Error("application stopped with error", zap.Error(err))
		logger.Sync()
		os.Exit(1)
	}
	logger.Info("application stopped")
}

func run(logger *core_logger.Logger, migrateOnly bool) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Infrastructure
	pool, err := core_postgres.NewPool(ctx, core_postgres.NewPostgresConfigMust())
	if err != nil {
		return fmt.Errorf("connect to postgres: %w", err)
	}
	defer pool.Close()
	logger.Info("connected to postgres")

	applied, err := core_postgres.Migrate(ctx, pool, migrations.FS)
	if err != nil {
		return fmt.Errorf("migrate database: %w", err)
	}
	logger.Info("database migrated", zap.Strings("applied", applied))
	if migrateOnly {
		return nil
	}

	redisClient, err := core_redis.NewClient(ctx, core_redis.NewRedisConfigMust())
	if err != nil {
		return fmt.Errorf("connect to redis: %w", err)
	}
	defer redisClient.Close()
	logger.Info("connected to redis")

	uploadConfig := upload_service.NewUploadConfigMust()
	uploadRepository, err := upload_repository.NewDiskRepository(uploadConfig.Dir)
	if err != nil {
		return fmt.Errorf("prepare uploads storage: %w", err)
	}
	logger.Info("uploads stored on disk", zap.String("dir", uploadConfig.Dir))

	tokenManager := core_auth.NewTokenManager(core_auth.NewAuthConfigMust())
	authMiddleware := core_http_middleware.Auth(tokenManager)
	hub := ws.NewHub(logger)

	// Services
	petService := pet_service.NewPetService(pet_repository.NewPetRepository(pool))
	authService := auth_service.NewAuthService(auth_repository.NewOwnerRepository(pool), tokenManager)
	ownerService := owner_service.NewOwnerService(owner_repository.NewOwnerRepository(pool))
	uploadService := upload_service.NewUploadService(uploadRepository)
	presenceService := presence_service.NewPresenceService(
		presence_repository.NewPresenceRepository(redisClient),
		hub,
		presence_service.NewPresenceConfigMust().TTL,
	)
	walkSpotService := walkspot_service.NewWalkSpotService(walkspot_repository.NewWalkSpotRepository(pool), presenceService, petService)
	announcementService := announcement_service.NewAnnouncementService(announcement_repository.NewAnnouncementRepository(pool), petService, hub)
	feedService := feed_service.NewFeedService(feed_repository.NewPostRepository(pool), petService)
	serviceOfferService := services_service.NewServiceOfferService(services_repository.NewServiceOfferRepository(pool))

	// HTTP
	r := chi.NewRouter()
	r.Use(core_http_middleware.RequestID())
	r.Use(core_http_middleware.CORS())
	r.Use(core_http_middleware.Logger(logger))
	r.Use(core_http_middleware.Trace())
	r.Use(core_http_middleware.Panic())

	uploadHttpHandler := upload_transport_http.NewUploadHttpHandler(uploadService, uploadConfig.PublicBaseUrl)

	r.Get("/health", healthHandler(pool, redisClient))
	r.Mount("/uploads", upload_transport_http.NewUploadedFilesRouter(uploadHttpHandler))
	r.Route("/api/v1", func(r chi.Router) {
		r.Mount("/auth", auth_transport_http.NewAuthRouter(auth_transport_http.NewAuthHttpHandler(authService)))
		r.Mount("/owners", owner_transport_http.NewOwnersRouter(owner_transport_http.NewOwnerHttpHandler(ownerService), authMiddleware))
		r.Mount("/pets", pet_transport_http.NewPetsRouter(pet_transport_http.NewPetHttpHandler(petService), authMiddleware))
		r.Mount("/walkspots", walkspot_transport_http.NewWalkSpotsRouter(walkspot_transport_http.NewWalkSpotHttpHandler(walkSpotService), authMiddleware))
		r.Mount("/announcements", announcement_transport_http.NewAnnouncementsRouter(announcement_transport_http.NewAnnouncementHttpHandler(announcementService), authMiddleware))
		r.Mount("/posts", feed_transport_http.NewPostsRouter(feed_transport_http.NewFeedHttpHandler(feedService), authMiddleware))
		r.Mount("/services", services_transport_http.NewServicesRouter(services_transport_http.NewServicesHttpHandler(serviceOfferService), authMiddleware))
		r.Mount("/uploads", upload_transport_http.NewUploadsRouter(uploadHttpHandler, authMiddleware))
		r.Mount("/ws", ws.NewWsRouter(ws.NewWsHttpHandler(hub), authMiddleware))
	})

	server := core_server_http.NewHttpServer(core_server_http.NewHttpServerConfigMust(), r, logger)
	server.RegisterOnShutdown(hub.Close)
	return server.Run(ctx)
}

// healthHandler reports 200 when Postgres and Redis answer, 503 otherwise.
func healthHandler(pool *pgxpool.Pool, redisClient *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		status := map[string]string{"postgres": "ok", "redis": "ok"}
		code := http.StatusOK
		if err := pool.Ping(ctx); err != nil {
			status["postgres"] = err.Error()
			code = http.StatusServiceUnavailable
		}
		if err := redisClient.Ping(ctx).Err(); err != nil {
			status["redis"] = err.Error()
			code = http.StatusServiceUnavailable
		}

		core_http_response.NewHttpResponseHandler(core_logger.FromContext(r.Context()), w).JsonResponse(status, code)
	}
}
