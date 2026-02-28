package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/cosmos-arbengine/backend/config"
	"github.com/cosmos-arbengine/backend/internal/api"
	"github.com/cosmos-arbengine/backend/internal/db"
	"github.com/cosmos-arbengine/backend/internal/engine"
	"github.com/cosmos-arbengine/backend/internal/events"
	"github.com/cosmos-arbengine/backend/internal/feeds"
	"github.com/cosmos-arbengine/backend/internal/relay"
	"github.com/cosmos-arbengine/backend/internal/types"
	"github.com/cosmos-arbengine/backend/internal/ws"
	"go.uber.org/zap"
)

func main() {
	// Logger
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	logger.Info("🚀 Cosmos ArbEngine Pro starting...")

	// Config
	cfg := config.Load()
	logger.Info("config loaded",
		zap.String("serverPort", cfg.Server.Port),
		zap.Bool("mockFeeds", cfg.Feeds.UseMock),
		zap.Int("chains", len(cfg.Feeds.Chains)),
	)

	// Context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Database
	database, err := db.NewDatabase(ctx, cfg.Database.URL, cfg.Database.MaxConns, logger)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer database.Close()

	// Run migrations
	migrationsDir := findMigrationsDir()
	if err := database.RunMigrations(ctx, migrationsDir); err != nil {
		logger.Fatal("failed to run migrations", zap.Error(err))
	}
	logger.Info("migrations complete")

	// Repository
	repo := db.NewRepository(database)

	// Event Bus
	eventBus := events.NewEventBus(logger)

	// Price Cache
	priceCache := feeds.NewPriceCache(logger)
	priceCache.StartCleanup(ctx, 30*time.Second)

	// WebSocket Hub
	hub := ws.NewHub(logger)
	go hub.Run()

	// Price Feeds
	feedManager := feeds.NewManager(priceCache, eventBus, logger)
	if cfg.Feeds.UseMock {
		logger.Info("using mock price feeds")
		for _, chainCfg := range cfg.Feeds.Chains {
			if chainCfg.Enabled {
				chainID := types.ChainID(chainCfg.ID)
				feed := feeds.NewMockFeed(chainID, priceCache, eventBus, logger, cfg.Feeds.PollInterval)
				feedManager.AddFeed(feed)
			}
		}
	}
	feedManager.StartAll(ctx)
	logger.Info("price feeds started", zap.Int("feedCount", feedManager.FeedCount()))

	// Arb Engine
	arbEngine := engine.NewArbEngine(
		priceCache, eventBus, repo, logger,
		cfg.Engine.MinNetProfitUSD,
		cfg.Engine.InputAmountUSD,
		cfg.Engine.OpportunityTTL,
	)
	go arbEngine.Start(ctx)
	logger.Info("arb engine started")

	// IBC Relay Monitor
	relayMonitor := relay.NewMonitor(
		repo, eventBus, logger,
		cfg.Relay.PollInterval,
		cfg.Relay.PendingPacketThreshold,
		cfg.Relay.StuckPacketAgeSec,
	)
	go relayMonitor.Start(ctx)
	logger.Info("IBC relay monitor started")

	// WebSocket broadcaster: forward opportunity events to WS clients
	go func() {
		oppCh := eventBus.Subscribe(events.EventOpportunity)
		for {
			select {
			case event := <-oppCh:
				hub.BroadcastJSON("opportunity", event.Payload)
			case <-ctx.Done():
				return
			}
		}
	}()

	// Forward relay alerts to WS
	go func() {
		relayCh := eventBus.Subscribe(events.EventRelayAlert)
		for {
			select {
			case event := <-relayCh:
				hub.BroadcastJSON("relay_alert", event.Payload)
			case <-ctx.Done():
				return
			}
		}
	}()

	// REST API
	handler := api.NewHandler(repo, priceCache, hub, logger)
	router := handler.SetupRouter(cfg.Server.CORSOrigin)

	// HTTP Server
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("HTTP server starting", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("HTTP server error", zap.Error(err))
		}
	}()

	logger.Info("✅ Cosmos ArbEngine Pro is running",
		zap.String("rest", fmt.Sprintf("http://localhost:%s", cfg.Server.Port)),
		zap.String("ws", fmt.Sprintf("ws://localhost:%s/ws/opportunities", cfg.Server.Port)),
	)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down...")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP server shutdown error", zap.Error(err))
	}

	feedManager.StopAll()
	logger.Info("goodbye 👋")
}

func findMigrationsDir() string {
	// Try relative paths
	candidates := []string{
		"migrations",
		"backend/migrations",
		"../migrations",
		"/app/migrations",
	}

	for _, dir := range candidates {
		abs, _ := filepath.Abs(dir)
		if _, err := os.Stat(abs); err == nil {
			return abs
		}
	}

	return "migrations"
}
