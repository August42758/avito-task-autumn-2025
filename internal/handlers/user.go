package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"pr-service/internal/dto"
	"pr-service/internal/helpers"
	"pr-service/internal/service"
	"pr-service/internal/validators"
)

type IUsersHandlers interface {
	SetIsActive(w http.ResponseWriter, r *http.Request)
	GetReview(w http.ResponseWriter, r *http.Request)
}

type UsersHandlers struct {
	UserService service.IUsersService
}

func (uh *UsersHandlers) SetIsActive(w http.ResponseWriter, r *http.Request) {
	//проверяем POST метод
	if r.Method != http.MethodPost {
		helpers.WriteErrorReponse(w, http.StatusMethodNotAllowed, "WRONG_METHOD", errWrongMethod.Error())
		return
	}

	//читаем тело запроса
	var requestDTO dto.IsActiveUserDTO
	if err := json.NewDecoder(r.Body).Decode(&requestDTO); err != nil {
		helpers.WriteErrorReponse(w, http.StatusInternalServerError, "SERVER_ERROR", errInternalServer.Error())
		return
	}

	//валидация
	validator := validators.NewTeamMemberValidator()
	validator.ValidateId(requestDTO.Id)
	if !validator.IsValid {
		helpers.WriteErrorReponse(w, http.StatusBadRequest, "WRONG_DATA_INPUT", errWrongDataInput.Error())
		return
	}

	responseDTO, err := uh.UserService.SetIsActiveById(&requestDTO)
	if err != nil {

		//если пользователя не существует
		if errors.Is(err, service.ErrNoResourse) {
			helpers.WriteErrorReponse(w, http.StatusNotFound, "NOT_FOUND", errNotFound.Error())
			return
		}

		//если ошибка после работы сервисного слоя со стороны сервера
		helpers.WriteErrorReponse(w, http.StatusInternalServerError, "SERVER_ERROR", errInternalServer.Error())
		return
	}

	//сереализируем тело ответа
	helpers.WriteSuccessfulResponse(w, http.StatusOK, responseDTO)
}

func (uh *UsersHandlers) GetReview(w http.ResponseWriter, r *http.Request) {
	//проверяем GET метод
	if r.Method != http.MethodGet {
		helpers.WriteErrorReponse(w, http.StatusMethodNotAllowed, "WRONG_METHOD", errWrongMethod.Error())
		return
	}

	//проверяем наличие квери параметра
	userId := r.URL.Query().Get("user_id")
	if userId == "" {
		helpers.WriteErrorReponse(w, http.StatusBadRequest, "MISSING_PARAM", errMissingParam.Error())
		return
	}

	//сервисная логика
	responseDTO, err := uh.UserService.GetPullRequestsByUserId(userId)
	if err != nil {

		//если пользователя не существует
		if errors.Is(err, service.ErrNoResourse) {
			helpers.WriteErrorReponse(w, http.StatusNotFound, "NOT_FOUND", errNotFound.Error())
			return
		}

		//если ошибка после работы сервисного слоя со стороны сервера
		helpers.WriteErrorReponse(w, http.StatusInternalServerError, "SERVER_ERROR", errInternalServer.Error())
		return
	}

	//сереализируем тело ответа
	helpers.WriteSuccessfulResponse(w, http.StatusOK, responseDTO)
}
