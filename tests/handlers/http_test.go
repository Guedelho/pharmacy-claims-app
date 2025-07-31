package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pharmacyclaims/internal/handlers"
	"pharmacyclaims/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockService struct {
	mock.Mock
}

func (m *MockService) ValidateClaim(request models.ClaimRequest) error {
	args := m.Called(request)
	return args.Error(0)
}

func (m *MockService) SubmitClaim(request models.ClaimRequest) (*models.ClaimResponse, error) {
	args := m.Called(request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ClaimResponse), args.Error(1)
}

func (m *MockService) ReverseClaim(request models.ReversalRequest) (*models.ReversalResponse, error) {
	args := m.Called(request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ReversalResponse), args.Error(1)
}

func TestNewHttpHandler(t *testing.T) {
	mockService := &MockService{}
	handler := handlers.NewHttpHandler(mockService)

	assert.NotNil(t, handler)
}

func TestSetupRoutes(t *testing.T) {
	mockService := &MockService{}
	handler := handlers.NewHttpHandler(mockService)
	mux := handler.SetupRoutes()

	assert.NotNil(t, mux)

	testCases := []struct {
		path   string
		method string
	}{
		{"/claim", "POST"},
		{"/reversal", "POST"},
		{"/health", "GET"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Route %s %s", tc.method, tc.path), func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rr := httptest.NewRecorder()

			mux.ServeHTTP(rr, req)

			assert.NotEqual(t, http.StatusNotFound, rr.Code)
		})
	}
}

