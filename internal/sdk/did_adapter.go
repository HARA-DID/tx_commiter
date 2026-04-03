package sdk

import (
	"context"
	"fmt"

	didfactory "github.com/HARA-DID/did-root-sdk/pkg/factory"
	haracontract "github.com/meQlause/hara-core-blockchain-lib/pkg/contract"
	harautils "github.com/meQlause/hara-core-blockchain-lib/utils"

	"github.com/myorg/worker-service/internal/config"
	"github.com/myorg/worker-service/internal/domain"
)

// DIDAdapter implements DID-related blockchain operations.
type DIDAdapter struct {
	provider *Provider
	factory  *didfactory.Factory
}

// NewDIDAdapter initializes the DID SDK factory with common resources.
func NewDIDAdapter(p *Provider, cfg config.BlockchainConfig) (*DIDAdapter, error) {
	initCtx := context.Background()
	var contract *haracontract.Contract
	var err error

	// ── Contract Resolution ──────────────────────────────────────────
	contract, err = p.Chain.ContractWithHNS(initCtx, cfg.DIDRootFactoryHNS)
	if err != nil {
		return nil, fmt.Errorf("resolve contract via HNS %q: %w", cfg.DIDRootFactoryHNS, err)
	}

	// ── Factory ─────────────────────────────────────────────────────
	factory := didfactory.NewFactory(
		p.WalletAddr,
		harautils.ABI{}, 
		p.Chain,
		contract,
	)

	return &DIDAdapter{
		provider: p,
		factory:  factory,
	}, nil
}

// ── Encode Methods ───────────────────────────────────────────────

func (a *DIDAdapter) EncodeCreateDID(p domain.CreateDIDPayload) ([]byte, error) {
	keyID, err := resolveKeyIdentifier(p.KeyIdentifier)
	if err != nil {
		return nil, err
	}

	argBuilder := a.provider.Network.ArgBuilder().
		Type("string").Value(p.DID)
	data := harautils.EncodeArgs(argBuilder)

	return a.encodeDID(didfactory.TypeCreateDID, data, keyID)
}

func (a *DIDAdapter) EncodeAddKey(p domain.AddKeyPayload) ([]byte, error) {
	keyID, err := resolveKeyIdentifier(p.KeyIdentifier)
	if err != nil {
		return nil, err
	}

	// The SDK hashes the public key to [32]byte
	keyHashed := harautils.HexToHash(p.PublicKey)

	argBuilder := a.provider.Network.ArgBuilder().
		Type("uint256").Value(p.DIDIndex).
		Type("bytes32").Value(keyHashed).
		Type("string").Value(""). // KeyIdentifierDst usually empty in these calls
		Type("uint8").Value(p.Purpose).
		Type("uint8").Value(p.KeyType)
	data := harautils.EncodeArgs(argBuilder)

	return a.encodeDID(didfactory.TypeAddKey, data, keyID)
}

func (a *DIDAdapter) EncodeAddClaim(p domain.AddClaimPayload) ([]byte, error) {
	keyID, err := resolveKeyIdentifier(p.KeyIdentifier)
	if err != nil {
		return nil, err
	}

	argBuilder := a.provider.Network.ArgBuilder().
		Type("uint256").Value(p.DIDIndex).
		Type("uint8").Value(p.Topic).
		Type("bytes").Value(p.Data).
		Type("string").Value(p.URI).
		Type("bytes").Value(p.Signature)
	data := harautils.EncodeArgs(argBuilder)

	return a.encodeDID(didfactory.TypeAddClaim, data, keyID)
}

func (a *DIDAdapter) EncodeStoreData(p domain.StoreDataPayload) ([]byte, error) {
	keyID, err := resolveKeyIdentifier(p.KeyIdentifier)
	if err != nil {
		return nil, err
	}

	argBuilder := a.provider.Network.ArgBuilder().
		Type("uint256").Value(p.DIDIndex).
		Type("string").Value(p.PropertyKey).
		Type("string").Value(p.Data)
	data := harautils.EncodeArgs(argBuilder)

	return a.encodeDID(didfactory.TypeStoreData, data, keyID)
}

// ── Private Helpers ──────────────────────────────────────────────

func (a *DIDAdapter) encodeDID(txType uint8, data []byte, keyIdentifier string) ([]byte, error) {
	method, ok := a.factory.ContractABI.Methods["callExternalDID"]
	if !ok {
		return nil, fmt.Errorf("method callExternalDID not found in ABI")
	}

	inputs, err := method.Inputs.Pack(txType, data, keyIdentifier)
	if err != nil {
		return nil, fmt.Errorf("failed to pack callExternalDID arguments: %w", err)
	}

	return append(method.ID, inputs...), nil
}

// ---------------------------------------------------------------------------
// DID Helpers
// ---------------------------------------------------------------------------

func resolveKeyIdentifier(provided string) (string, error) {
	if provided != "" {
		return provided, nil
	}
	keyID, err := didfactory.GenerateKeyIdentifier()
	if err != nil {
		return "", fmt.Errorf("generate key identifier: %w", err)
	}
	return keyID, nil
}
