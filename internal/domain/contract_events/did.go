package contract_events

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type KeyAdded struct {
	DIDIndex *big.Int
	KeyType  uint8
	Purpose  uint8
	Key      common.Address
}

type DIDCreated struct {
	DIDIndex  *big.Int
	Creator   common.Address
	Timestamp *big.Int
}

type DIDUpdated struct {
	DIDIndex  *big.Int
	Timestamp *big.Int
}

type DIDDeactivated struct {
	DIDIndex  *big.Int
	Timestamp *big.Int
}

type DIDReactivated struct {
	DIDIndex  *big.Int
	Timestamp *big.Int
}

type DIDTransferred struct {
	DIDIndex *big.Int
	OldOwner common.Address
	NewOwner common.Address
}

type KeyRemoved struct {
	DIDIndex *big.Int
	KeyData  [32]byte
}

type ClaimAdded struct {
	DIDIndex  *big.Int
	ClaimId   [32]byte
	Topic     uint8
	Issuer    common.Address
	ClaimData []byte
}

type ClaimRemoved struct {
	DIDIndex *big.Int
	ClaimId  [32]byte
}

type DataChanged struct {
	DIDIndex *big.Int
	Key      string
	Value    string
}

type DataDeleted struct {
	DIDIndex *big.Int
	Key      string
}