func TestSubmitClaim_Success(t *testing.T) {
	mockService := &MockService{}
	handler := handlers.NewHttpHandler(mockService)

	claimRequest := models.ClaimRequest{
		NDC:      "1234567890",
		Quantity: 10.0,
		NPI:      "1234567890",
		Price:    29.99,
	}

	claimID := uuid.New()
	expectedResponse := &models.ClaimResponse{
		Status:  "claim submitted",
		ClaimID: claimID,
	}

	mockService.On("ValidateClaim", claimRequest).Return(nil)
	mockService.On("SubmitClaim", claimRequest).Return(expectedResponse, nil)

	requestBody, _ := json.Marshal(claimRequest)
	req := httptest.NewRequest("POST", "/claim", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.SubmitClaim(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var response models.ClaimResponse
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, expectedResponse.Status, response.Status)
	assert.Equal(t, expectedResponse.ClaimID, response.ClaimID)

	mockService.AssertExpectations(t)
}

func TestSubmitClaim_MethodNotAllowed(t *testing.T) {
	mockService := &MockService{}
	handler := handlers.NewHttpHandler(mockService)

	req := httptest.NewRequest("GET", "/claim", nil)
	rr := httptest.NewRecorder()

	handler.SubmitClaim(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var errorResponse models.ErrorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &errorResponse)
	require.NoError(t, err)
	assert.Equal(t, "Method not allowed", errorResponse.Error)
	assert.Equal(t, "Only POST method is allowed", errorResponse.Message)
}

func TestSubmitClaim_InvalidJSON(t *testing.T) {
	mockService := &MockService{}
	handler := handlers.NewHttpHandler(mockService)

	req := httptest.NewRequest("POST", "/claim", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.SubmitClaim(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var errorResponse models.ErrorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &errorResponse)
	require.NoError(t, err)
	assert.Equal(t, "Invalid JSON format", errorResponse.Error)
}

func TestSubmitClaim_ValidationFailed(t *testing.T) {
	mockService := &MockService{}
	handler := handlers.NewHttpHandler(mockService)

	claimRequest := models.ClaimRequest{
		NDC:      "invalid",
		Quantity: 10.0,
		NPI:      "1234567890",
		Price:    29.99,
	}

	mockService.On("ValidateClaim", claimRequest).Return(fmt.Errorf("invalid NDC format"))

	requestBody, _ := json.Marshal(claimRequest)
	req := httptest.NewRequest("POST", "/claim", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.SubmitClaim(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var errorResponse models.ErrorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &errorResponse)
	require.NoError(t, err)
	assert.Equal(t, "Validation failed", errorResponse.Error)
	assert.Equal(t, "invalid NDC format", errorResponse.Message)

	mockService.AssertExpectations(t)
}

func TestSubmitClaim_PharmacyNotFound(t *testing.T) {
	mockService := &MockService{}
	handler := handlers.NewHttpHandler(mockService)

	claimRequest := models.ClaimRequest{
		NDC:      "1234567890",
		Quantity: 10.0,
		NPI:      "9999999999",
		Price:    29.99,
	}

	mockService.On("ValidateClaim", claimRequest).Return(nil)
	mockService.On("SubmitClaim", claimRequest).Return(nil, fmt.Errorf("pharmacy with NPI %s not found", claimRequest.NPI))

	requestBody, _ := json.Marshal(claimRequest)
	req := httptest.NewRequest("POST", "/claim", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.SubmitClaim(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)

	var errorResponse models.ErrorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &errorResponse)
	require.NoError(t, err)
	assert.Equal(t, "Pharmacy not found", errorResponse.Error)

	mockService.AssertExpectations(t)
}

func TestSubmitClaim_InternalServerError(t *testing.T) {
	mockService := &MockService{}
	handler := handlers.NewHttpHandler(mockService)

	claimRequest := models.ClaimRequest{
		NDC:      "1234567890",
		Quantity: 10.0,
		NPI:      "1234567890",
		Price:    29.99,
	}

	mockService.On("ValidateClaim", claimRequest).Return(nil)
	mockService.On("SubmitClaim", claimRequest).Return(nil, fmt.Errorf("database connection failed"))

	requestBody, _ := json.Marshal(claimRequest)
	req := httptest.NewRequest("POST", "/claim", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.SubmitClaim(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)

	var errorResponse models.ErrorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &errorResponse)
	require.NoError(t, err)
	assert.Equal(t, "Failed to submit claim", errorResponse.Error)
	assert.Equal(t, "database connection failed", errorResponse.Message)

	mockService.AssertExpectations(t)
}

func TestReverseClaim_Success(t *testing.T) {
	mockService := &MockService{}
	handler := handlers.NewHttpHandler(mockService)

	claimID := uuid.New()
	reversalRequest := models.ReversalRequest{
		ClaimID: claimID,
		Reason:  "Customer returned item",
	}

	expectedResponse := &models.ReversalResponse{
		Status:  "claim reversed",
		ClaimID: claimID,
	}

	mockService.On("ReverseClaim", reversalRequest).Return(expectedResponse, nil)

	requestBody, _ := json.Marshal(reversalRequest)
	req := httptest.NewRequest("POST", "/reversal", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ReverseClaim(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var response models.ReversalResponse
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, expectedResponse.Status, response.Status)
	assert.Equal(t, expectedResponse.ClaimID, response.ClaimID)

	mockService.AssertExpectations(t)
}

func TestReverseClaim_MethodNotAllowed(t *testing.T) {
	mockService := &MockService{}
	handler := handlers.NewHttpHandler(mockService)

	req := httptest.NewRequest("GET", "/reversal", nil)
	rr := httptest.NewRecorder()

	handler.ReverseClaim(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)

	var errorResponse models.ErrorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &errorResponse)
	require.NoError(t, err)
	assert.Equal(t, "Method not allowed", errorResponse.Error)
	assert.Equal(t, "Only POST method is allowed", errorResponse.Message)
}

func TestReverseClaim_InvalidJSON(t *testing.T) {
	mockService := &MockService{}
	handler := handlers.NewHttpHandler(mockService)

	req := httptest.NewRequest("POST", "/reversal", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ReverseClaim(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var errorResponse models.ErrorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &errorResponse)
	require.NoError(t, err)
	assert.Equal(t, "Invalid JSON format", errorResponse.Error)
}

func TestReverseClaim_InvalidClaimID(t *testing.T) {
	mockService := &MockService{}
	handler := handlers.NewHttpHandler(mockService)

	reversalRequest := models.ReversalRequest{
		ClaimID: uuid.Nil,
		Reason:  "Customer returned item",
	}

	requestBody, _ := json.Marshal(reversalRequest)
	req := httptest.NewRequest("POST", "/reversal", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ReverseClaim(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var errorResponse models.ErrorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &errorResponse)
	require.NoError(t, err)
	assert.Equal(t, "Invalid claim_id", errorResponse.Error)
	assert.Equal(t, "claim_id must be a valid UUID", errorResponse.Message)
}

func TestReverseClaim_ClaimNotFound(t *testing.T) {
	mockService := &MockService{}
	handler := handlers.NewHttpHandler(mockService)

	claimID := uuid.New()
	reversalRequest := models.ReversalRequest{
		ClaimID: claimID,
		Reason:  "Customer returned item",
	}

	mockService.On("ReverseClaim", reversalRequest).Return(nil, fmt.Errorf("claim with ID %s not found", claimID.String()))

	requestBody, _ := json.Marshal(reversalRequest)
	req := httptest.NewRequest("POST", "/reversal", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ReverseClaim(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)

	var errorResponse models.ErrorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &errorResponse)
	require.NoError(t, err)
	assert.Equal(t, "Claim not found", errorResponse.Error)

	mockService.AssertExpectations(t)
}

func TestReverseClaim_ClaimAlreadyReversed(t *testing.T) {
	mockService := &MockService{}
	handler := handlers.NewHttpHandler(mockService)

	claimID := uuid.New()
	reversalRequest := models.ReversalRequest{
		ClaimID: claimID,
		Reason:  "Customer returned item",
	}

	mockService.On("ReverseClaim", reversalRequest).Return(nil, fmt.Errorf("claim is already reversed"))

	requestBody, _ := json.Marshal(reversalRequest)
	req := httptest.NewRequest("POST", "/reversal", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ReverseClaim(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)

	var errorResponse models.ErrorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &errorResponse)
	require.NoError(t, err)
	assert.Equal(t, "Claim already reversed", errorResponse.Error)
	assert.Equal(t, "claim is already reversed", errorResponse.Message)

	mockService.AssertExpectations(t)
}

func TestReverseClaim_InternalServerError(t *testing.T) {
	mockService := &MockService{}
	handler := handlers.NewHttpHandler(mockService)

	claimID := uuid.New()
	reversalRequest := models.ReversalRequest{
		ClaimID: claimID,
		Reason:  "Customer returned item",
	}

	mockService.On("ReverseClaim", reversalRequest).Return(nil, fmt.Errorf("database connection failed"))

	requestBody, _ := json.Marshal(reversalRequest)
	req := httptest.NewRequest("POST", "/reversal", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ReverseClaim(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)

	var errorResponse models.ErrorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &errorResponse)
	require.NoError(t, err)
	assert.Equal(t, "Failed to reverse claim", errorResponse.Error)
	assert.Equal(t, "database connection failed", errorResponse.Message)

	mockService.AssertExpectations(t)
}

func TestHealthCheck_Success(t *testing.T) {
	mockService := &MockService{}
	handler := handlers.NewHttpHandler(mockService)

	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()

	handler.HealthCheck(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var response map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])
}

func TestHealthCheck_MethodNotAllowed(t *testing.T) {
	mockService := &MockService{}
	handler := handlers.NewHttpHandler(mockService)

	req := httptest.NewRequest("POST", "/health", nil)
	rr := httptest.NewRecorder()

	handler.HealthCheck(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)

	var errorResponse models.ErrorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &errorResponse)
	require.NoError(t, err)
	assert.Equal(t, "Method not allowed", errorResponse.Error)
	assert.Equal(t, "Only GET method is allowed", errorResponse.Message)
}

func TestSendJSONResponse(t *testing.T) {
	mockService := &MockService{}
	handler := handlers.NewHttpHandler(mockService)

	rr := httptest.NewRecorder()

	handler.HealthCheck(rr, httptest.NewRequest("GET", "/health", nil))

	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])
}

func TestSendErrorResponse(t *testing.T) {
	mockService := &MockService{}
	handler := handlers.NewHttpHandler(mockService)

	req := httptest.NewRequest("POST", "/health", nil)
	rr := httptest.NewRecorder()

	handler.HealthCheck(rr, req)

	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)

	var errorResponse models.ErrorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &errorResponse)
	require.NoError(t, err)
	assert.Equal(t, "Method not allowed", errorResponse.Error)
	assert.Equal(t, "Only GET method is allowed", errorResponse.Message)
}

