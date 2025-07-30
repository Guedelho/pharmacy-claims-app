package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"pharmacyclaims/internal/models"

	"github.com/google/uuid"
)

type ServiceInterface interface {
	ValidateClaim(request models.ClaimRequest) error
	SubmitClaim(request models.ClaimRequest) (*models.ClaimResponse, error)
	ReverseClaim(request models.ReversalRequest) (*models.ReversalResponse, error)
}

type HttpHandler struct {
	service ServiceInterface
}

func NewHttpHandler(service ServiceInterface) *HttpHandler {
	return &HttpHandler{service: service}
}

func (h *HttpHandler) SetupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/claim", h.SubmitClaim)
	mux.HandleFunc("/reversal", h.ReverseClaim)
	mux.HandleFunc("/health", h.HealthCheck)

	return mux
}

func (h *HttpHandler) SubmitClaim(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.sendErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed", "Only POST method is allowed")
		return
	}

	var request models.ClaimRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "Invalid JSON format", err.Error())
		return
	}

	if err := h.service.ValidateClaim(request); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	response, err := h.service.SubmitClaim(request)
	if err != nil {
		if err.Error() == "pharmacy with NPI "+request.NPI+" not found" {
			h.sendErrorResponse(w, http.StatusNotFound, "Pharmacy not found", err.Error())
			return
		}
		h.sendErrorResponse(w, http.StatusInternalServerError, "Failed to submit claim", err.Error())
		return
	}

	h.sendJSONResponse(w, http.StatusCreated, response)
}

func (h *HttpHandler) ReverseClaim(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.sendErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed", "Only POST method is allowed")
		return
	}

	var request models.ReversalRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "Invalid JSON format", err.Error())
		return
	}

	if request.ClaimID == uuid.Nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "Invalid claim_id", "claim_id must be a valid UUID")
		return
	}

	response, err := h.service.ReverseClaim(request)
	if err != nil {
		if err.Error() == "claim with ID "+request.ClaimID.String()+" not found" {
			h.sendErrorResponse(w, http.StatusNotFound, "Claim not found", err.Error())
			return
		}
		if err.Error() == "claim is already reversed" {
			h.sendErrorResponse(w, http.StatusConflict, "Claim already reversed", err.Error())
			return
		}
		h.sendErrorResponse(w, http.StatusInternalServerError, "Failed to reverse claim", err.Error())
		return
	}

	h.sendJSONResponse(w, http.StatusOK, response)
}

func (h *HttpHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed", "Only GET method is allowed")
		return
	}

	h.sendJSONResponse(w, http.StatusOK, map[string]string{"status": "healthy"})
}

func (h *HttpHandler) sendJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Failed to encode JSON response: %v", err)
	}
}

func (h *HttpHandler) sendErrorResponse(w http.ResponseWriter, statusCode int, error, message string) {
	response := models.ErrorResponse{
		Error:   error,
		Message: message,
	}

	h.sendJSONResponse(w, statusCode, response)
}
