package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"pr-service/internal/dto"
	"pr-service/internal/helpers"
	"pr-service/internal/service"
	"pr-service/internal/validators"
)

type ITeamsHandlers interface {
	AddTeam(w http.ResponseWriter, r *http.Request)
	GetTeam(w http.ResponseWriter, r *http.Request)
}

type TeamsHandlers struct {
	TeamService service.ITeamsService
}

func (th *TeamsHandlers) AddTeam(w http.ResponseWriter, r *http.Request) {
	// проверяем POST метод
	if r.Method != http.MethodPost {
		helpers.WriteErrorReponse(w, http.StatusMethodNotAllowed, "WRONG_METHOD", errWrongMethod.Error())
		return
	}

	// чиатет тело запроса
	var requestDTO dto.TeamDTO
	if err := json.NewDecoder(r.Body).Decode(&requestDTO); err != nil {
		if err == io.EOF {
			helpers.WriteErrorReponse(w, http.StatusBadRequest, "EMPTY_BODY", errEmptyBody.Error())
			return
		}

		helpers.WriteErrorReponse(w, http.StatusBadRequest, "WRONG_DATA_INPUT", errWrongDataInput.Error())
		return
	}

	// валидация

	validator := validators.NewValidator()

	validator.ValidateTeamName(requestDTO.TeamName)

	for _, member := range requestDTO.Members {
		validator.ValidateUserId(member.UserId)
		validator.ValidateUsername(member.Username)
	}

	if !validator.GetIsValid() {
		helpers.WriteErrorReponse(w, http.StatusBadRequest, "WRONG_DATA_INPUT", errWrongDataInput.Error())
		return
	}

	// испольуем сервисный слой, чтобы добавить команду и ее пользователей
	responseDTO, err := th.TeamService.AddTeamWithMembers(&requestDTO)
	if err != nil {
		// если команда уже существует
		if errors.Is(err, service.ErrTeamExists) {
			helpers.WriteErrorReponse(w, http.StatusBadRequest, "TEAM_EXISTS", errTeamExists.Error())
			return
		}

		// если пользователь уже существует
		if errors.Is(err, service.ErrUserExists) {
			helpers.WriteErrorReponse(w, http.StatusBadRequest, "USER_EXISTS", errUserExists.Error())
			return
		}

		// если ошибка после работы сервисного слоя со стороны сервера
		helpers.WriteErrorReponse(w, http.StatusInternalServerError, "SERVER_ERROR", errInternalServer.Error())
		return
	}

	// сереализируем тело ответа
	helpers.WriteSuccessfulResponse(w, http.StatusCreated, responseDTO)
}

func (th *TeamsHandlers) GetTeam(w http.ResponseWriter, r *http.Request) {
	// проверяем GET метод
	if r.Method != http.MethodGet {
		helpers.WriteErrorReponse(w, http.StatusMethodNotAllowed, "WRONG_METHOD", errWrongMethod.Error())
		return
	}

	// проверяем наличие квери параметра
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		helpers.WriteErrorReponse(w, http.StatusBadRequest, "MISSING_PARAM", errMissingParam.Error())
		return
	}

	responseDTO, err := th.TeamService.GetTeamWithMembers(teamName)
	if err != nil {
		// если команда не существует
		if errors.Is(err, service.ErrNoResourse) {
			helpers.WriteErrorReponse(w, http.StatusNotFound, "NOT_FOUND", errNotFound.Error())
			return
		}

		// если ошибка после работы сервисного слоя со стороны сервера
		helpers.WriteErrorReponse(w, http.StatusInternalServerError, "SERVER_ERROR", errInternalServer.Error())
		return
	}

	helpers.WriteSuccessfulResponse(w, http.StatusOK, responseDTO)
}
