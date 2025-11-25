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
	"pr-service/internal/models"
	"pr-service/internal/repository"
	"pr-service/internal/service"
	"pr-service/internal/testhelpers"
	"pr-service/internal/testutils"
	"pr-service/internal/validators"
)

func TestAddTeamHandler(t *testing.T) {
	// тестовая база данных на время теста
	db := testutils.NewTestDB(t)
	defer testutils.DeleteDb(t, db)

	// создаем репозитории
	usersRepository := &repository.UsersRepository{Db: db}
	teamsRepository := &repository.TeamsRepository{Db: db}

	lgr := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	// создаем сервис
	teamService := &service.TeamsService{
		UsersRepository: usersRepository,
		TeamsRepository: teamsRepository,
		Lgr:             lgr,
	}

	validator := validators.NewValidator()

	// создаем сам хендлер
	teamHandler := handlers.TeamsHandlers{
		TeamService: teamService,
		Validator:   validator,
	}

	t.Run("successful team creation", func(t *testing.T) {
		// формируем запрос
		requestDTO := dto.TeamDTO{
			TeamName: "backend",
			Members: []*dto.TeamMemberDTO{
				{UserId: "u1", Username: "Alice", IsActive: true},
				{UserId: "u2", Username: "Bob", IsActive: true},
			},
		}

		b, err := json.Marshal(requestDTO)
		if err != nil {
			t.Fatal(err)
		}

		request := httptest.NewRequest("POST", "/team/add", bytes.NewReader(b))
		responseWriter := httptest.NewRecorder()

		// делаем запрос
		teamHandler.AddTeam(responseWriter, request)
		responseResult := responseWriter.Result()

		// код ответа должен совпадать
		testhelpers.Equal(t, responseResult.StatusCode, http.StatusCreated)

		// десереализируем ответ
		var responseDTO dto.ResponseTeamDTO
		if err := json.NewDecoder(responseResult.Body).Decode(&responseDTO); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		// название команды ответа должно совпадать
		testhelpers.Equal(t, responseDTO.Team.TeamName, requestDTO.TeamName)

		// кол-во участников должно совпадать
		testhelpers.Equal(t, len(responseDTO.Team.Members), len(requestDTO.Members))

		// проверяем записи в БД
		exists, err := teamsRepository.IsExist(requestDTO.TeamName)
		if err != nil {
			t.Fatalf("Failed to check team existence: %v", err)
		}

		// команда должна существовать
		testhelpers.Equal(t, exists, true)

		// проверяем, что пользователи создались в БД
		userModels, err := usersRepository.GetUsersByTeam(nil, requestDTO.TeamName)
		if err != nil {
			t.Fatalf("Failed to get team userModels: %v", err)
		}

		// кол-во участников должно совпадать
		testhelpers.Equal(t, len(userModels), len(requestDTO.Members))

		// проверяем данные каждого пользователя
		userMap := make(map[string]*models.UserModel)
		for _, user := range userModels {
			userMap[user.Id] = user
		}

		for _, expectedMember := range requestDTO.Members {
			user, exists := userMap[expectedMember.UserId]
			if !exists {
				t.Errorf("User %s was not created in database", expectedMember.UserId)
				continue
			}

			// пользователь должен существовать
			testhelpers.Equal(t, exists, true)

			// кол-во участников должно совпадать
			testhelpers.Equal(t, len(userModels), len(requestDTO.Members))

			// имена должны совпадать
			testhelpers.Equal(t, user.Username, expectedMember.Username)

			// активность должна совпадать
			testhelpers.Equal(t, user.IsActive, expectedMember.IsActive)

			// команда должна совпадать
			testhelpers.Equal(t, user.TeamName, requestDTO.TeamName)
		}
	})

	t.Run("duplicate team creation", func(t *testing.T) {
		// формируем запрос
		requestDTO := dto.TeamDTO{
			TeamName: "frontend",
			Members: []*dto.TeamMemberDTO{
				{UserId: "u3", Username: "Charlie", IsActive: true},
			},
		}

		b, err := json.Marshal(requestDTO)
		if err != nil {
			t.Fatal(err)
		}

		request := httptest.NewRequest("POST", "/team/add", bytes.NewReader(b))
		responseWriter := httptest.NewRecorder()

		// делаем запрос
		teamHandler.AddTeam(responseWriter, request)
		responseResult := responseWriter.Result()

		// код ответа должен совпадать
		testhelpers.Equal(t, responseResult.StatusCode, http.StatusCreated)

		// проверяем записи в БД после первого создания
		exists, err := teamsRepository.IsExist(requestDTO.TeamName)
		if err != nil {
			t.Fatalf("Failed to check team existence: %v", err)
		}

		// команда должна существовать
		testhelpers.Equal(t, exists, true)

		// проверяем, что пользователи создались в БД
		userModels, err := usersRepository.GetUsersByTeam(nil, requestDTO.TeamName)
		if err != nil {
			t.Fatalf("Failed to get team userModels: %v", err)
		}

		// кол-во участников должно совпадать
		testhelpers.Equal(t, len(userModels), len(requestDTO.Members))

		// создаем дубликат команды
		responseWriter = httptest.NewRecorder()
		request = httptest.NewRequest("POST", "/team/add", bytes.NewReader(b))

		// делаем повторный запрос
		teamHandler.AddTeam(responseWriter, request)
		responseResult = responseWriter.Result()

		// код ответа должен совпадать
		testhelpers.Equal(t, responseResult.StatusCode, http.StatusBadRequest)

		// десереализируем ответ
		var responseDTO dto.ErrorResponseDTO
		if err := json.NewDecoder(responseResult.Body).Decode(&responseDTO); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		// код ошибки должен совпадать
		testhelpers.Equal(t, responseDTO.Error.Code, "TEAM_EXISTS")

		// проверяем что данные не дублировались в БД
		userModelsAfterSecondAttempt, err := usersRepository.GetUsersByTeam(nil, requestDTO.TeamName)
		if err != nil {
			t.Fatalf("Failed to get team userModels after duplicate attempt: %v", err)
		}

		// кол-во участников должно остаться прежним
		testhelpers.Equal(t, len(userModelsAfterSecondAttempt), len(requestDTO.Members))
	})

	t.Run("duplicate user in same team", func(t *testing.T) {
		requestDTO := dto.TeamDTO{
			TeamName: "mobile",
			Members: []*dto.TeamMemberDTO{
				{UserId: "u4", Username: "David", IsActive: true},
				{UserId: "u4", Username: "David Duplicate", IsActive: true},
			},
		}
		b, err := json.Marshal(requestDTO)
		if err != nil {
			t.Fatal(err)
		}

		request := httptest.NewRequest("POST", "/team/add", bytes.NewReader(b))
		responseWriter := httptest.NewRecorder()

		teamHandler.AddTeam(responseWriter, request)
		responseResult := responseWriter.Result()

		// код ответа должен совпадать
		testhelpers.Equal(t, responseResult.StatusCode, http.StatusBadRequest)

		// десереализируем ответ
		var responseDTO dto.ErrorResponseDTO
		if err := json.NewDecoder(responseResult.Body).Decode(&responseDTO); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		// код ошибки должен совпадать
		testhelpers.Equal(t, responseDTO.Error.Code, "USER_EXISTS")

		// проверяем что ничего не создалось в БД (транзакция откатилась)
		exists, err := teamsRepository.IsExist(requestDTO.TeamName)
		if err != nil {
			t.Fatalf("Failed to check team existence: %v", err)
		}

		// команда не должна существовать
		testhelpers.Equal(t, exists, false)

		// проверяем, что пользователи не создались в БД
		userModels, err := usersRepository.GetUsersByTeam(nil, requestDTO.TeamName)
		if err != nil {
			t.Fatalf("Failed to get team userModels: %v", err)
		}

		// кол-во участников должно быть 0
		testhelpers.Equal(t, len(userModels), 0)
	})

	t.Run("invalid HTTP method", func(t *testing.T) {
		request := httptest.NewRequest("GET", "/team/add", nil)
		responseWriter := httptest.NewRecorder()

		teamHandler.AddTeam(responseWriter, request)
		responseResult := responseWriter.Result()

		// код ответа должен совпадать
		testhelpers.Equal(t, responseResult.StatusCode, http.StatusMethodNotAllowed)
	})
}
