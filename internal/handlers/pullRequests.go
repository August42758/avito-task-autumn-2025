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

type IPullRequestsHandlers interface {
	AddPullRequest(w http.ResponseWriter, r *http.Request)
	MergePullRequest(w http.ResponseWriter, r *http.Request)
	ReassignReviewer(w http.ResponseWriter, r *http.Request)
}

type PullRequestsHandlers struct {
	PullRequestService service.IPullRequestsService
}

func (ph *PullRequestsHandlers) AddPullRequest(w http.ResponseWriter, r *http.Request) {
	//проверяем POST метод
	if r.Method != http.MethodPost {
		helpers.WriteErrorReponse(w, http.StatusMethodNotAllowed, "WRONG_METHOD", WRONG_METHOD)
		return
	}

	//чиатет тело запроса
	var requestDTO dto.RequestPullrequestDTO
	if err := json.NewDecoder(r.Body).Decode(&requestDTO); err != nil {
		helpers.WriteErrorReponse(w, http.StatusInternalServerError, "SERVER_ERROR", SERVER_ERROR)
		return
	}

	//валидация
	validator := validators.NewPullRequestValidator()
	validator.ValidatePullRequestId(requestDTO.PullRequestId)
	validator.ValidatePullRequestName(requestDTO.PullRequestName)
	validator.ValidateAuthorId(requestDTO.AuthorID)
	if !validator.IsValid {
		helpers.WriteErrorReponse(w, http.StatusBadRequest, "WRONG_DATA_INPUT", WRONG_DATA_INPUT)
		return
	}

	//сервисная логика добавления pr
	responseDTO, err := ph.PullRequestService.AddPullRequest(&requestDTO)
	if err != nil {

		// если автор не найден
		if errors.Is(err, service.ErrNoResourse) {
			helpers.WriteErrorReponse(w, http.StatusNotFound, "NOT_FOUND", NOT_FOUND)
			return
		}

		//если pr id уже существует
		if errors.Is(err, service.ErrPRExists) {
			helpers.WriteErrorReponse(w, http.StatusConflict, "PR_EXISTS", PR_EXISTS)
			return
		}

		//если произошла ошибка в процессе сервисной логики
		helpers.WriteErrorReponse(w, http.StatusInternalServerError, "SERVER_ERROR", SERVER_ERROR)
		return
	}

	//сереализируем тело ответа
	helpers.WriteSuccessfulResponse(w, http.StatusCreated, responseDTO)
}

func (ph *PullRequestsHandlers) MergePullRequest(w http.ResponseWriter, r *http.Request) {
	//проверяем POST метод
	if r.Method != http.MethodPost {
		helpers.WriteErrorReponse(w, http.StatusMethodNotAllowed, "WRONG_METHOD", WRONG_METHOD)
		return
	}

	//чиатет тело запроса
	var requestDTO dto.PullRequestIdDTO
	if err := json.NewDecoder(r.Body).Decode(&requestDTO); err != nil {
		helpers.WriteErrorReponse(w, http.StatusInternalServerError, "SERVER_ERROR", SERVER_ERROR)
		return
	}

	//валидация
	validator := validators.NewPullRequestValidator()
	validator.ValidatePullRequestId(requestDTO.PullRequestId)
	if !validator.IsValid {
		helpers.WriteErrorReponse(w, http.StatusBadRequest, "WRONG_DATA_INPUT", WRONG_DATA_INPUT)
		return
	}

	//сервисная логика изменения статуса pr
	responseDTO, err := ph.PullRequestService.MergePullRequest(requestDTO.PullRequestId)
	if err != nil {

		// если pr не найден
		if errors.Is(err, service.ErrNoResourse) {
			helpers.WriteErrorReponse(w, http.StatusNotFound, "NOT_FOUND", NOT_FOUND)
			return
		}

		//если нет ревьеверов уже существует
		if errors.Is(err, service.ErrNoReviewrs) {
			helpers.WriteErrorReponse(w, http.StatusBadRequest, "NO_REVIEWERS", NO_REVIEWERS)
			return
		}

		//если произошла ошибка в процессе сервисной логики
		helpers.WriteErrorReponse(w, http.StatusInternalServerError, "SERVER_ERROR", SERVER_ERROR)
		return
	}

	//сереализируем тело ответа
	helpers.WriteSuccessfulResponse(w, http.StatusOK, responseDTO)
}

func (ph *PullRequestsHandlers) ReassignReviewer(w http.ResponseWriter, r *http.Request) {
	//проверяем POST метод
	if r.Method != http.MethodPost {
		helpers.WriteErrorReponse(w, http.StatusMethodNotAllowed, "WRONG_METHOD", WRONG_METHOD)
		return
	}

	//валидируем тело запроса
	var requestDTO dto.RequestReassignDTO
	if err := json.NewDecoder(r.Body).Decode(&requestDTO); err != nil {
		helpers.WriteErrorReponse(w, http.StatusInternalServerError, "SERVER_ERROR", SERVER_ERROR)
		return
	}

	//валидация
	validator := validators.NewPullRequestValidator()
	validator.ValidatePullRequestId(requestDTO.PullRequestId)
	validator.ValidateAuthorId(requestDTO.OldUserId) // это не автор, но пускай будет так
	if !validator.IsValid {
		helpers.WriteErrorReponse(w, http.StatusBadRequest, "WRONG_DATA_INPUT", WRONG_DATA_INPUT)
		return
	}

	//сервисная логика переназначения ревьювера
	responseDTO, err := ph.PullRequestService.ReassignReviewer(&requestDTO)
	if err != nil {

		//если pr или юзер не найден
		if errors.Is(err, service.ErrNoResourse) {
			helpers.WriteErrorReponse(w, http.StatusNotFound, "NOT_FOUND", NOT_FOUND)
			return
		}

		//если pr уже имеет статус MERGED
		if errors.Is(err, service.ErrPrMerged) {
			helpers.WriteErrorReponse(w, http.StatusConflict, "PR_MERGED", PR_MERGED)
			return
		}

		//если данный пользователь не назначен на данный pr
		if errors.Is(err, service.ErrNoSuchReviewer) {
			helpers.WriteErrorReponse(w, http.StatusConflict, "NOT_ASSIGNED", NOT_ASSIGNED)
			return
		}

		//если нет пользователей на назначение
		if errors.Is(err, service.ErrNoReviewrsToAssign) {
			helpers.WriteErrorReponse(w, http.StatusConflict, "NO_CANDIDATE", NO_CANDIDATE)
			return
		}

		//TODO: ErrNoreviewers не обрабатывается

		//если произошла ошибка в процессе сервисной логики
		helpers.WriteErrorReponse(w, http.StatusInternalServerError, "SERVER_ERROR", SERVER_ERROR)
		return
	}

	//сереализируем тело ответа
	helpers.WriteSuccessfulResponse(w, http.StatusOK, responseDTO)
}
