package domain

import (
	"math/big"
)

// Verifiable Credentials Event Types
const (
	EventTypeIssueCredential     EventType = "ISSUE_CREDENTIAL"
	EventTypeBurnCredential      EventType = "BURN_CREDENTIAL"
	EventTypeUpdateMetadata      EventType = "UPDATE_METADATA"
	EventTypeRevokeCredential    EventType = "REVOKE_CREDENTIAL"
	EventTypeApproveCredentialOrg EventType = "APPROVE_CREDENTIAL_ORG"
	EventTypeApproveCredential    EventType = "APPROVE_CREDENTIAL"
	EventTypeSetDidRootStorage    EventType = "SET_DID_ROOT_STORAGE"
	EventTypeSetDidOrgStorage     EventType = "SET_DID_ORG_STORAGE"
)

// VC Payloads

type IssueCredentialPayload struct {
	TargetAddress    string   `json:"target_address"`
	Option           uint8    `json:"option"`
	DIDRecipient     [32]byte `json:"did_recipient"`
	Issuer           [32]byte `json:"issuer"`
	ExpiredAt        *big.Int `json:"expired_at"`
	OffchainHash     [32]byte `json:"offchain_hash"`
	MerkleTreeRoot   [32]byte `json:"merkle_tree_root"`
	PublicIdentity   [32]byte `json:"public_identity"`
	MultipleRPCCalls bool     `json:"multiple_rpc_calls,omitempty"`
}

type BurnCredentialPayload struct {
	TargetAddress    string   `json:"target_address"`
	Option           uint8    `json:"option"`
	DID              [32]byte `json:"did"`
	TokenID          *big.Int `json:"token_id"`
	MultipleRPCCalls bool     `json:"multiple_rpc_calls,omitempty"`
}

type UpdateMetadataPayload struct {
	TargetAddress    string   `json:"target_address"`
	Option           uint8    `json:"option"`
	TokenID          *big.Int `json:"token_id"`
	ExpiredAt        *big.Int `json:"expired_at"`
	OffchainHash     [32]byte `json:"offchain_hash"`
	MultipleRPCCalls bool     `json:"multiple_rpc_calls,omitempty"`
}

type RevokeCredentialPayload struct {
	TargetAddress    string   `json:"target_address"`
	Option           uint8    `json:"option"`
	TokenID          *big.Int `json:"token_id"`
	MultipleRPCCalls bool     `json:"multiple_rpc_calls,omitempty"`
}

type ApproveCredentialOrgPayload struct {
	TargetAddress    string   `json:"target_address"`
	Option           uint8    `json:"option"`
	TokenID          *big.Int `json:"token_id"`
	OrgDIDHash       [32]byte `json:"org_did_hash"`
	UserDIDHash      [32]byte `json:"user_did_hash"`
	Signature        []byte   `json:"signature"`
	MultipleRPCCalls bool     `json:"multiple_rpc_calls,omitempty"`
}

type ApproveCredentialPayload struct {
	TargetAddress    string   `json:"target_address"`
	Option           uint8    `json:"option"`
	TokenID          *big.Int `json:"token_id"`
	Signature        []byte   `json:"signature"`
	MultipleRPCCalls bool     `json:"multiple_rpc_calls,omitempty"`
}

type SetAddressPayload struct {
	TargetAddress    string `json:"target_address"`
	Address          string `json:"address"`
	MultipleRPCCalls bool   `json:"multiple_rpc_calls,omitempty"`
}

// Result types for queries (if needed by worker)
type TokenIdsResult struct {
	TokenIds []*big.Int
	Total    *big.Int
}
