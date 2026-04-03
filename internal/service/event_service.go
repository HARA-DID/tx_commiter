package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/HARA-DID/did-queueing-engine/internal/domain"
	"github.com/HARA-DID/did-queueing-engine/internal/repository"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// EventService orchestrates event processing: idempotency → DB → blockchain → DB.
type EventService struct {
	jobRepo    repository.JobRepository
	blockchain BlockchainService
	log        *logrus.Logger
}

func NewEventService(
	jobRepo repository.JobRepository,
	blockchain BlockchainService,
	log *logrus.Logger,
) *EventService {
	return &EventService{
		jobRepo:    jobRepo,
		blockchain: blockchain,
		log:        log,
	}
}

func (s *EventService) Process(ctx context.Context, event *domain.Event) error {
	log := s.log.WithFields(logrus.Fields{
		"event_id":   event.ID,
		"event_type": string(event.Type),
	})

	existing, err := s.jobRepo.FindByEventID(ctx, event.ID)
	if err != nil {
		return fmt.Errorf("idempotency check: %w", err)
	}
	if existing != nil {
		if existing.Status == domain.JobStatusSuccess {
			log.Info("event already processed successfully, skipping")
			return domain.ErrAlreadyProcessed
		}
		log.WithField("job_id", existing.ID).Warn("re-processing previously failed event")
	}

	job := &domain.Job{
		ID:      uuid.NewString(),
		EventID: event.ID,
		Type:    string(event.Type),
		Status:  domain.JobStatusPending,
	}
	if existing != nil {
		job = existing // reuse the row if it already exists
	} else {
		if err := s.jobRepo.Create(ctx, job); err != nil {
			return fmt.Errorf("create pending job: %w", err)
		}
	}

	log = log.WithField("job_id", job.ID)
	log.Info("processing event")

	result, bcErr := s.dispatch(ctx, event)

	if bcErr != nil {
		errMsg := bcErr.Error()
		_ = s.jobRepo.UpdateStatus(ctx, job.ID, domain.JobStatusFailed, nil, errMsg)
		log.WithError(bcErr).Error("blockchain operation failed")
		return &domain.ErrBlockchain{Op: string(event.Type), Err: bcErr}
	}

	if err := s.jobRepo.UpdateStatus(ctx, job.ID, domain.JobStatusSuccess, result.TxHashes, ""); err != nil {
		log.WithError(err).Error("failed to update job status to success")
	}

	log.WithField("tx_hashes", result.TxHashes).Info("event processed successfully")
	return nil
}

