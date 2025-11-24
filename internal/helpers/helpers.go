package helpers

import (
	"encoding/json"
	"net/http"

	"pr-service/internal/dto"
)

func WriteErrorReponse(w http.ResponseWriter, statusCode int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	errorDTO := dto.NewErrorResponseDTO(code, message)
	json.NewEncoder(w).Encode(errorDTO)
}

func WriteSuccessfulResponse(w http.ResponseWriter, statusCode int, reponseDTO any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(reponseDTO)
}
