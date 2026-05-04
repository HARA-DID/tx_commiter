package main

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/HARA-DID/did-queueing-engine/internal/callback"
	"github.com/HARA-DID/did-queueing-engine/internal/config"
	"github.com/HARA-DID/did-queueing-engine/internal/repository"
	"github.com/HARA-DID/did-queueing-engine/internal/sdk"
	"github.com/HARA-DID/did-queueing-engine/internal/service"
	"github.com/HARA-DID/did-queueing-engine/internal/worker"
	"github.com/HARA-DID/did-queueing-engine/pkg"
	"github.com/redis/go-redis/v9"
)

func initTxCheckService(cfg config.BlockchainConfig, jobRepo repository.JobRepository, callbackRegistry *callback.Registry, log *logrus.Logger) (*service.TxCheckService, error) {
	provider, err := sdk.NewProvider(cfg, cfg.PrivateKeys[0])
	if err != nil {
		return nil, err
	}
	if blockchainSvc, err := initBlockchainService(cfg, provider); err != nil {
		return nil, err
	} else {
		templateEventSvc := service.NewEventService(jobRepo, blockchainSvc, log)

		return service.NewTxCheckService(
			jobRepo,
			provider.Chain,
			callbackRegistry,
			templateEventSvc.EventCallbacks,
			log,
			cfg.TxCheckChannelBuffer,
		), nil
	}
}

type WorkerDependencies struct {
	Config           *config.Config
	JobRepo          repository.JobRepository
	RedisClient      *redis.Client
	CallbackRegistry *callback.Registry
	TxCheckSvc       *service.TxCheckService
	Metrics          *pkg.Metrics
	RetryCfg         pkg.RetryConfig
}

func startWorker(ctx context.Context, pk string, index int, deps *WorkerDependencies, log *logrus.Logger) error {
	consumerName := fmt.Sprintf("%s-%d", deps.Config.Worker.ConsumerName, index)
	workerLog := log.WithField("consumer", consumerName)

	provider, err := sdk.NewProvider(deps.Config.Blockchain, pk)
	if err != nil {
		return err
	}

	if blockchainSvc, err := initBlockchainService(deps.Config.Blockchain, provider); err != nil {
		return err
	} else {
		eventSvc := service.NewEventService(deps.JobRepo, blockchainSvc, log)
		eventSvc.SetTxCheckService(deps.TxCheckSvc)

		handler := worker.NewHandler(eventSvc, deps.RetryCfg, deps.Metrics, log)

		workerCfg := deps.Config.Worker
		workerCfg.ConsumerName = consumerName

		pool := worker.NewPool(deps.RedisClient, handler, deps.JobRepo, workerCfg, deps.Config.Redis, deps.Metrics, log)

		workerLog.Info("worker pool starting")
		pool.Run(ctx)
		return nil
	}
}

func initBlockchainService(bcConfig config.BlockchainConfig, provider *sdk.Provider) (service.BlockchainService, error) {
	didAdapter, err := sdk.NewDIDAdapter(provider, bcConfig)
	if err != nil {
		return nil, err
	}

	aaAdapter, err := sdk.NewAAAdapter(provider, bcConfig)
	if err != nil {
		return nil, err
	}

	vcAdapter, err := sdk.NewVCAdapter(provider, bcConfig)
	if err != nil {
		return nil, err
	}

	aliasAdapter, err := sdk.NewAliasAdapter(provider, bcConfig)
	if err != nil {
		return nil, err
	}

	BlockhainSvc := sdk.NewCompositeAdapter(didAdapter, aaAdapter, vcAdapter, aliasAdapter)
	return BlockhainSvc, nil
}
