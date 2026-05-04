package sdk

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	aapkg "github.com/HARA-DID/account-abstraction-sdk/pkg/entrypoint"
	"github.com/HARA-DID/account-abstraction-sdk/pkg/gasmanager"
	"github.com/HARA-DID/account-abstraction-sdk/pkg/walletfactory"
	"github.com/ethereum/go-ethereum/accounts/abi"
	harautils "github.com/HARA-DID/hara-core-blockchain-lib/utils"

	"github.com/HARA-DID/did-queueing-engine/internal/config"
	"github.com/HARA-DID/did-queueing-engine/internal/domain"
	"github.com/HARA-DID/did-queueing-engine/internal/domain/contract_events"
	"github.com/ethereum/go-ethereum/core/types"
)

// AAAdapter implements Account Abstraction related blockchain operations.
type AAAdapter struct {
	provider      *Provider
	entryPoint    *aapkg.EntryPoint
	gasManager    *gasmanager.GasManager
	walletFactory *walletfactory.WalletFactory
}

func NewAAAdapter(p *Provider, cfg config.BlockchainConfig) (*AAAdapter, error) {
	initCtx := context.Background()

	entryPoint, err := aapkg.NewEntryPointWithHNS(initCtx, cfg.EntryPointHNS, p.Chain)
	if err != nil {
		return nil, fmt.Errorf("resolve EntryPoint via HNS %q: %w", cfg.EntryPointHNS, err)
	}

	gasMgr, err := gasmanager.NewGasManagerWithHNS(initCtx, cfg.GasManagerHNS, p.Chain)
	if err != nil {
		return nil, fmt.Errorf("resolve GasManager via HNS %q: %w", cfg.GasManagerHNS, err)
	}

	walletFact, err := walletfactory.NewWalletFactoryWithHNS(initCtx, cfg.WalletFactoryHNS, p.Chain)
	if err != nil {
		return nil, fmt.Errorf("resolve WalletFactory via HNS %q: %w", cfg.WalletFactoryHNS, err)
	}

	return &AAAdapter{
		provider:      p,
		entryPoint:    entryPoint,
		gasManager:    gasMgr,
		walletFactory: walletFact,
	}, nil
}

func (a *AAAdapter) GetEntryPointAddress() harautils.Address {
	return a.entryPoint.Address
}

func (a *AAAdapter) GetWalletFactoryAddress() harautils.Address {
	return a.walletFactory.Address
}

func (a *AAAdapter) DecodeWalletDeployed(log *types.Log) (*contract_events.WalletDeployed, error) {
	return contract_events.DecodeWalletDeployed(&a.walletFactory.ContractABI, log)
}

func (a *AAAdapter) GetWalletFactoryABI() *abi.ABI {
	return &a.walletFactory.ContractABI
}

// ── BlockchainService implementation for AA ──────────────────────

func (a *AAAdapter) HandleOps(ctx context.Context, p domain.HandleOpsPayload) (*domain.BlockchainResult, error) {
	sender, err := a.provider.Wallet.GetAddress()
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet address: %w", err)
	}

	var nonce *big.Int
	if p.UserNonce != "" {
		n, ok := new(big.Int).SetString(p.UserNonce, 0)
		if !ok {
			return nil, fmt.Errorf("invalid user_nonce format: %s", p.UserNonce)
		}
		nonce = n
	} else {
		n, err := a.entryPoint.GetNonce(ctx, sender, big.NewInt(0))
		if err != nil {
			return nil, fmt.Errorf("failed to get nonce from EntryPoint: %w", err)
		}
		nonce = n
	}

	val := big.NewInt(0)
	if p.Value != "" {
		if v, ok := new(big.Int).SetString(p.Value, 0); ok {
			val = v
		}
	}

	blockNum := big.NewInt(0)
	if p.ClientBlockNumber != "" {
		if bn, ok := new(big.Int).SetString(p.ClientBlockNumber, 0); ok {
			blockNum = bn
		}
	}

	userOp := aapkg.UserOp{
		Target:            harautils.HexToAddress(p.Target),
		Value:             val,
		Data:              p.Data,
		ClientBlockNumber: blockNum,
		UserNonce:         nonce,
		Signature:         p.Signature,
	}

	senderAddr := harautils.HexToAddress(p.Sender)
	
	if p.Sender == "" {
		// TODO: Fallback to relayer for backward compatibility (not recommended for production AA)
		addr, _ := a.provider.Wallet.GetAddress()
		senderAddr = addr
	}

	params := aapkg.HandleOpsParams{
		Wallet: senderAddr,
		UserOp: userOp,
	}

	txHashes, err := a.entryPoint.HandleOps(ctx, a.provider.Wallet, params, p.MultipleRPCCalls)
	if err != nil {
		return nil, fmt.Errorf("entryPoint.HandleOps: %w", err)
	}
	return &domain.BlockchainResult{TxHashes: txHashes}, nil
}

func (a *AAAdapter) DeployWallet(ctx context.Context, p domain.DeployWalletPayload) (*domain.BlockchainResult, error) {
	saltBytes, err := hex.DecodeString(strings.TrimPrefix(p.Salt, "0x"))
	if err != nil {
		return nil, fmt.Errorf("invalid salt hex: %w", err)
	}
	var salt [32]byte
	copy(salt[:], saltBytes)

	params := walletfactory.DeployWalletParams{
		Owner: harautils.HexToAddress(p.Owner),
		Salt:  salt,
	}

	txHashes, err := a.walletFactory.DeployWallet(ctx, a.provider.Wallet, params, p.MultipleRPCCalls)
	if err != nil {
		return nil, fmt.Errorf("walletFactory.DeployWallet: %w", err)
	}
	return &domain.BlockchainResult{TxHashes: txHashes}, nil
}
