package contract_events

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/types"
)

func DecodeWalletDeployed(contractABI *abi.ABI, log *types.Log) (*WalletDeployed, error) {
	if log == nil || !IsTopicMatch(log.Topics, TopicWalletDeployed) {
		return nil, nil
	}

	event := new(WalletDeployed)
	
	if len(log.Data) > 0 {
		unpacked, err := contractABI.Unpack("WalletDeployed", log.Data)
		if err == nil && len(unpacked) > 0 {
			if ts, ok := unpacked[0].(*big.Int); ok {
				event.Timestamp = ts
			}
		}
	}

	event.WalletAddress = decodeAddress(log, 1)
	event.Sender = decodeAddress(log, 2)
	if len(log.Topics) > 3 {
		copy(event.Salt[:], log.Topics[3].Bytes())
	}

	return event, nil
}
