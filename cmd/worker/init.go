package main

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/HARA-DID/did-queueing-engine/internal/callback"
	"github.com/HARA-DID/did-queueing-engine/internal/config"
	"github.com/HARA-DID/did-queueing-engine/internal/domain/contract_events"
	"github.com/HARA-DID/did-queueing-engine/internal/repository"
	"github.com/HARA-DID/did-queueing-engine/internal/sdk"
	"github.com/HARA-DID/did-queueing-engine/internal/service"
	"github.com/HARA-DID/did-queueing-engine/internal/worker"
	"github.com/HARA-DID/did-queueing-engine/pkg"
	"github.com/ethereum/go-ethereum/core/types"
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

		svc := service.NewTxCheckService(
			jobRepo,
			provider.Chain,
			provider.Network,
			callbackRegistry,
			templateEventSvc.EventCallbacks,
			log,
			cfg.TxCheckChannelBuffer,
		)

		// Register AA decoders
		walletFactoryABI := blockchainSvc.GetWalletFactoryABI()
		svc.RegisterDecoder(contract_events.TopicWalletDeployed, func(l *types.Log) (any, error) {
			return contract_events.DecodeWalletDeployed(walletFactoryABI, l)
		})

		// Register DID decoders
		didFactoryABI := blockchainSvc.GetDIDFactoryABI()
		svc.RegisterDecoder(contract_events.TopicKeyAdded, func(l *types.Log) (any, error) {
			return contract_events.DecodeKeyAdded(didFactoryABI, l)
		})
		svc.RegisterDecoder(contract_events.TopicDIDCreated, func(l *types.Log) (any, error) {
			return contract_events.DecodeDIDCreated(didFactoryABI, l)
		})
		svc.RegisterDecoder(contract_events.TopicDIDUpdated, func(l *types.Log) (any, error) {
			return contract_events.DecodeDIDUpdated(didFactoryABI, l)
		})
		svc.RegisterDecoder(contract_events.TopicDIDDeactivated, func(l *types.Log) (any, error) {
			return contract_events.DecodeDIDLifecycle(didFactoryABI, l, contract_events.TopicDIDDeactivated, "DIDDeactivated")
		})
		svc.RegisterDecoder(contract_events.TopicDIDReactivated, func(l *types.Log) (any, error) {
			return contract_events.DecodeDIDLifecycle(didFactoryABI, l, contract_events.TopicDIDReactivated, "DIDReactivated")
		})
		svc.RegisterDecoder(contract_events.TopicDIDTransferred, func(l *types.Log) (any, error) {
			return contract_events.DecodeDIDTransferred(didFactoryABI, l)
		})
		svc.RegisterDecoder(contract_events.TopicKeyRemoved, func(l *types.Log) (any, error) {
			return contract_events.DecodeKeyRemoved(didFactoryABI, l)
		})
		svc.RegisterDecoder(contract_events.TopicClaimAdded, func(l *types.Log) (any, error) {
			return contract_events.DecodeClaimAdded(didFactoryABI, l)
		})
		svc.RegisterDecoder(contract_events.TopicClaimRemoved, func(l *types.Log) (any, error) {
			return contract_events.DecodeClaimRemoved(didFactoryABI, l)
		})
		svc.RegisterDecoder(contract_events.TopicDataChanged, func(l *types.Log) (any, error) {
			return contract_events.DecodeDataChanged(didFactoryABI, l)
		})
		svc.RegisterDecoder(contract_events.TopicDataDeleted, func(l *types.Log) (any, error) {
			return contract_events.DecodeDataDeleted(didFactoryABI, l)
		})

		return svc, nil
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
