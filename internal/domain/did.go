package domain

import (
	"math/big"
)

// DID Event Types
const (
	EventTypeCreateDID EventType = "CREATE_DID"
	EventTypeAddKey    EventType = "ADD_KEY"
	EventTypeAddClaim  EventType = "ADD_CLAIM"
	EventTypeStoreData EventType = "STORE_DATA"
)

// CreateDIDPayload represents the payload for registering a new DID.
type CreateDIDPayload struct {
	TargetAddress    string `json:"target_address"`
	DID              string `json:"did"`
	KeyIdentifier    string `json:"key_identifier,omitempty"`
	MultipleRPCCalls bool   `json:"multiple_rpc_calls,omitempty"`
}

// AddKeyPayload represents the payload for adding a verification key.
type AddKeyPayload struct {
	TargetAddress    string   `json:"target_address"`
	DIDIndex         *big.Int `json:"did_index"`
	KeyType          uint8    `json:"key_type"`
	PublicKey        string   `json:"public_key"` // will be hashed to [32]byte
	Purpose          uint8    `json:"purpose"`
	KeyIdentifier    string   `json:"key_identifier,omitempty"`
	MultipleRPCCalls bool     `json:"multiple_rpc_calls,omitempty"`
}

// AddClaimPayload represents the payload for attaching a claim.
type AddClaimPayload struct {
	TargetAddress    string   `json:"target_address"`
	DIDIndex         *big.Int `json:"did_index"`
	Topic            uint8    `json:"topic"`
	IssuerAddress    string   `json:"issuer_address"`
	Data             []byte   `json:"data"`
	URI              string   `json:"uri"`
	Signature        []byte   `json:"signature"`
	KeyIdentifier    string   `json:"key_identifier,omitempty"`
	MultipleRPCCalls bool     `json:"multiple_rpc_calls,omitempty"`
}

// StoreDataPayload represents the payload for storing arbitrary data.
type StoreDataPayload struct {
	TargetAddress    string   `json:"target_address"`
	DIDIndex         *big.Int `json:"did_index"`
	PropertyKey      string   `json:"property_key"`
	Data             string   `json:"data"` // string value as per SDK
	KeyIdentifier    string   `json:"key_identifier,omitempty"`
	MultipleRPCCalls bool     `json:"multiple_rpc_calls,omitempty"`
}
