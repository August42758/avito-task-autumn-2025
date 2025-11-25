package handlers

import (
	"io"
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
		helpers.WriteErrorReponse(w, http.StatusMethodNotAllowed, "WRONG_METHOD", errWrongDataInput.Error())
		return
	}

	responseDTO, err := sh.StatsService.GetStats()
	if err != nil {
		if err == io.EOF {
			helpers.WriteErrorReponse(w, http.StatusBadRequest, "EMPTY_BODY", errEmptyBody.Error())
			return
		}

		helpers.WriteErrorReponse(w, http.StatusBadRequest, "WRONG_DATA_INPUT", errWrongDataInput.Error())
		return
	}

	helpers.WriteSuccessfulResponse(w, http.StatusOK, responseDTO)
}
