package test

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"pr-service/internal/dto"
	"pr-service/internal/enums"
	"pr-service/internal/handlers"
	"pr-service/internal/repository"
	"pr-service/internal/service"
	"pr-service/internal/testhelpers"
	"pr-service/internal/testutils"
)

func TestAddPullRequestHandler(t *testing.T) {
	// тестовая база данных на время теста
	db := testutils.NewTestDB(t)
	defer testutils.DeleteDb(t, db)

	// создаем репозитории
	usersRepository := &repository.UsersRepository{Db: db}
	pullRequestsRepository := &repository.PullRequestsRepository{Db: db}
	reviewersRepository := &repository.ReviewersRepository{Db: db}

	lgr := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	// создаем сервис
	pullRequestService := &service.PullRequestsService{
		UsersRepository:        usersRepository,
		PullRequestsRepository: pullRequestsRepository,
		ReviewersRepository:    reviewersRepository,
		Lgr:                    lgr,
	}

	// создаем сам хендлер
	pullRequestHandler := handlers.PullRequestsHandlers{
		PullRequestService: pullRequestService,
	}

	// Предварительно создаем тестовые данные
	testutils.RunQuery(t, db, "./testdata/insertUsers.sql")

	t.Run("successful pull request creation", func(t *testing.T) {
		// формируем запрос
		requestDTO := dto.RequestPullrequestDTO{
			PullRequestId:   "pr-1001",
			PullRequestName: "Add new feature",
			AuthorID:        "u1",
		}

		b, err := json.Marshal(requestDTO)
		if err != nil {
			t.Fatal(err)
		}

		request := httptest.NewRequest("POST", "/pullRequest/create", bytes.NewReader(b))
		responseWriter := httptest.NewRecorder()

		// делаем запрос
		pullRequestHandler.AddPullRequest(responseWriter, request)
		responseResult := responseWriter.Result()

		// код ответа должен совпадать
		testhelpers.Equal(t, responseResult.StatusCode, http.StatusCreated)

		// десереализируем ответ
		var responseDTO dto.ResponsePullrequestDTO
		if err := json.NewDecoder(responseResult.Body).Decode(&responseDTO); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		// ID PR должен совпадать
		testhelpers.Equal(t, responseDTO.PR.PullRequestId, requestDTO.PullRequestId)

		// название PR должно совпадать
		testhelpers.Equal(t, responseDTO.PR.PullRequestName, requestDTO.PullRequestName)

		// автор должен совпадать
		testhelpers.Equal(t, responseDTO.PR.AuthorID, requestDTO.AuthorID)

		// статус должен быть OPEN
		testhelpers.Equal(t, responseDTO.PR.Status, enums.OPEN)

		// проверяем записи в БД

		// проверяем что PR создался в БД
		pr, err := pullRequestsRepository.GetPullRequestById(requestDTO.PullRequestId)
		if err != nil {
			t.Fatalf("Failed to get PR from database: %v", err)
		}

		// PR должен существовать
		testhelpers.Equal(t, pr.PullRequestId, requestDTO.PullRequestId)

		// название PR должно совпадать
		testhelpers.Equal(t, pr.PullRequestName, requestDTO.PullRequestName)

		// автор должен совпадать
		testhelpers.Equal(t, pr.AuthorID, requestDTO.AuthorID)

		// статус должен быть OPEN
		testhelpers.Equal(t, pr.Status, enums.OPEN)

		// проверяем что ревьюверы назначились
		reviewerIds, err := reviewersRepository.GetReviewersIdByPullRequestId(requestDTO.PullRequestId)
		if err != nil {
			t.Fatalf("Failed to get reviewers from database: %v", err)
		}

		// должны быть назначены ревьюверы (2 человека)
		testhelpers.Equal(t, len(reviewerIds), 2)

		//проверяем что автор не назначен себе ревьювером
		for _, reviewerId := range reviewerIds {
			testhelpers.Equal(t, reviewerId != requestDTO.AuthorID, true)
		}
	})

	t.Run("duplicate pull request creation", func(t *testing.T) {
		// формируем запрос
		requestDTO := dto.RequestPullrequestDTO{
			PullRequestId:   "pr-1002",
			PullRequestName: "Duplicate PR test",
			AuthorID:        "u2",
		}

		b, err := json.Marshal(requestDTO)
		if err != nil {
			t.Fatal(err)
		}

		request := httptest.NewRequest("POST", "/pullRequest/create", bytes.NewReader(b))
		responseWriter := httptest.NewRecorder()

		// делаем первый запрос
		pullRequestHandler.AddPullRequest(responseWriter, request)
		responseResult := responseWriter.Result()

		// код ответа должен совпадать
		testhelpers.Equal(t, responseResult.StatusCode, http.StatusCreated)

		// проверяем что PR создался в БД
		_, err = pullRequestsRepository.GetPullRequestById(requestDTO.PullRequestId)
		if err != nil {
			t.Fatalf("Failed to get PR from database: %v", err)
		}

		// создаем дубликат PR
		responseWriter = httptest.NewRecorder()
		request = httptest.NewRequest("POST", "/pullRequest/create", bytes.NewReader(b))

		// делаем повторный запрос
		pullRequestHandler.AddPullRequest(responseWriter, request)
		responseResult = responseWriter.Result()

		// код ответа должен совпадать
		testhelpers.Equal(t, responseResult.StatusCode, http.StatusConflict)

		// десереализируем ответ
		var responseDTO dto.ErrorResponseDTO
		if err := json.NewDecoder(responseResult.Body).Decode(&responseDTO); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		// код ошибки должен совпадать
		testhelpers.Equal(t, responseDTO.Error.Code, "PR_EXISTS")
	})

	t.Run("author not found", func(t *testing.T) {
		// формируем запрос с несуществующим автором
		requestDTO := dto.RequestPullrequestDTO{
			PullRequestId:   "pr-1003",
			PullRequestName: "PR with non-existent author",
			AuthorID:        "u999",
		}

		b, err := json.Marshal(requestDTO)
		if err != nil {
			t.Fatal(err)
		}

		request := httptest.NewRequest("POST", "/pullRequest/create", bytes.NewReader(b))
		responseWriter := httptest.NewRecorder()

		// делаем запрос
		pullRequestHandler.AddPullRequest(responseWriter, request)
		responseResult := responseWriter.Result()

		// код ответа должен совпадать
		testhelpers.Equal(t, responseResult.StatusCode, http.StatusNotFound)

		// десереализируем ответ
		var responseDTO dto.ErrorResponseDTO
		if err := json.NewDecoder(responseResult.Body).Decode(&responseDTO); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		// код ошибки должен совпадать
		testhelpers.Equal(t, responseDTO.Error.Code, "NOT_FOUND")

		// проверяем что PR не создался в БД
		_, err = pullRequestsRepository.GetPullRequestById(requestDTO.PullRequestId)
		testhelpers.Equal(t, errors.Is(err, repository.ErrNoRecord), true)
	})

	t.Run("pull request without available reviewers", func(t *testing.T) {
		// Создаем команду с одним пользователем (автором)
		testutils.RunQuery(t, db, "./testdata/insertSoloTeam.sql")

		// формируем запрос от единственного пользователя
		requestDTO := dto.RequestPullrequestDTO{
			PullRequestId:   "pr-1111",
			PullRequestName: "Solo developer PR",
			AuthorID:        "u0",
		}

		b, err := json.Marshal(requestDTO)
		if err != nil {
			t.Fatal(err)
		}

		request := httptest.NewRequest("POST", "/pullRequest/create", bytes.NewReader(b))
		responseWriter := httptest.NewRecorder()

		// делаем запрос
		pullRequestHandler.AddPullRequest(responseWriter, request)
		responseResult := responseWriter.Result()

		// код ответа должен совпадать (PR создается без ревьюверов)
		testhelpers.Equal(t, responseResult.StatusCode, http.StatusCreated)

		// проверяем что PR создался в БД
		pr, err := pullRequestsRepository.GetPullRequestById(requestDTO.PullRequestId)
		if err != nil {
			t.Fatalf("Failed to get PR from database: %v", err)
		}

		// PR должен существовать
		testhelpers.Equal(t, pr.PullRequestId, requestDTO.PullRequestId)

		// проверяем что ревьюверы не назначены
		reviewerIds, err := reviewersRepository.GetReviewersIdByPullRequestId(requestDTO.PullRequestId)
		if err != nil {
			t.Fatalf("Failed to get reviewers from database: %v", err)
		}

		// не должно быть ревьюверов
		testhelpers.Equal(t, len(reviewerIds), 0)
	})

	t.Run("invalid HTTP method", func(t *testing.T) {
		request := httptest.NewRequest("GET", "/pullRequest/create", nil)
		responseWriter := httptest.NewRecorder()

		pullRequestHandler.AddPullRequest(responseWriter, request)
		responseResult := responseWriter.Result()

		// код ответа должен совпадать
		testhelpers.Equal(t, responseResult.StatusCode, http.StatusMethodNotAllowed)

		// десереализируем ответ
		var responseDTO dto.ErrorResponseDTO
		if err := json.NewDecoder(responseResult.Body).Decode(&responseDTO); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		// код ошибки должен совпадать
		testhelpers.Equal(t, responseDTO.Error.Code, "WRONG_METHOD")
	})
}
