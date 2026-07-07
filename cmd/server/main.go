package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/iPFSoftwares/custos/internal/api"
	"github.com/iPFSoftwares/custos/internal/api/handler"
	"github.com/iPFSoftwares/custos/internal/config"
	"github.com/iPFSoftwares/custos/internal/domain"
	"github.com/iPFSoftwares/custos/internal/notification/googlechat"
	"github.com/iPFSoftwares/custos/internal/notification/webhook"
	"github.com/iPFSoftwares/custos/internal/provider"
	kafkaimpl "github.com/iPFSoftwares/custos/internal/queue/kafka"
	"github.com/iPFSoftwares/custos/internal/service"
	pgstore "github.com/iPFSoftwares/custos/internal/store/postgres"
)

func main() {
	// 1. Configure structured logging. DEBUG level when env is development.
	logLevel := slog.LevelInfo
	if os.Getenv("CUSTOS_ENV") == "development" {
		logLevel = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})))

	// 2. Load config.
	cfg := config.Load()

	// 3. Connect Postgres and run migrations.
	db, err := pgstore.Open(cfg.DB.DSN)
	if err != nil {
		slog.Error("failed to connect to postgres", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()

	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		migrationsPath = "migrations"
	}
	slog.Info("running migrations", slog.String("path", migrationsPath))
	if err := pgstore.RunMigrations(db, migrationsPath); err != nil {
		slog.Error("failed to run migrations", slog.String("error", err.Error()))
		os.Exit(1) //nolint:gocritic // exitAfterDefer: startup failure — process exit cleans up the connection
	}

	// 4. Init Kafka producer and consumer.
	kafkaProducer := kafkaimpl.NewProducer(cfg)
	kafkaConsumer := kafkaimpl.NewConsumer(cfg)
	defer kafkaProducer.Close()
	defer kafkaConsumer.Close()

	// 5. Load AI provider (optional — system continues without analysis).
	analyzer, err := provider.Load(cfg)
	if err != nil {
		slog.Warn("AI provider not configured — analysis disabled",
			slog.String("reason", err.Error()))
	}

	// 6. Build stores.
	issueStore := pgstore.NewIssueStore(db)
	eventStore := pgstore.NewEventStore(db)
	projectStore := pgstore.NewProjectStore(db)
	apiKeyStore := pgstore.NewAPIKeyStore(db)

	// 7. Build services.
	notificationSvc := buildNotificationService(cfg)
	ingestionSvc := service.NewIngestionService(eventStore, issueStore, kafkaProducer, cfg)

	var analysisSvc *service.AnalysisService
	if analyzer != nil {
		analysisSvc = service.NewAnalysisService(issueStore, analyzer, notificationSvc, kafkaConsumer)
	}

	// 8. Build router.
	router := api.NewRouter(api.RouterDeps{
		DB:        db,
		APIKeys:   apiKeyStore,
		Projects:  projectStore,
		Ingest:    handler.NewIngestHandler(ingestionSvc),
		Issues:    handler.NewIssuesHandler(issueStore),
		Analytics: handler.NewAnalyticsHandler(issueStore),
		ProjectsH: handler.NewProjectsHandler(projectStore, apiKeyStore),
	})

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 9. Root context — cancelled on shutdown signal.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 10. Start analysis consumer as a supervised goroutine.
	if analysisSvc != nil {
		go runSupervised(ctx, "analysis-service", func() error {
			return analysisSvc.Run(ctx, cfg.Kafka.AnalysisTopic)
		})
	}

	// 11. Start HTTP server.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		slog.Info("starting server",
			slog.String("addr", srv.Addr),
			slog.String("env", cfg.Server.Env),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	<-quit
	slog.Info("shutting down server")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server forced to close", slog.String("error", err.Error()))
	}
	slog.Info("server exited")
}

// runSupervised runs fn in the current goroutine, recovering from any panic.
// It logs the panic or error and exits — a crashed background worker is fatal.
func runSupervised(ctx context.Context, name string, fn func() error) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("worker panicked", slog.String("worker", name), slog.Any("panic", r))
			os.Exit(1)
		}
	}()
	if err := fn(); err != nil && ctx.Err() == nil {
		slog.Error("worker stopped unexpectedly",
			slog.String("worker", name),
			slog.String("error", err.Error()),
		)
		os.Exit(1) //nolint:gocritic // exitAfterDefer: intentional — unexpected worker exit is fatal
	}
}

// buildNotificationService wires all enabled notification channels.
func buildNotificationService(cfg config.Config) *service.NotificationService {
	var notifiers []domain.Notifier
	if cfg.Notification.GoogleChatWebhookURL != "" {
		notifiers = append(notifiers, googlechat.New(cfg.Notification.GoogleChatWebhookURL))
	}
	if cfg.Notification.WebhookURL != "" {
		notifiers = append(notifiers, webhook.New(cfg.Notification.WebhookURL))
	}
	return service.NewNotificationService(notifiers...)
}
