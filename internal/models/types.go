package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type CustomTime struct {
	time.Time
}

func (ct *CustomTime) UnmarshalJSON(data []byte) error {
	str := string(data[1 : len(data)-1])

	layouts := []string{
		"2006-01-02T15:04:05",
		time.RFC3339,
		"2006-01-02T15:04:05Z07:00",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, str); err == nil {
			ct.Time = t
			return nil
		}
	}

	return json.Unmarshal(data, &ct.Time)
}

type Pharmacy struct {
	ID    int    `json:"id" db:"id"`
	NPI   string `json:"npi" db:"npi"`
	Chain string `json:"chain" db:"chain"`
}

type Claim struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	NDC       string     `json:"ndc" db:"ndc"`
	Quantity  float64    `json:"quantity" db:"quantity"`
	NPI       string     `json:"npi" db:"npi"`
	Price     float64    `json:"price" db:"price"`
	Timestamp CustomTime `json:"timestamp" db:"timestamp"`
}

type Reversal struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	ClaimID   uuid.UUID  `json:"claim_id" db:"claim_id"`
	Timestamp CustomTime `json:"timestamp" db:"timestamp"`
}

type ClaimRequest struct {
	NDC      string  `json:"ndc"`
	Quantity float64 `json:"quantity"`
	NPI      string  `json:"npi"`
	Price    float64 `json:"price"`
}

type ClaimResponse struct {
	Status  string    `json:"status"`
	ClaimID uuid.UUID `json:"claim_id"`
}

type ReversalRequest struct {
	ClaimID uuid.UUID `json:"claim_id"`
	Reason  string    `json:"reason,omitempty"`
}

type ReversalResponse struct {
	Status  string    `json:"status"`
	ClaimID uuid.UUID `json:"claim_id"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}