func (s *EventService) dispatch(ctx context.Context, event *domain.Event) (*domain.BlockchainResult, error) {
	switch event.Type {
	case domain.EventTypeCreateDID:
		var p domain.CreateDIDPayload
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return nil, fmt.Errorf("unmarshal CreateDIDPayload: %w", err)
		}
		if p.DID == "" {
			return nil, &domain.ErrValidation{Field: "did", Message: "required"}
		}
		return s.blockchain.CreateDID(ctx, p)

	case domain.EventTypeAddKey:
		var p domain.AddKeyPayload
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return nil, fmt.Errorf("unmarshal AddKeyPayload: %w", err)
		}
		if p.PublicKey == "" {
			return nil, &domain.ErrValidation{Field: "public_key", Message: "required"}
		}
		return s.blockchain.AddKey(ctx, p)

	case domain.EventTypeAddClaim:
		var p domain.AddClaimPayload
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return nil, fmt.Errorf("unmarshal AddClaimPayload: %w", err)
		}
		if p.IssuerAddress == "" {
			return nil, &domain.ErrValidation{Field: "issuer_address", Message: "required"}
		}
		return s.blockchain.AddClaim(ctx, p)

	case domain.EventTypeStoreData:
		var p domain.StoreDataPayload
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return nil, fmt.Errorf("unmarshal StoreDataPayload: %w", err)
		}
		if p.PropertyKey == "" {
			return nil, &domain.ErrValidation{Field: "property_key", Message: "required"}
		}
		return s.blockchain.StoreData(ctx, p)

	case domain.EventTypeHandleOps:
		var p domain.HandleOpsPayload
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return nil, fmt.Errorf("unmarshal HandleOpsPayload: %w", err)
		}
		return s.blockchain.HandleOps(ctx, p)

	case domain.EventTypeDeployWallet:
		var p domain.DeployWalletPayload
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return nil, fmt.Errorf("unmarshal DeployWalletPayload: %w", err)
		}
		return s.blockchain.DeployWallet(ctx, p)

	case domain.EventTypeUpdateDID:
		var p domain.UpdateDIDPayload
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return nil, fmt.Errorf("unmarshal UpdateDIDPayload: %w", err)
		}
		return s.blockchain.UpdateDID(ctx, p)

	case domain.EventTypeDeactivateDID:
		var p domain.DIDLifecyclePayload
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return nil, fmt.Errorf("unmarshal DIDLifecyclePayload: %w", err)
		}
		return s.blockchain.DeactivateDID(ctx, p)

	case domain.EventTypeReactivateDID:
		var p domain.DIDLifecyclePayload
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return nil, fmt.Errorf("unmarshal DIDLifecyclePayload: %w", err)
		}
		return s.blockchain.ReactivateDID(ctx, p)

	case domain.EventTypeTransferDID:
		var p domain.TransferDIDOwnerPayload
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return nil, fmt.Errorf("unmarshal TransferDIDOwnerPayload: %w", err)
		}
		return s.blockchain.TransferDIDOwner(ctx, p)

	case domain.EventTypeDeleteData:
		var p domain.DeleteDataPayload
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return nil, fmt.Errorf("unmarshal DeleteDataPayload: %w", err)
		}
		return s.blockchain.DeleteData(ctx, p)

	case domain.EventTypeRemoveKey:
		var p domain.RemoveKeyPayload
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return nil, fmt.Errorf("unmarshal RemoveKeyPayload: %w", err)
		}
		return s.blockchain.RemoveKey(ctx, p)

	case domain.EventTypeRemoveClaim:
		var p domain.RemoveClaimPayload
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return nil, fmt.Errorf("unmarshal RemoveClaimPayload: %w", err)
		}
		return s.blockchain.RemoveClaim(ctx, p)

	case domain.EventTypeGeneralExecute:
		var p domain.GeneralExecutePayload
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return nil, fmt.Errorf("unmarshal GeneralExecutePayload: %w", err)
		}
		return s.blockchain.GeneralExecute(ctx, p)

	case domain.EventTypeCreateOrg:
		var p domain.CreateOrgPayload
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return nil, fmt.Errorf("unmarshal CreateOrgPayload: %w", err)
		}
		return s.blockchain.CreateOrg(ctx, p)

	case domain.EventTypeAddMember:
		var p domain.OrgMemberPayload
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return nil, fmt.Errorf("unmarshal OrgMemberPayload: %w", err)
		}
		return s.blockchain.AddMember(ctx, p)

	case domain.EventTypeRemoveMember:
		var p domain.OrgMemberPayload
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return nil, fmt.Errorf("unmarshal OrgMemberPayload: %w", err)
		}
		return s.blockchain.RemoveMember(ctx, p)

	case domain.EventTypeUpdateMember:
		var p domain.OrgMemberPayload
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return nil, fmt.Errorf("unmarshal OrgMemberPayload: %w", err)
		}
		return s.blockchain.UpdateMember(ctx, p)

	case domain.EventTypeDeactivateOrg:
		var p domain.OrgLifecyclePayload
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return nil, fmt.Errorf("unmarshal OrgLifecyclePayload: %w", err)
		}
		return s.blockchain.DeactivateOrg(ctx, p)

	case domain.EventTypeReactivateOrg:
		var p domain.OrgLifecyclePayload
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return nil, fmt.Errorf("unmarshal OrgLifecyclePayload: %w", err)
		}
		return s.blockchain.ReactivateOrg(ctx, p)

	case domain.EventTypeTransferOrgOwner:
		var p domain.OrgTransferPayload
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return nil, fmt.Errorf("unmarshal OrgTransferPayload: %w", err)
		}
		return s.blockchain.TransferOrgOwner(ctx, p)

	// ── Verifiable Credentials ─────────────────────────────────────
	case domain.EventTypeIssueCredential:
		var p domain.IssueCredentialPayload
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return nil, fmt.Errorf("unmarshal IssueCredentialPayload: %w", err)
		}
		return s.blockchain.IssueCredential(ctx, p)

	case domain.EventTypeBurnCredential:
		var p domain.BurnCredentialPayload
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return nil, fmt.Errorf("unmarshal BurnCredentialPayload: %w", err)
		}
		return s.blockchain.BurnCredential(ctx, p)

	case domain.EventTypeUpdateMetadata:
		var p domain.UpdateMetadataPayload
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return nil, fmt.Errorf("unmarshal UpdateMetadataPayload: %w", err)
		}
		return s.blockchain.UpdateMetadata(ctx, p)

	case domain.EventTypeRevokeCredential:
		var p domain.RevokeCredentialPayload
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return nil, fmt.Errorf("unmarshal RevokeCredentialPayload: %w", err)
		}
		return s.blockchain.RevokeCredential(ctx, p)

	case domain.EventTypeApproveCredentialOrg:
		var p domain.ApproveCredentialOrgPayload
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return nil, fmt.Errorf("unmarshal ApproveCredentialOrgPayload: %w", err)
		}
		return s.blockchain.ApproveCredentialOrg(ctx, p)

	case domain.EventTypeApproveCredential:
		var p domain.ApproveCredentialPayload
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return nil, fmt.Errorf("unmarshal ApproveCredentialPayload: %w", err)
		}
		return s.blockchain.ApproveCredential(ctx, p)

	case domain.EventTypeSetDidRootStorage:
		var p domain.SetAddressPayload
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return nil, fmt.Errorf("unmarshal SetAddressPayload: %w", err)
		}
		return s.blockchain.SetDidRootStorage(ctx, p)

	case domain.EventTypeSetDidOrgStorage:
		var p domain.SetAddressPayload
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return nil, fmt.Errorf("unmarshal SetAddressPayload: %w", err)
		}
		return s.blockchain.SetDidOrgStorage(ctx, p)

	// ── Alias ──────────────────────────────────────────────────────
	case domain.EventTypeRegisterTLD:
		var p domain.RegisterTLDPayload
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return nil, fmt.Errorf("unmarshal RegisterTLDPayload: %w", err)
		}
		return s.blockchain.RegisterTLD(ctx, p)

	case domain.EventTypeRegisterDomain:
		var p domain.RegisterDomainPayload
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return nil, fmt.Errorf("unmarshal RegisterDomainPayload: %w", err)
		}
		return s.blockchain.RegisterDomain(ctx, p)

	case domain.EventTypeSetDIDAlias:
		var p domain.SetDIDAliasPayload
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return nil, fmt.Errorf("unmarshal SetDIDAliasPayload: %w", err)
		}
		return s.blockchain.SetDIDAlias(ctx, p)

	case domain.EventTypeSetDIDOrgAlias:
		var p domain.SetDIDOrgAliasPayload
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return nil, fmt.Errorf("unmarshal SetDIDOrgAliasPayload: %w", err)
		}
		return s.blockchain.SetDIDOrgAlias(ctx, p)

	case domain.EventTypeExtendRegistration:
		var p domain.ExtendRegistrationPayload
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return nil, fmt.Errorf("unmarshal ExtendRegistrationPayload: %w", err)
		}
		return s.blockchain.ExtendRegistration(ctx, p)

	case domain.EventTypeRevokeAlias:
		var p domain.RevokeAliasPayload
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return nil, fmt.Errorf("unmarshal RevokeAliasPayload: %w", err)
		}
		return s.blockchain.RevokeAlias(ctx, p)

	case domain.EventTypeUnrevokeAlias:
		var p domain.UnrevokeAliasPayload
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return nil, fmt.Errorf("unmarshal UnrevokeAliasPayload: %w", err)
		}
		return s.blockchain.UnrevokeAlias(ctx, p)

	case domain.EventTypeRegisterSubdomain:
		var p domain.RegisterSubdomainPayload
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return nil, fmt.Errorf("unmarshal RegisterSubdomainPayload: %w", err)
		}
		return s.blockchain.RegisterSubdomain(ctx, p)

	case domain.EventTypeTransferAliasOwnership:
		var p domain.TransferAliasOwnershipPayload
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return nil, fmt.Errorf("unmarshal TransferAliasOwnershipPayload: %w", err)
		}
		return s.blockchain.TransferAliasOwnership(ctx, p)

	case domain.EventTypeTransferTLD:
		var p domain.TransferTLDPayload
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return nil, fmt.Errorf("unmarshal TransferTLDPayload: %w", err)
		}
		return s.blockchain.TransferTLD(ctx, p)

	case domain.EventTypeSetAliasRootStorage:
		var p domain.SetAliasAddressPayload
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return nil, fmt.Errorf("unmarshal SetAliasAddressPayload: %w", err)
		}
		return s.blockchain.SetAliasRootStorage(ctx, p)

	case domain.EventTypeSetAliasOrgStorage:
		var p domain.SetAliasAddressPayload
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return nil, fmt.Errorf("unmarshal SetAliasAddressPayload: %w", err)
		}
		return s.blockchain.SetAliasOrgStorage(ctx, p)

	case domain.EventTypeSetFactoryContract:
		var p domain.SetFactoryContractPayload
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return nil, fmt.Errorf("unmarshal SetFactoryContractPayload: %w", err)
		}
		return s.blockchain.SetFactoryContract(ctx, p)

	default:
		return nil, fmt.Errorf("unknown event type: %q", event.Type)
	}
}

func (s *EventService) RecordRetry(ctx context.Context, jobID, errMsg string) {
	if err := s.jobRepo.IncrementRetry(ctx, jobID, errMsg); err != nil {
		s.log.WithError(err).WithField("job_id", jobID).Error("failed to record retry")
	}
}

func IsAlreadyProcessed(err error) bool {
	return errors.Is(err, domain.ErrAlreadyProcessed)
}
