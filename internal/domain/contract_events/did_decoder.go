package contract_events

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func decodeIndexed(log *types.Log, index int) *big.Int {
	if index >= len(log.Topics) {
		return big.NewInt(0)
	}
	return new(big.Int).SetBytes(log.Topics[index].Bytes())
}

func decodeAddress(log *types.Log, index int) common.Address {
	if index >= len(log.Topics) {
		return common.Address{}
	}
	return common.BytesToAddress(log.Topics[index].Bytes())
}

func DecodeKeyAdded(contractABI *abi.ABI, log *types.Log) (*KeyAdded, error) {
	if log == nil || !IsTopicMatch(log.Topics, TopicKeyAdded) {
		return nil, nil
	}

	event := new(KeyAdded)
	if len(log.Topics) < 4 {
		return nil, fmt.Errorf("insufficient topics for KeyAdded")
	}

	event.DIDIndex = decodeIndexed(log, 1)
	event.KeyType = uint8(decodeIndexed(log, 2).Uint64())
	event.Purpose = uint8(decodeIndexed(log, 3).Uint64())

	if len(log.Data) > 0 {
		unpacked, err := contractABI.Unpack("KeyAdded", log.Data)
		if err == nil && len(unpacked) > 0 {
			if addr, ok := unpacked[0].(common.Address); ok {
				event.Key = addr
			}
		}
	}

	if event.Key == (common.Address{}) && len(log.Topics) > 4 {
		event.Key = decodeAddress(log, 4)
	}

	return event, nil
}

func DecodeDIDCreated(contractABI *abi.ABI, log *types.Log) (*DIDCreated, error) {
	if log == nil || !IsTopicMatch(log.Topics, TopicDIDCreated) {
		return nil, nil
	}

	event := new(DIDCreated)
	if len(log.Topics) < 3 {
		return nil, fmt.Errorf("insufficient topics for DIDCreated")
	}

	event.DIDIndex = decodeIndexed(log, 1)
	event.Creator = decodeAddress(log, 2)

	if len(log.Data) > 0 {
		unpacked, err := contractABI.Unpack("DIDCreated", log.Data)
		if err == nil && len(unpacked) > 0 {
			if ts, ok := unpacked[0].(*big.Int); ok {
				event.Timestamp = ts
			}
		}
	}

	return event, nil
}

func DecodeDIDUpdated(contractABI *abi.ABI, log *types.Log) (*DIDUpdated, error) {
	if log == nil || !IsTopicMatch(log.Topics, TopicDIDUpdated) {
		return nil, nil
	}

	event := new(DIDUpdated)
	event.DIDIndex = decodeIndexed(log, 1)

	unpacked, err := contractABI.Unpack("DIDUpdated", log.Data)
	if err == nil && len(unpacked) > 0 {
		if ts, ok := unpacked[0].(*big.Int); ok {
			event.Timestamp = ts
		}
	}

	return event, nil
}

func DecodeDIDLifecycle(contractABI *abi.ABI, log *types.Log, topic common.Hash, name string) (any, error) {
	if log == nil || !IsTopicMatch(log.Topics, topic) {
		return nil, nil
	}

	didIndex := decodeIndexed(log, 1)
	var timestamp *big.Int
	unpacked, err := contractABI.Unpack(name, log.Data)
	if err == nil && len(unpacked) > 0 {
		if ts, ok := unpacked[0].(*big.Int); ok {
			timestamp = ts
		}
	}

	if name == "DIDDeactivated" {
		return &DIDDeactivated{DIDIndex: didIndex, Timestamp: timestamp}, nil
	}
	return &DIDReactivated{DIDIndex: didIndex, Timestamp: timestamp}, nil
}

func DecodeDIDTransferred(contractABI *abi.ABI, log *types.Log) (*DIDTransferred, error) {
	if log == nil || !IsTopicMatch(log.Topics, TopicDIDTransferred) {
		return nil, nil
	}

	event := new(DIDTransferred)
	event.DIDIndex = decodeIndexed(log, 1)
	event.OldOwner = decodeAddress(log, 2)
	event.NewOwner = decodeAddress(log, 3)

	return event, nil
}

func DecodeKeyRemoved(contractABI *abi.ABI, log *types.Log) (*KeyRemoved, error) {
	if log == nil || !IsTopicMatch(log.Topics, TopicKeyRemoved) {
		return nil, nil
	}

	event := new(KeyRemoved)
	event.DIDIndex = decodeIndexed(log, 1)
	if len(log.Topics) > 2 {
		copy(event.KeyData[:], log.Topics[2].Bytes())
	}

	return event, nil
}

func DecodeClaimAdded(contractABI *abi.ABI, log *types.Log) (*ClaimAdded, error) {
	if log == nil || !IsTopicMatch(log.Topics, TopicClaimAdded) {
		return nil, nil
	}

	event := new(ClaimAdded)
	event.DIDIndex = decodeIndexed(log, 1)
	if len(log.Topics) > 2 {
		copy(event.ClaimId[:], log.Topics[2].Bytes())
	}
	event.Topic = uint8(decodeIndexed(log, 3).Uint64())

	unpacked, err := contractABI.Unpack("ClaimAdded", log.Data)
	if err == nil && len(unpacked) >= 2 {
		if addr, ok := unpacked[0].(common.Address); ok {
			event.Issuer = addr
		}
		if data, ok := unpacked[1].([]byte); ok {
			event.ClaimData = data
		}
	}

	return event, nil
}

func DecodeClaimRemoved(contractABI *abi.ABI, log *types.Log) (*ClaimRemoved, error) {
	if log == nil || !IsTopicMatch(log.Topics, TopicClaimRemoved) {
		return nil, nil
	}

	event := new(ClaimRemoved)
	event.DIDIndex = decodeIndexed(log, 1)
	if len(log.Topics) > 2 {
		copy(event.ClaimId[:], log.Topics[2].Bytes())
	}

	return event, nil
}

func DecodeDataChanged(contractABI *abi.ABI, log *types.Log) (*DataChanged, error) {
	if log == nil || !IsTopicMatch(log.Topics, TopicDataChanged) {
		return nil, nil
	}

	event := new(DataChanged)
	event.DIDIndex = decodeIndexed(log, 1)

	unpacked, err := contractABI.Unpack("DataChanged", log.Data)
	if err == nil && len(unpacked) >= 2 {
		if k, ok := unpacked[0].(string); ok {
			event.Key = k
		}
		if v, ok := unpacked[1].(string); ok {
			event.Value = v
		}
	}

	return event, nil
}

func DecodeDataDeleted(contractABI *abi.ABI, log *types.Log) (*DataDeleted, error) {
	if log == nil || !IsTopicMatch(log.Topics, TopicDataDeleted) {
		return nil, nil
	}

	event := new(DataDeleted)
	event.DIDIndex = decodeIndexed(log, 1)

	unpacked, err := contractABI.Unpack("DataDeleted", log.Data)
	if err == nil && len(unpacked) >= 1 {
		if k, ok := unpacked[0].(string); ok {
			event.Key = k
		}
	}

	return event, nil
}
