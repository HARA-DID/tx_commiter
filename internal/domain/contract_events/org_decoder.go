package contract_events

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func DecodeOrgCreated(contractABI *abi.ABI, log *types.Log) (*OrgCreated, error) {
	if log == nil || !IsTopicMatch(log.Topics, TopicOrgCreated) {
		return nil, nil
	}

	event := new(OrgCreated)
	if len(log.Topics) >= 4 {
		event.OrgDIDCounter = decodeIndexed(log, 1)
		event.OrgDIDHash = common.BytesToHash(log.Topics[2].Bytes())
		event.Creator = decodeAddress(log, 3)
		return event, nil
	}

	// Fallback to data
	unpacked, err := contractABI.Unpack("OrgCreated", log.Data)
	if err == nil {
		idx := 1
		if event.OrgDIDCounter == nil && len(unpacked) >= idx {
			event.OrgDIDCounter = unpacked[idx-1].(*big.Int)
			idx++
		}
		if event.OrgDIDHash == (common.Hash{}) && len(unpacked) >= idx {
			event.OrgDIDHash = unpacked[idx-1].([32]byte)
			idx++
		}
		if event.Creator == (common.Address{}) && len(unpacked) >= idx {
			event.Creator = unpacked[idx-1].(common.Address)
		}
	}

	if event.OrgDIDCounter == nil {
		event.OrgDIDCounter = decodeIndexed(log, 1)
	}
	return event, nil
}

func DecodeOrgLifecycle(contractABI *abi.ABI, log *types.Log, topic common.Hash, name string) (any, error) {
	if log == nil || !IsTopicMatch(log.Topics, topic) {
		return nil, nil
	}

	var orgDIDHash common.Hash
	if len(log.Topics) >= 2 {
		orgDIDHash = common.BytesToHash(log.Topics[1].Bytes())
	} else {
		unpacked, err := contractABI.Unpack(name, log.Data)
		if err == nil && len(unpacked) >= 1 {
			orgDIDHash = unpacked[0].([32]byte)
		}
	}

	if name == "OrgDeactivated" {
		return &OrgDeactivated{OrgDIDHash: orgDIDHash}, nil
	}
	return &OrgReactivated{OrgDIDHash: orgDIDHash}, nil
}

func DecodeOrgOwnershipTransferred(contractABI *abi.ABI, log *types.Log) (*OrgOwnershipTransferred, error) {
	if log == nil || !IsTopicMatch(log.Topics, TopicOrgOwnershipTransferred) {
		return nil, nil
	}

	event := new(OrgOwnershipTransferred)
	if len(log.Topics) >= 4 {
		event.OrgDIDHash = common.BytesToHash(log.Topics[1].Bytes())
		event.OldOwner = decodeAddress(log, 2)
		event.NewOwner = decodeAddress(log, 3)
		return event, nil
	}

	unpacked, err := contractABI.Unpack("OrgOwnershipTransferred", log.Data)
	if err == nil {
		if len(unpacked) >= 3 {
			event.OrgDIDHash = unpacked[0].([32]byte)
			event.OldOwner = unpacked[1].(common.Address)
			event.NewOwner = unpacked[2].(common.Address)
		}
	}

	return event, nil
}

func DecodeMemberEvent(contractABI *abi.ABI, log *types.Log, topic common.Hash, name string) (any, error) {
	if log == nil || !IsTopicMatch(log.Topics, topic) {
		return nil, nil
	}

	orgDIDIndex := decodeIndexed(log, 1)
	var userDIDHash common.Hash
	if len(log.Topics) >= 3 {
		userDIDHash = common.BytesToHash(log.Topics[2].Bytes())
	}

	unpacked, err := contractABI.Unpack(name, log.Data)
	if err != nil {
		return nil, err
	}

	switch name {
	case "MemberAdded":
		event := &MemberAdded{OrgDIDIndex: orgDIDIndex, UserDIDHash: userDIDHash}
		if len(unpacked) >= 2 {
			event.Role = unpacked[0].(uint8)
			event.Timestamp = unpacked[1].(*big.Int)
		}
		return event, nil
	case "MemberRemoved":
		event := &MemberRemoved{OrgDIDIndex: orgDIDIndex, UserDIDHash: userDIDHash}
		if len(unpacked) >= 1 {
			event.Timestamp = unpacked[0].(*big.Int)
		}
		return event, nil
	case "MemberUpdated":
		event := &MemberUpdated{OrgDIDIndex: orgDIDIndex, UserDIDHash: userDIDHash}
		if len(unpacked) >= 2 {
			event.Role = unpacked[0].(uint8)
			event.Timestamp = unpacked[1].(*big.Int)
		}
		return event, nil
	}

	return nil, fmt.Errorf("unknown member event: %s", name)
}
