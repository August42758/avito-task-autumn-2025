package test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"pr-service/internal/dto"
	"pr-service/internal/handlers"
	"pr-service/internal/repository"
	"pr-service/internal/service"
	"pr-service/internal/testhelpers"
	"pr-service/internal/testutils"
	"pr-service/internal/validators"
)

func TestSetIsActiveHandler(t *testing.T) {
	// тестовая база данных на время теста
	db := testutils.NewTestDB(t)
	defer testutils.DeleteDb(t, db)

	// создаем репозитории
	usersRepository := &repository.UsersRepository{Db: db}
	reviewersRepository := &repository.ReviewersRepository{Db: db}
	pullRequestsRepository := &repository.PullRequestsRepository{Db: db}

	lgr := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	// создаем сервис
	userService := &service.UsersService{
		UsersRepository:        usersRepository,
		ReviewersRepository:    reviewersRepository,
		PullRequestsRepository: pullRequestsRepository,
		Lgr:                    lgr,
	}

	validator := validators.NewValidator()

	// создаем сам хендлер
	userHandler := handlers.UsersHandlers{
		UserService: userService,
		Validator:   validator,
	}

	// Предварительно создаем команду и пользователя для тестов
	testutils.RunQuery(t, db, "./testdata/insertUsers.sql")

	t.Run("successful user activation", func(t *testing.T) {
		// формируем запрос
		requestDTO := dto.IsActiveUserDTO{
			Id:       "u1",
			IsActive: false, // Деактивируем пользователя
		}

		b, err := json.Marshal(requestDTO)
		if err != nil {
			t.Fatal(err)
		}

		request := httptest.NewRequest("POST", "/users/setIsActive", bytes.NewReader(b))
		responseWriter := httptest.NewRecorder()

		// делаем запрос
		userHandler.SetIsActive(responseWriter, request)
		responseResult := responseWriter.Result()

		// код ответа должен совпадать
		testhelpers.Equal(t, responseResult.StatusCode, http.StatusOK)

		// десереализируем ответ
		var responseDTO dto.UserDTO
		if err := json.NewDecoder(responseResult.Body).Decode(&responseDTO); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		// ID пользователя должен совпадать
		testhelpers.Equal(t, responseDTO.User.UserId, requestDTO.Id)

		// статус активности должен совпадать
		testhelpers.Equal(t, responseDTO.User.IsActive, requestDTO.IsActive)

		// проверяем запись в БД
		user, err := usersRepository.GetUserById(nil, requestDTO.Id)
		if err != nil {
			t.Fatalf("Failed to get user from database: %v", err)
		}

		// пользователь должен существовать
		testhelpers.Equal(t, user.Id, "u1")

		// активность должна совпадать с запрошенной
		testhelpers.Equal(t, user.IsActive, requestDTO.IsActive)
	})

	t.Run("successful user deactivation", func(t *testing.T) {
		// формируем запрос
		requestDTO := dto.IsActiveUserDTO{
			Id:       "u2",
			IsActive: true, // Активируем пользователя
		}

		b, err := json.Marshal(requestDTO)
		if err != nil {
			t.Fatal(err)
		}

		request := httptest.NewRequest("POST", "/users/setIsActive", bytes.NewReader(b))
		responseWriter := httptest.NewRecorder()

		// делаем запрос
		userHandler.SetIsActive(responseWriter, request)
		responseResult := responseWriter.Result()

		// код ответа должен совпадать
		testhelpers.Equal(t, responseResult.StatusCode, http.StatusOK)

		// десереализируем ответ
		var responseDTO dto.UserDTO
		if err := json.NewDecoder(responseResult.Body).Decode(&responseDTO); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		// ID пользователя должен совпадать
		testhelpers.Equal(t, responseDTO.User.UserId, requestDTO.Id)

		// статус активности должен совпадать
		testhelpers.Equal(t, responseDTO.User.IsActive, requestDTO.IsActive)

		// проверяем запись в БД
		user, err := usersRepository.GetUserById(nil, "u2")
		if err != nil {
			t.Fatalf("Failed to get user from database: %v", err)
		}

		// пользователь должен существовать
		testhelpers.Equal(t, user.Id, "u2")

		// активность должна совпадать с запрошенной
		testhelpers.Equal(t, user.IsActive, requestDTO.IsActive)
	})

	t.Run("user not found", func(t *testing.T) {
		// формируем запрос для несуществующего пользователя
		requestDTO := dto.IsActiveUserDTO{
			Id:       "u999",
			IsActive: true,
		}

		b, err := json.Marshal(requestDTO)
		if err != nil {
			t.Fatal(err)
		}

		request := httptest.NewRequest("POST", "/users/setIsActive", bytes.NewReader(b))
		responseWriter := httptest.NewRecorder()

		// делаем запрос
		userHandler.SetIsActive(responseWriter, request)
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
	})

	t.Run("no change in active status", func(t *testing.T) {
		// Сначала проверяем текущий статус пользователя
		userId := "u3"

		user, err := usersRepository.GetUserById(nil, userId)
		if err != nil {
			t.Fatalf("Failed to get user from database: %v", err)
		}

		currentStatus := user.IsActive

		// формируем запрос с тем же статусом
		requestDTO := dto.IsActiveUserDTO{
			Id:       userId,
			IsActive: currentStatus, // Тот же статус
		}

		b, err := json.Marshal(requestDTO)
		if err != nil {
			t.Fatal(err)
		}

		request := httptest.NewRequest("POST", "/users/setIsActive", bytes.NewReader(b))
		responseWriter := httptest.NewRecorder()

		// делаем запрос
		userHandler.SetIsActive(responseWriter, request)
		responseResult := responseWriter.Result()

		// код ответа должен быть OK, то есть идемпотентным
		testhelpers.Equal(t, responseResult.StatusCode, http.StatusOK)

		// десереализируем ответ
		var responseDTO dto.UserDTO
		if err := json.NewDecoder(responseResult.Body).Decode(&responseDTO); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		// статус активности должен остаться прежним
		testhelpers.Equal(t, responseDTO.User.IsActive, currentStatus)

		// проверяем что в БД ничего не изменилось
		userAfter, err := usersRepository.GetUserById(nil, userId)
		if err != nil {
			t.Fatalf("Failed to get user from database: %v", err)
		}

		// активность должна остаться прежней
		testhelpers.Equal(t, userAfter.IsActive, currentStatus)
	})

	t.Run("invalid HTTP method", func(t *testing.T) {
		request := httptest.NewRequest("GET", "/users/setIsActive", nil)
		responseWriter := httptest.NewRecorder()

		userHandler.SetIsActive(responseWriter, request)
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
