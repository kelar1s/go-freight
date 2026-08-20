package rest

import (
	"encoding/json"
	"net/http"

	"github.com/kelar1s/go-freight/internal/inventory/transport/rest/dto"
)

func WriteError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	errResp := dto.ErrorResponse{
		Error: message,
	}
	_ = json.NewEncoder(w).Encode(errResp)
}
