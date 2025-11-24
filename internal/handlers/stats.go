package handlers

import (
	"net/http"

	"pr-service/internal/helpers"
	"pr-service/internal/service"
)

type IStatsHandlers interface {
	GetStats(w http.ResponseWriter, r *http.Request)
}

type StatsHandlers struct {
	StatsService service.IStatsService
}

func (sh *StatsHandlers) GetStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helpers.WriteErrorReponse(w, http.StatusMethodNotAllowed, "WRONG_METHOD", WRONG_METHOD)
		return
	}

	responseDTO, err := sh.StatsService.GetStats()
	if err != nil {
		helpers.WriteErrorReponse(w, http.StatusInternalServerError, "SERVER_ERROR", SERVER_ERROR)
		return
	}

	helpers.WriteSuccessfulResponse(w, http.StatusOK, responseDTO)
}
