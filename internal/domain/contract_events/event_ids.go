package contract_events

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	TopicWalletDeployed = crypto.Keccak256Hash([]byte("WalletDeployed(address,address,bytes32,uint256)"))
	TopicKeyAdded       = crypto.Keccak256Hash([]byte("KeyAdded(uint256,uint8,uint8,address)"))
	TopicDIDCreated     = crypto.Keccak256Hash([]byte("DIDCreated(uint256,address,uint256)"))
	TopicDIDUpdated     = crypto.Keccak256Hash([]byte("DIDUpdated(uint256,uint256)"))
	TopicDIDDeactivated = crypto.Keccak256Hash([]byte("DIDDeactivated(uint256,uint256)"))
	TopicDIDReactivated = crypto.Keccak256Hash([]byte("DIDReactivated(uint256,uint256)"))
	TopicDIDTransferred = crypto.Keccak256Hash([]byte("DIDTransferred(uint256,address,address)"))
	TopicKeyRemoved     = crypto.Keccak256Hash([]byte("KeyRemoved(uint256,bytes32)"))
	TopicClaimAdded     = crypto.Keccak256Hash([]byte("ClaimAdded(uint256,bytes32,uint8,address,bytes)"))
	TopicClaimRemoved   = crypto.Keccak256Hash([]byte("ClaimRemoved(uint256,bytes32)"))
	TopicDataChanged    = crypto.Keccak256Hash([]byte("DataChanged(uint256,string,string)"))
	TopicDataDeleted    = crypto.Keccak256Hash([]byte("DataDeleted(uint256,string)"))
)

func IsTopicMatch(logTopics []common.Hash, topic common.Hash) bool {
	if len(logTopics) == 0 {
		return false
	}
	return logTopics[0] == topic
}
