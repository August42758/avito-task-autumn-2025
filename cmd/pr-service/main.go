package main

import (
	"log/slog"
	"net/http"
	"os"

	"pr-service/internal/config"
	"pr-service/internal/database"
	"pr-service/internal/handlers"
	"pr-service/internal/repository"
	"pr-service/internal/routes.go"
	"pr-service/internal/service"
	"pr-service/internal/validators"
)

func main() {
	// создаем логгер
	lgr := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	// загрузка .env
	cfg, err := config.Load("./.env")
	if err != nil {
		lgr.With(
			slog.String("error", err.Error()),
		).Error("Failed to load config from .env file")
		return
	}

	// получаем адресс  базы данных
	dbAddr := database.GetDbAddres(cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBUser, cfg.DBPassword)

	// проверяем наличие миграций
	if err := database.RunMigrations(dbAddr); err != nil {
		lgr.With(
			slog.String("error", err.Error()),
		).Error("Failed to migrate")
		return
	}

	// подключение к БД
	db, err := database.ConnectDB(dbAddr)
	if err != nil {
		lgr.With(
			slog.String("error", err.Error()),
		).Error("Failed to connect to DB")
		return
	}
	defer func() {
		if err := database.CloseDD(db); err != nil {
			lgr.With(
				slog.String("error", err.Error()),
			).Error("Failed to close DB")
		}
	}()

	// создаем репозитории
	usersRepository := &repository.UsersRepository{Db: db}
	teamsRepository := &repository.TeamsRepository{Db: db}
	reviewersRepository := &repository.ReviewersRepository{Db: db}
	pullRequestsRepository := &repository.PullRequestsRepository{Db: db}

	// создаем сервисы
	usersService := &service.UsersService{
		UsersRepository:        usersRepository,
		ReviewersRepository:    reviewersRepository,
		PullRequestsRepository: pullRequestsRepository,
		Lgr:                    lgr,
	}

	teamsService := &service.TeamsService{
		UsersRepository: usersRepository,
		TeamsRepository: teamsRepository,
		Lgr:             lgr,
	}

	pullRequestsService := &service.PullRequestsService{
		UsersRepository:        usersRepository,
		ReviewersRepository:    reviewersRepository,
		PullRequestsRepository: pullRequestsRepository,
		Lgr:                    lgr,
	}

	statsService := &service.StatsService{
		UsersRepository:        usersRepository,
		ReviewersRepository:    reviewersRepository,
		PullRequestsRepository: pullRequestsRepository,
	}

	validator := validators.NewValidator()

	// создаем handlers
	usersHandler := &handlers.UsersHandlers{
		UserService: usersService,
		Validator:   validator,
	}

	teamsHandler := &handlers.TeamsHandlers{
		TeamService: teamsService,
		Validator:   validator,
	}

	pullRequestsHandler := &handlers.PullRequestsHandlers{
		PullRequestService: pullRequestsService,
		Validator:          validator,
	}

	statsHandler := &handlers.StatsHandlers{
		StatsService: statsService,
	}

	// создаем роутер
	router := routes.NewRouter(teamsHandler, usersHandler, pullRequestsHandler, statsHandler)

	lgr.Info("Server initialization was passed successfully")

	if err := http.ListenAndServe(":8080", router); err != nil {
		lgr.With(
			slog.String("error", err.Error()),
		).Error("Error during the server's listening")
		return
	}
}
