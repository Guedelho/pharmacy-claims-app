package service

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"pharmacyclaims/internal/core"
	"pharmacyclaims/internal/models"
	"pharmacyclaims/internal/repository"
	"pharmacyclaims/internal/utility"
)

const (
	DefaultBatchSize     = 1000
	MaxBatchSize         = 10000
	MaxConcurrentWorkers = 10
)

type LoaderService struct {
	repo      *repository.Postgres
	logger    *core.Logger
	validator *utility.Validator
	batchSize int
}

func NewLoaderService(repo *repository.Postgres, logger *core.Logger) *LoaderService {
	return NewLoaderServiceWithBatchSize(repo, logger, DefaultBatchSize)
}

func NewLoaderServiceWithBatchSize(repo *repository.Postgres, logger *core.Logger, batchSize int) *LoaderService {
	if batchSize <= 0 || batchSize > MaxBatchSize {
		log.Printf("Invalid batch size %d, using default %d", batchSize, DefaultBatchSize)
		batchSize = DefaultBatchSize
	}

	return &LoaderService{
		repo:      repo,
		logger:    logger,
		validator: utility.NewValidator(),
		batchSize: batchSize,
	}
}

func (ls *LoaderService) LoadPharmaciesFromData(dataDir string) error {
	count, err := ls.repo.CountPharmacies()
	if err != nil {
		log.Printf("Warning: Failed to check pharmacy count: %v", err)
	} else if count > 0 {
		log.Printf("Pharmacies already loaded (%d records found), skipping data loading", count)
		return nil
	}

	pharmaciesDir := filepath.Join(dataDir, "pharmacies")

	if _, err := os.Stat(pharmaciesDir); os.IsNotExist(err) {
		return fmt.Errorf("pharmacies directory not found: %s", pharmaciesDir)
	}

	files, err := filepath.Glob(filepath.Join(pharmaciesDir, "*.csv"))
	if err != nil {
		return fmt.Errorf("failed to glob pharmacy files: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no pharmacy CSV files found in %s", pharmaciesDir)
	}

	totalLoaded := 0

	for _, file := range files {
		loaded, err := ls.loadPharmaciesFromCSV(file)
		if err != nil {
			log.Printf("Failed to load pharmacies from %s: %v", file, err)
			continue
		}
		totalLoaded += loaded
	}

	if totalLoaded == 0 {
		return fmt.Errorf("no pharmacies loaded from data directory")
	}

	log.Printf("Successfully loaded %d pharmacies from data directory", totalLoaded)
	return nil
}

