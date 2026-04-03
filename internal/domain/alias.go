package domain

// Alias Event Types
const (
	EventTypeRegisterTLD          EventType = "REGISTER_TLD"
	EventTypeRegisterDomain       EventType = "REGISTER_DOMAIN"
	EventTypeSetDIDAlias         EventType = "SET_DID_ALIAS"
	EventTypeSetDIDOrgAlias      EventType = "SET_DID_ORG_ALIAS"
	EventTypeExtendRegistration   EventType = "EXTEND_REGISTRATION"
	EventTypeRevokeAlias          EventType = "REVOKE_ALIAS"
	EventTypeUnrevokeAlias        EventType = "UNREVOKE_ALIAS"
	EventTypeRegisterSubdomain    EventType = "REGISTER_SUBDOMAIN"
	EventTypeTransferAliasOwnership EventType = "TRANSFER_ALIAS_OWNERSHIP"
	EventTypeTransferTLD          EventType = "TRANSFER_TLD"
	EventTypeSetAliasRootStorage  EventType = "SET_ALIAS_ROOT_STORAGE"
	EventTypeSetAliasOrgStorage   EventType = "SET_ALIAS_ORG_STORAGE"
	EventTypeSetFactoryContract    EventType = "SET_FACTORY_CONTRACT"
)

// Alias Payloads

type RegisterTLDPayload struct {
	TargetAddress    string `json:"target_address"`
	TLD              string `json:"tld"`
	Owner            string `json:"owner"`
	MultipleRPCCalls bool   `json:"multiple_rpc_calls,omitempty"`
}

type RegisterDomainPayload struct {
	TargetAddress    string `json:"target_address"`
	Label            string `json:"label"`
	TLD              string `json:"tld"`
	Period           uint8  `json:"period"` // 0: 1yr, 1: 2yrs, 2: 3yrs
	MultipleRPCCalls bool   `json:"multiple_rpc_calls,omitempty"`
}

type SetDIDAliasPayload struct {
	TargetAddress    string `json:"target_address"`
	Name             string `json:"name"`
	DID              string `json:"did"` // hex string for [32]byte
	MultipleRPCCalls bool   `json:"multiple_rpc_calls,omitempty"`
}

type SetDIDOrgAliasPayload struct {
	TargetAddress    string `json:"target_address"`
	Name             string `json:"name"`
	OrgDIDHash       string `json:"org_did_hash"`  // hex string for [32]byte
	UserDIDHash      string `json:"user_did_hash"` // hex string for [32]byte
	MultipleRPCCalls bool   `json:"multiple_rpc_calls,omitempty"`
}

type ExtendRegistrationPayload struct {
	TargetAddress    string `json:"target_address"`
	Node             string `json:"node"` // hex string for [32]byte
	Period           uint8  `json:"period"`
	MultipleRPCCalls bool   `json:"multiple_rpc_calls,omitempty"`
}

type RevokeAliasPayload struct {
	TargetAddress    string `json:"target_address"`
	Node             string `json:"node"` // hex string for [32]byte
	MultipleRPCCalls bool   `json:"multiple_rpc_calls,omitempty"`
}

type UnrevokeAliasPayload struct {
	TargetAddress    string `json:"target_address"`
	Node             string `json:"node"` // hex string for [32]byte
	MultipleRPCCalls bool   `json:"multiple_rpc_calls,omitempty"`
}

type RegisterSubdomainPayload struct {
	TargetAddress    string `json:"target_address"`
	Label            string `json:"label"`
	ParentDomain     string `json:"parent_domain"`
	Period           uint8  `json:"period"`
	MultipleRPCCalls bool   `json:"multiple_rpc_calls,omitempty"`
}

type TransferAliasOwnershipPayload struct {
	TargetAddress    string `json:"target_address"`
	Node             string `json:"node"` // hex string for [32]byte
	NewOwner         string `json:"new_owner"`
	MultipleRPCCalls bool   `json:"multiple_rpc_calls,omitempty"`
}

type TransferTLDPayload struct {
	TargetAddress    string `json:"target_address"`
	TLD              string `json:"tld"`
	NewOwner         string `json:"new_owner"`
	MultipleRPCCalls bool   `json:"multiple_rpc_calls,omitempty"`
}

type SetAliasAddressPayload struct {
	TargetAddress    string `json:"target_address"`
	Address          string `json:"address"`
	MultipleRPCCalls bool   `json:"multiple_rpc_calls,omitempty"`
}

type SetFactoryContractPayload struct {
	TargetAddress    string `json:"target_address"`
	FactoryContract  string `json:"factory_contract"`
	MultipleRPCCalls bool   `json:"multiple_rpc_calls,omitempty"`
}