func TestSubmitClaim_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		requestBody string
		statusCode  int
		errorMsg    string
	}{
		{
			name:        "Empty request body",
			requestBody: "",
			statusCode:  http.StatusBadRequest,
			errorMsg:    "Invalid JSON format",
		},
		{
			name:        "Malformed JSON",
			requestBody: `{"ndc": "123", "quantity":}`,
			statusCode:  http.StatusBadRequest,
			errorMsg:    "Invalid JSON format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockService{}
			handler := handlers.NewHttpHandler(mockService)

			req := httptest.NewRequest("POST", "/claim", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			handler.SubmitClaim(rr, req)

			assert.Equal(t, tt.statusCode, rr.Code)

			if tt.statusCode != http.StatusOK && tt.statusCode != http.StatusCreated {
				var errorResponse models.ErrorResponse
				err := json.Unmarshal(rr.Body.Bytes(), &errorResponse)
				require.NoError(t, err)
				assert.Equal(t, tt.errorMsg, errorResponse.Error)
			}
		})
	}
}

func TestSubmitClaim_NullRequestBody(t *testing.T) {
	mockService := &MockService{}
	handler := handlers.NewHttpHandler(mockService)

	mockService.On("ValidateClaim", models.ClaimRequest{}).Return(fmt.Errorf("invalid NDC format: must be 9-11 digits"))

	req := httptest.NewRequest("POST", "/claim", strings.NewReader("null"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.SubmitClaim(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var errorResponse models.ErrorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &errorResponse)
	require.NoError(t, err)
	assert.Equal(t, "Validation failed", errorResponse.Error)

	mockService.AssertExpectations(t)
}

func TestReverseClaim_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		requestBody string
		statusCode  int
		errorMsg    string
	}{
		{
			name:        "Empty request body",
			requestBody: "",
			statusCode:  http.StatusBadRequest,
			errorMsg:    "Invalid JSON format",
		},
		{
			name:        "Malformed JSON",
			requestBody: `{"claim_id": "invalid-uuid"}`,
			statusCode:  http.StatusBadRequest,
			errorMsg:    "Invalid JSON format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockService{}
			handler := handlers.NewHttpHandler(mockService)

			req := httptest.NewRequest("POST", "/reversal", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			handler.ReverseClaim(rr, req)

			assert.Equal(t, tt.statusCode, rr.Code)

			if tt.statusCode != http.StatusOK {
				var errorResponse models.ErrorResponse
				err := json.Unmarshal(rr.Body.Bytes(), &errorResponse)
				require.NoError(t, err)
				assert.Equal(t, tt.errorMsg, errorResponse.Error)
			}
		})
	}
}

func TestReverseClaim_NullRequestBody(t *testing.T) {
	mockService := &MockService{}
	handler := handlers.NewHttpHandler(mockService)

	req := httptest.NewRequest("POST", "/reversal", strings.NewReader("null"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ReverseClaim(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var errorResponse models.ErrorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &errorResponse)
	require.NoError(t, err)
	assert.Equal(t, "Invalid claim_id", errorResponse.Error)
	assert.Equal(t, "claim_id must be a valid UUID", errorResponse.Message)
}
