package service

import (
	"fmt"
	"log"
	"time"

	"pharmacyclaims/internal/core"
	"pharmacyclaims/internal/models"
	"pharmacyclaims/internal/repository"
	"pharmacyclaims/internal/utility"

	"github.com/google/uuid"
)

type ClaimsService struct {
	repo      *repository.Postgres
	logger    *core.Logger
	validator *utility.Validator
}

func NewClaimsService(repo *repository.Postgres, logger *core.Logger) *ClaimsService {
	return &ClaimsService{
		repo:      repo,
		logger:    logger,
		validator: utility.NewValidator(),
	}
}

func (cs *ClaimsService) SubmitClaim(request models.ClaimRequest) (*models.ClaimResponse, error) {
	if err := cs.ValidateClaim(request); err != nil {
		return nil, err
	}

	pharmacy, err := cs.repo.GetPharmacyByNPI(request.NPI)
	if err != nil {
		return nil, fmt.Errorf("failed to validate pharmacy: %w", err)
	}
	if pharmacy == nil {
		return nil, fmt.Errorf("pharmacy with NPI %s not found", request.NPI)
	}

	claim := &models.Claim{
		ID:        uuid.New(),
		NDC:       request.NDC,
		Quantity:  request.Quantity,
		NPI:       request.NPI,
		Price:     request.Price,
		Timestamp: models.CustomTime{Time: time.Now()},
	}

	err = cs.repo.CreateClaim(claim)
	if err != nil {
		return nil, fmt.Errorf("failed to create claim: %w", err)
	}

	cs.logger.LogEvent("claim_submitted", map[string]interface{}{
		"claim_id": claim.ID.String(),
		"ndc":      claim.NDC,
		"quantity": claim.Quantity,
		"npi":      claim.NPI,
		"price":    claim.Price,
		"chain":    pharmacy.Chain,
	})

	return &models.ClaimResponse{
		Status:  "claim submitted",
		ClaimID: claim.ID,
	}, nil
}

func (cs *ClaimsService) ReverseClaim(request models.ReversalRequest) (*models.ReversalResponse, error) {
	claim, err := cs.repo.GetClaimByID(request.ClaimID)
	if err != nil {
		return nil, fmt.Errorf("failed to get claim: %w", err)
	}
	if claim == nil {
		return nil, fmt.Errorf("claim with ID %s not found", request.ClaimID.String())
	}

	err = cs.repo.ReverseClaim(request.ClaimID, request.Reason)
	if err != nil {
		return nil, fmt.Errorf("failed to reverse claim: %w", err)
	}

	pharmacy, err := cs.repo.GetPharmacyByNPI(claim.NPI)
	if err != nil {
		log.Printf("Failed to get pharmacy for logging: %v", err)
	}

	logPayload := map[string]interface{}{
		"claim_id":          claim.ID.String(),
		"original_ndc":      claim.NDC,
		"original_quantity": claim.Quantity,
		"original_npi":      claim.NPI,
		"original_price":    claim.Price,
		"reason":            request.Reason,
	}
	if pharmacy != nil {
		logPayload["chain"] = pharmacy.Chain
	}

	cs.logger.LogEvent("claim_reversed", logPayload)

	return &models.ReversalResponse{
		Status:  "claim reversed",
		ClaimID: claim.ID,
	}, nil
}

func (cs *ClaimsService) ValidateClaim(request models.ClaimRequest) error {
	return cs.validator.ValidateClaimRequest(request)
}
