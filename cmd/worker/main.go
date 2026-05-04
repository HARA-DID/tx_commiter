package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"

	infradb "github.com/HARA-DID/did-queueing-engine/internal/infra/db"
	redisinfra "github.com/HARA-DID/did-queueing-engine/internal/infra/redis"

	"github.com/HARA-DID/did-queueing-engine/internal/callback"
	"github.com/HARA-DID/did-queueing-engine/internal/config"
	"github.com/HARA-DID/did-queueing-engine/internal/worker"
	"github.com/HARA-DID/did-queueing-engine/pkg"
)

func main() {
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.InfoLevel)

	_ = godotenv.Load()

	// ── Config ─────────────────────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		log.WithError(err).Fatal("failed to load config")
	}

	// ── Database ───────────────────────────────────────────────────────────
	db, err := infradb.Connect(cfg.DB)
	if err != nil {
		log.WithError(err).Fatal("failed to connect to database")
	}
	defer db.Close()

	// ── Redis ──────────────────────────────────────────────────────────────
	redisClient, err := redisinfra.NewClient(cfg.Redis)
	if err != nil {
		log.WithError(err).Fatal("failed to connect to redis")
	}
	defer redisClient.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := redisinfra.EnsureConsumerGroup(ctx, redisClient, cfg.Redis.StreamName, cfg.Redis.GroupName); err != nil {
		log.WithError(err).Fatal("failed to ensure consumer group")
	}
	log.WithFields(logrus.Fields{
		"stream": cfg.Redis.StreamName,
		"group":  cfg.Redis.GroupName,
	}).Info("consumer group ready")

	// ── HTTP server (health + metrics) ─────────────────────────────────────
	httpSrv := worker.NewHTTPServer(cfg.Server.Port, log)
	httpSrv.Start()

	// ── Metrics, Repositories, Callbacks ───────────────────────────────────
	jobRepo := infradb.NewPostgresJobRepository(db)
	callbackRegistry := callback.NewDefaultRegistry()
	metrics := pkg.NewMetrics(prometheus.DefaultRegisterer)
	retryCfg := pkg.DefaultRetryConfig(cfg.Worker.MaxRetry, cfg.Worker.RetryBaseDelay)

	// ── Global Transaction Check Service ──────────────────────────────────
	txCheckSvc, err := initTxCheckService(cfg.Blockchain, jobRepo, callbackRegistry, log)
	if err != nil {
		log.WithError(err).Fatal("failed to initialize transaction check service")
	}

	// ── Worker pool per private key ────────────────────────────────────────
	log.WithField("worker_count", len(cfg.Blockchain.PrivateKeys)).Info("starting worker pools per identity")
	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		txCheckSvc.Start(gCtx)
		return nil
	})
	deps := &WorkerDependencies{
		Config:           cfg,
		JobRepo:          jobRepo,
		RedisClient:      redisClient,
		CallbackRegistry: callbackRegistry,
		TxCheckSvc:       txCheckSvc,
		Metrics:          metrics,
		RetryCfg:         retryCfg,
	}

	for i, pk := range cfg.Blockchain.PrivateKeys {
		pkCaptured := pk
		workerIndex := i + 1
		g.Go(func() error {
			return startWorker(gCtx, pkCaptured, workerIndex, deps, log)
		})
	}

	_ = g.Wait()

	// ── Graceful shutdown ──────────────────────────────────────────────────
	log.Info("shutting down HTTP server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Worker.ShutdownTimeout)
	defer cancel()
	httpSrv.Shutdown(shutdownCtx)

	log.Info("worker service stopped cleanly")
}
