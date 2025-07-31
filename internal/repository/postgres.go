package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"pharmacyclaims/internal/database"
	"pharmacyclaims/internal/models"

	"github.com/google/uuid"
)

type Postgres struct {
	db *database.DB
}

func NewPostgresRepository(db *database.DB) *Postgres {
	return &Postgres{db: db}
}

func (pr *Postgres) GetPharmacyByNPI(npi string) (*models.Pharmacy, error) {
	query := `
		SELECT id, npi, chain
		FROM pharmacies
		WHERE npi = $1`

	pharmacy := &models.Pharmacy{}
	err := pr.db.QueryRow(query, npi).Scan(
		&pharmacy.ID,
		&pharmacy.NPI,
		&pharmacy.Chain,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get pharmacy by NPI: %w", err)
	}

	return pharmacy, nil
}

func (pr *Postgres) CreateClaim(claim *models.Claim) error {
	query := `
		INSERT INTO claims (id, ndc, quantity, npi, price, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := pr.db.Exec(query,
		claim.ID,
		claim.NDC,
		claim.Quantity,
		claim.NPI,
		claim.Price,
		claim.Timestamp.Time,
	)

	if err != nil {
		return fmt.Errorf("failed to create claim: %w", err)
	}

	return nil
}

func (pr *Postgres) GetClaimByID(id uuid.UUID) (*models.Claim, error) {
	query := `
		SELECT id, ndc, quantity, npi, price, timestamp
		FROM claims
		WHERE id = $1`

	claim := &models.Claim{}
	var timestamp time.Time
	err := pr.db.QueryRow(query, id).Scan(
		&claim.ID,
		&claim.NDC,
		&claim.Quantity,
		&claim.NPI,
		&claim.Price,
		&timestamp,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get claim by ID: %w", err)
	}

	claim.Timestamp = models.CustomTime{Time: timestamp}
	return claim, nil
}

func (pr *Postgres) ReverseClaim(claimID uuid.UUID, reason string) error {
	return pr.db.ExecuteInTransaction(func(tx *sql.Tx) error {
		var exists bool
		checkQuery := `SELECT EXISTS(SELECT 1 FROM claims WHERE id = $1)`
		err := tx.QueryRow(checkQuery, claimID).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check if claim exists: %w", err)
		}

		if !exists {
			return fmt.Errorf("claim not found")
		}

		var reversalExists bool
		reversalCheckQuery := `SELECT EXISTS(SELECT 1 FROM reversals WHERE claim_id = $1)`
		err = tx.QueryRow(reversalCheckQuery, claimID).Scan(&reversalExists)
		if err != nil {
			return fmt.Errorf("failed to check if claim already reversed: %w", err)
		}

		if reversalExists {
			return fmt.Errorf("claim already reversed")
		}

		insertReversalQuery := `
			INSERT INTO reversals (id, claim_id, timestamp)
			VALUES ($1, $2, $3)`

		reversalID := uuid.New()
		now := time.Now()
		_, err = tx.Exec(insertReversalQuery, reversalID, claimID, now)
		if err != nil {
			return fmt.Errorf("failed to create reversal record: %w", err)
		}

		return nil
	})
}

func (pr *Postgres) BatchCreatePharmacies(pharmacies []models.Pharmacy) error {
	columns := []string{"npi", "chain"}
	values := make([][]interface{}, len(pharmacies))

	for i, pharmacy := range pharmacies {
		values[i] = []interface{}{pharmacy.NPI, pharmacy.Chain}
	}

	return pr.batchInsert("pharmacies", columns, values)
}

func (pr *Postgres) BatchCreateClaims(claims []models.Claim) error {
	columns := []string{"id", "ndc", "quantity", "npi", "price", "timestamp"}
	values := make([][]interface{}, len(claims))

	for i, claim := range claims {
		values[i] = []interface{}{claim.ID, claim.NDC, claim.Quantity, claim.NPI, claim.Price, claim.Timestamp.Time}
	}

	return pr.batchInsert("claims", columns, values)
}

func (pr *Postgres) BatchCreateReversals(reversals []models.Reversal) error {
	columns := []string{"id", "claim_id", "timestamp"}
	values := make([][]interface{}, len(reversals))

	for i, reversal := range reversals {
		values[i] = []interface{}{reversal.ID, reversal.ClaimID, reversal.Timestamp.Time}
	}

	return pr.batchInsert("reversals", columns, values)
}

func (pr *Postgres) batchInsert(tableName string, columns []string, values [][]interface{}) error {
	tx, err := pr.db.BeginTx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	placeholders := make([]string, len(columns))
	for i := range placeholders {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s) ON CONFLICT DO NOTHING",
		tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	stmt, err := tx.Prepare(query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, row := range values {
		_, err := stmt.Exec(row...)
		if err != nil {
			return fmt.Errorf("failed to execute insert: %w", err)
		}
	}

	return tx.Commit()
}

func (pr *Postgres) CountPharmacies() (int, error) {
	return pr.countRows("pharmacies")
}

func (pr *Postgres) CountClaims() (int, error) {
	return pr.countRows("claims")
}

func (pr *Postgres) CountReversals() (int, error) {
	return pr.countRows("reversals")
}

func (pr *Postgres) countRows(tableName string) (int, error) {
	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
	err := pr.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count rows in %s: %w", tableName, err)
	}
	return count, nil
}