func (ls *LoaderService) loadPharmaciesFromCSV(filename string) (int, error) {
	file, err := os.Open(filename)
	if err != nil {
		return 0, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	if _, err := reader.Read(); err != nil {
		return 0, fmt.Errorf("failed to read header: %w", err)
	}

	var batch []models.Pharmacy
	totalLoaded := 0
	lineNumber := 1

	for {
		record, err := reader.Read()
		if err != nil {
			if err.Error() == "EOF" {
				if len(batch) > 0 {
					if err := ls.processPharmaciesBatch(batch); err != nil {
						return totalLoaded, fmt.Errorf("failed to process final batch: %w", err)
					}
					totalLoaded += len(batch)
				}
				break
			}
			log.Printf("Error reading line %d in %s: %v", lineNumber+1, filename, err)
			lineNumber++
			continue
		}

		lineNumber++

		if len(record) < 2 {
			log.Printf("Skipping invalid record at line %d: %v", lineNumber, record)
			continue
		}

		pharmacy := models.Pharmacy{
			Chain: strings.TrimSpace(record[0]),
			NPI:   strings.TrimSpace(record[1]),
		}

		if err := ls.validator.ValidateNPI(pharmacy.NPI); err != nil {
			log.Printf("%v", err)
			continue
		}

		batch = append(batch, pharmacy)

		if len(batch) >= ls.batchSize {
			if err := ls.processPharmaciesBatch(batch); err != nil {
				return totalLoaded, fmt.Errorf("failed to process batch at line %d: %w", lineNumber, err)
			}
			totalLoaded += len(batch)
			batch = batch[:0]
		}
	}

	return totalLoaded, nil
}

func (ls *LoaderService) processPharmaciesBatch(pharmacies []models.Pharmacy) error {
	if err := ls.repo.BatchCreatePharmacies(pharmacies); err != nil {
		return fmt.Errorf("failed to batch create pharmacies: %w", err)
	}

	for _, pharmacy := range pharmacies {
		ls.logger.LogEvent("pharmacy_loaded", map[string]interface{}{
			"npi":   pharmacy.NPI,
			"chain": pharmacy.Chain,
		})
	}

	return nil
}

func loadDataFromFiles[T any](
	ls *LoaderService,
	dataDir, subDir, filePattern string,
	countFunc func() (int, error),
	fileLoader func(string) ([]T, error),
	batchProcessor func([]T) error,
	dataTypeName string,
) error {
	count, err := countFunc()
	if err != nil {
		log.Printf("Warning: Failed to check %s count: %v", dataTypeName, err)
	} else if count > 0 {
		log.Printf("%s already loaded (%d records found), skipping data loading",
			dataTypeName, count)
		return nil
	}

	targetDir := filepath.Join(dataDir, subDir)
	files, err := filepath.Glob(filepath.Join(targetDir, filePattern))
	if err != nil {
		return fmt.Errorf("failed to read %s directory: %v", dataTypeName, err)
	}

	if len(files) == 0 {
		log.Printf("No %s files found in %s", dataTypeName, targetDir)
		return nil
	}

	log.Printf("Found %d %s files to load", len(files), dataTypeName)

	dataChan := make(chan []T, len(files))
	errorChan := make(chan error, len(files))

	maxWorkers := MaxConcurrentWorkers
	if len(files) < maxWorkers {
		maxWorkers = len(files)
	}

	workChan := make(chan string, len(files))

	for i := 0; i < maxWorkers; i++ {
		go func() {
			for filename := range workChan {
				data, err := fileLoader(filename)
				if err != nil {
					log.Printf("Warning: Failed to load %s from %s: %v", dataTypeName, filename, err)
					errorChan <- err
					continue
				}
				log.Printf("Loaded %d %s from %s", len(data), dataTypeName, filepath.Base(filename))
				dataChan <- data
			}
		}()
	}

	for _, file := range files {
		workChan <- file
	}
	close(workChan)

	totalItems := 0
	var batch []T
	filesProcessed := 0

	for filesProcessed < len(files) {
		select {
		case data := <-dataChan:
			for _, item := range data {
				batch = append(batch, item)

				if len(batch) >= ls.batchSize {
					if err := batchProcessor(batch); err != nil {
						log.Printf("Warning: Failed to process %s batch: %v", dataTypeName, err)
					} else {
						totalItems += len(batch)
						log.Printf("Processed batch of %d %s", len(batch), dataTypeName)
					}
					batch = batch[:0]
				}
			}
			filesProcessed++
		case <-errorChan:
			filesProcessed++
		}
	}

	if len(batch) > 0 {
		if err := batchProcessor(batch); err != nil {
			log.Printf("Warning: Failed to process final %s batch: %v", dataTypeName, err)
		} else {
			totalItems += len(batch)
			log.Printf("Processed final batch of %d %s", len(batch), dataTypeName)
		}
	}

	log.Printf("Successfully loaded %d total %s from %d files", totalItems, dataTypeName, len(files))
	return nil
}

func loadJSONFromFile[T any](filename string) ([]T, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	jsonStr := strings.TrimSuffix(string(data), "%")
	jsonStr = strings.TrimSpace(jsonStr)

	var items []T
	err = json.Unmarshal([]byte(jsonStr), &items)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	return items, nil
}

func (ls *LoaderService) LoadClaimsFromData(dataDir string) error {
	return loadDataFromFiles(
		ls,
		dataDir,
		"claims",
		"*.json",
		ls.repo.CountClaims,
		loadJSONFromFile[models.Claim],
		ls.processClaimsBatch,
		"claims",
	)
}

func (ls *LoaderService) processClaimsBatch(claims []models.Claim) error {
	if err := ls.repo.BatchCreateClaims(claims); err != nil {
		return fmt.Errorf("failed to batch create claims: %w", err)
	}

	for _, claim := range claims {
		ls.logger.LogEvent("claim_loaded", map[string]interface{}{
			"id":       claim.ID,
			"ndc":      claim.NDC,
			"npi":      claim.NPI,
			"quantity": claim.Quantity,
			"price":    claim.Price,
		})
	}

	return nil
}

func (ls *LoaderService) LoadReversalsFromData(dataDir string) error {
	return loadDataFromFiles(
		ls,
		dataDir,
		"reverts",
		"*.json",
		ls.repo.CountReversals,
		loadJSONFromFile[models.Reversal],
		ls.processReversalsBatch,
		"reversals",
	)
}

func (ls *LoaderService) processReversalsBatch(reversals []models.Reversal) error {
	if err := ls.repo.BatchCreateReversals(reversals); err != nil {
		return fmt.Errorf("failed to batch create reversals: %w", err)
	}

	for _, reversal := range reversals {
		ls.logger.LogEvent("reversal_loaded", map[string]interface{}{
			"id":       reversal.ID,
			"claim_id": reversal.ClaimID,
		})
	}

	return nil
}
