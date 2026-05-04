package contract_events

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type OrgCreated struct {
	OrgDIDCounter *big.Int
	OrgDIDHash    common.Hash
	Creator       common.Address
}

type OrgDeactivated struct {
	OrgDIDHash common.Hash
}

type OrgReactivated struct {
	OrgDIDHash common.Hash
}

type OrgOwnershipTransferred struct {
	OrgDIDHash common.Hash
	OldOwner   common.Address
	NewOwner   common.Address
}

type MemberAdded struct {
	OrgDIDIndex *big.Int
	UserDIDHash common.Hash
	Role        uint8
	Timestamp   *big.Int
}

type MemberRemoved struct {
	OrgDIDIndex *big.Int
	UserDIDHash common.Hash
	Timestamp   *big.Int
}

type MemberUpdated struct {
	OrgDIDIndex *big.Int
	UserDIDHash common.Hash
	Role        uint8
	Timestamp   *big.Int
}
