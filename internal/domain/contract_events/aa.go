package contract_events

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type WalletDeployed struct {
	WalletAddress common.Address
	Sender        common.Address
	Salt          [32]byte
	Timestamp     *big.Int
}