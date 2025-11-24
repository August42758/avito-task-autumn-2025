package service

import (
	"errors"
	"log/slog"
	"pr-service/internal/dto"
	"pr-service/internal/models"
	"pr-service/internal/repository"
)

type ITeamsService interface {
	AddTeamWithMembers(team *dto.TeamDTO) (*dto.ResponseTeamDTO, error)
	GetTeamWithMembers(teamName string) (*dto.TeamDTO, error)
}

type TeamsService struct {
	TeamsRepository repository.ITeamsRepository
	UsersRepository repository.IUsersRepository
	Lgr             *slog.Logger
}

func (ts *TeamsService) AddTeamWithMembers(team *dto.TeamDTO) (*dto.ResponseTeamDTO, error) {
	ts.Lgr.Info("starting team creation with members")

	// создаем транзакцию
	tx, err := ts.TeamsRepository.GetDB().Begin()
	if err != nil {
		ts.Lgr.With(
			slog.String("error", err.Error()),
		).Error("failed to begin transaction")
		return nil, err
	}

	// если возникла ошибка в сервисной функции
	defer func() {
		if err != nil {
			ts.Lgr.With(
				slog.String("error", err.Error()),
			).Error("rolling back transaction due to error")
			if errRollback := tx.Rollback(); errRollback != nil {
				ts.Lgr.With(
					slog.String("error", errRollback.Error()),
				).Error("rolling back was failed")
			}
		}
	}()

	// добавляем команду
	if err := ts.TeamsRepository.AddTeam(tx, team.TeamName); err != nil {
		ts.Lgr.With(
			slog.String("error", err.Error()),
		).Error("failed to add team")
		if errors.Is(err, repository.ErrDuplicatedTeamName) {
			return nil, ErrTeamExists
		}
		return nil, err
	}

	// добавляем пользователей в команде
	for _, member := range team.Members {
		user := &models.UserModel{
			Id:       member.UserId,
			Username: member.Username,
			IsActive: member.IsActive,
			TeamName: team.TeamName,
		}

		if err := ts.UsersRepository.AddUser(tx, user); err != nil {
			ts.Lgr.With(
				slog.String("error", err.Error()),
				slog.String("user_id", user.Id),
			).Error("failed to add team member")
			if errors.Is(err, repository.ErrDuplicatedUserId) {
				return nil, ErrUserExists
			}
			return nil, err
		}
	}

	// успешно завершаем транзакцию
	err = tx.Commit()
	if err != nil {
		ts.Lgr.With(
			slog.String("error", err.Error()),
		).Error("failed to commit transaction")
		return nil, err
	}

	ts.Lgr.Info("team creation completed successfully")
	return &dto.ResponseTeamDTO{Team: team}, nil
}

func (ts *TeamsService) GetTeamWithMembers(teamName string) (*dto.TeamDTO, error) {
	ts.Lgr.Info("retrieving team with members")

	// проверяем существование команды
	isExists, err := ts.TeamsRepository.IsExist(teamName)
	if err != nil {
		ts.Lgr.With(
			slog.String("error", err.Error()),
		).Error("failed to check team existence")
		return nil, err
	}

	// если команды не существует
	if !isExists {
		ts.Lgr.Error("team not found")
		return nil, ErrNoResourse
	}

	// ищем пользователей данной команды
	userModels, err := ts.UsersRepository.GetUsersByTeam(nil, teamName)
	if err != nil {
		ts.Lgr.With(
			slog.String("error", err.Error()),
		).Error("failed to get team members")
		return nil, err
	}

	// формируем DTO для транспортного слоя
	teamDTO := &dto.TeamDTO{
		TeamName: teamName,
		Members:  make([]*dto.TeamMemberDTO, 0, len(userModels)),
	}

	activeMembers := 0
	for _, user := range userModels {
		member := &dto.TeamMemberDTO{
			UserId:   user.Id,
			Username: user.Username,
			IsActive: user.IsActive,
		}
		teamDTO.Members = append(teamDTO.Members, member)

		if user.IsActive {
			activeMembers++
		}
	}

	ts.Lgr.Info("team retrieved successfully")

	return teamDTO, nil
}
