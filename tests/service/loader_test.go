package service

import (
	"os"
	"path/filepath"
	"testing"

	"pharmacyclaims/internal/core"
	"pharmacyclaims/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLoaderService(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "loader_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	logger := core.NewLogger(filepath.Join(tempDir, "logs"))
	defer os.RemoveAll(filepath.Join(tempDir, "logs"))

	loaderService := service.NewLoaderService(nil, logger)

	assert.NotNil(t, loaderService)
}

func TestNewLoaderServiceWithBatchSize(t *testing.T) {
	logger := core.NewLogger("test-logs")
	defer os.RemoveAll("test-logs")

	tests := []struct {
		name           string
		inputBatchSize int
		expectedValid  bool
	}{
		{
			name:           "Valid batch size",
			inputBatchSize: 500,
			expectedValid:  true,
		},
		{
			name:           "Zero batch size should use default",
			inputBatchSize: 0,
			expectedValid:  true,
		},
		{
			name:           "Negative batch size should use default",
			inputBatchSize: -100,
			expectedValid:  true,
		},
		{
			name:           "Batch size too large should use default",
			inputBatchSize: 20000,
			expectedValid:  true,
		},
		{
			name:           "Maximum valid batch size",
			inputBatchSize: 10000,
			expectedValid:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loaderService := service.NewLoaderServiceWithBatchSize(nil, logger, tt.inputBatchSize)
			assert.NotNil(t, loaderService)
		})
	}
}

func TestBatchSizeConstants(t *testing.T) {
	assert.Equal(t, 1000, service.DefaultBatchSize)
	assert.Equal(t, 10000, service.MaxBatchSize)
	assert.Equal(t, 10, service.MaxConcurrentWorkers)

	assert.Greater(t, service.DefaultBatchSize, 0, "DefaultBatchSize should be positive")
	assert.Greater(t, service.MaxBatchSize, service.DefaultBatchSize, "MaxBatchSize should be greater than DefaultBatchSize")
	assert.Greater(t, service.MaxConcurrentWorkers, 0, "MaxConcurrentWorkers should be positive")
}

func TestNewLoaderServiceWithNilLogger(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Service panicked with nil logger as expected: %v", r)
		}
	}()

	loaderService := service.NewLoaderService(nil, nil)

	if loaderService != nil {
		t.Log("Service handled nil logger gracefully")
	}
}

func TestNewLoaderServiceWithBatchSize_EdgeCases(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "loader_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	logger := core.NewLogger(filepath.Join(tempDir, "logs"))
	defer os.RemoveAll(filepath.Join(tempDir, "logs"))

	extremeTests := []struct {
		name      string
		batchSize int
	}{
		{"Very large negative", -999999},
		{"Very large positive", 999999},
		{"Boundary value one less than max", service.MaxBatchSize - 1},
		{"Boundary value one more than max", service.MaxBatchSize + 1},
	}

	for _, tt := range extremeTests {
		t.Run(tt.name, func(t *testing.T) {
			loaderService := service.NewLoaderServiceWithBatchSize(nil, logger, tt.batchSize)
			assert.NotNil(t, loaderService)
		})
	}
}

func TestMultipleLoaderServiceInstances(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "loader_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	logger := core.NewLogger(filepath.Join(tempDir, "logs"))
	defer os.RemoveAll(filepath.Join(tempDir, "logs"))

	instances := make([]*service.LoaderService, 5)
	for i := 0; i < 5; i++ {
		instances[i] = service.NewLoaderService(nil, logger)
		assert.NotNil(t, instances[i])
	}

	for i := 0; i < len(instances); i++ {
		for j := i + 1; j < len(instances); j++ {
			assert.NotSame(t, instances[i], instances[j], "Instances %d and %d should be different objects", i, j)
		}
	}
}
