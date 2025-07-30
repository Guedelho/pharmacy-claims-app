package utility

import (
	"fmt"
	"strconv"

	"pharmacyclaims/internal/models"
)

type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) ValidateClaimRequest(request models.ClaimRequest) error {
	if err := v.ValidateNDC(request.NDC); err != nil {
		return err
	}

	if err := v.ValidateNPI(request.NPI); err != nil {
		return err
	}

	if err := v.ValidateQuantity(request.Quantity); err != nil {
		return err
	}

	if err := v.ValidatePrice(request.Price); err != nil {
		return err
	}

	return nil
}

func (v *Validator) ValidateNDC(ndc string) error {
	if len(ndc) < 9 || len(ndc) > 11 {
		return fmt.Errorf("invalid NDC format: must be 9-11 digits")
	}
	if _, err := strconv.Atoi(ndc); err != nil {
		return fmt.Errorf("invalid NDC format: must be numeric")
	}
	return nil
}

func (v *Validator) ValidateNPI(npi string) error {
	if len(npi) != 10 {
		return fmt.Errorf("invalid NPI: must be exactly 10 digits")
	}
	if _, err := strconv.Atoi(npi); err != nil {
		return fmt.Errorf("invalid NPI: must be numeric")
	}
	return nil
}

func (v *Validator) ValidateQuantity(quantity float64) error {
	if quantity <= 0 {
		return fmt.Errorf("invalid quantity: must be greater than 0")
	}
	return nil
}

func (v *Validator) ValidatePrice(price float64) error {
	if price < 0 {
		return fmt.Errorf("invalid price: must be non-negative")
	}
	return nil
}
