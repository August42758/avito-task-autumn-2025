package service

import (
	"errors"
	"log/slog"
	"pr-service/internal/dto"
	"pr-service/internal/repository"
)

type IUsersService interface {
	SetIsActiveById(isActiveUserDTO *dto.IsActiveUserDTO) (*dto.UserDTO, error)
	GetPullRequestsByUserId(id string) (*dto.UserPullRequestsDTO, error)
}

type UsersService struct {
	UsersRepository        repository.IUsersRepository
	ReviewersRepository    repository.IReviewersRepository
	PullRequestsRepository repository.IPullRequestsRepository
	Lgr                    *slog.Logger
}

func (us *UsersService) SetIsActiveById(isActiveUserDTO *dto.IsActiveUserDTO) (*dto.UserDTO, error) {
	us.Lgr.Info("starting user active status operation")

	// проверяем наличие пользователя в бд
	user, err := us.UsersRepository.GetUserById(nil, isActiveUserDTO.Id)
	if err != nil {
		us.Lgr.With(
			slog.String("user_id", isActiveUserDTO.Id),
			slog.String("error", err.Error()),
		).Error("user not found")
		if errors.Is(err, repository.ErrNoRecord) {
			return nil, ErrNoResourse
		}
		return nil, err
	}

	// проверяем значение до изменения
	if user.IsActive != isActiveUserDTO.IsActive {
		if err := us.UsersRepository.UpdateUserIsActive(isActiveUserDTO.Id, isActiveUserDTO.IsActive); err != nil {
			us.Lgr.With(
				slog.String("user_id", isActiveUserDTO.Id),
				slog.String("error", err.Error()),
			).Error("failed to update user active status")
			return nil, err
		}
		user.IsActive = isActiveUserDTO.IsActive
	}

	us.Lgr.Info("user active status operation completed")

	return &dto.UserDTO{
		User: &dto.User{
			UserId:   user.Id,
			Username: user.Username,
			TeamName: user.TeamName,
			IsActive: isActiveUserDTO.IsActive,
		},
	}, nil
}

func (us *UsersService) GetPullRequestsByUserId(id string) (*dto.UserPullRequestsDTO, error) {
	us.Lgr.Info("retrieving pull requests for user")

	// проверяем наличие пользователя в бд
	isExists, err := us.UsersRepository.IsExist(nil, id)
	if err != nil {
		us.Lgr.With(
			slog.String("user_id", id),
			slog.String("error", err.Error()),
		).Error("user not found")
	}

	if !isExists {
		us.Lgr.Error("user not found")
		return nil, ErrNoResourse
	}

	// получаем PR, на которые назначен пользователь
	pullRequestIds, err := us.ReviewersRepository.GetPullRequestIDsWithReviewersByUserId(id)
	if err != nil {
		us.Lgr.With(
			slog.String("user_id", id),
			slog.String("error", err.Error()),
		).Error("failed to get user's pull requests")
		return nil, err
	}

	// формируем DTO
	responseDTO := &dto.UserPullRequestsDTO{
		UserId:       id,
		PullRequests: make([]*dto.UserPullRequestDTO, 0, len(pullRequestIds)),
	}
	for _, pullRequestId := range pullRequestIds {
		pullRequestModel, err := us.PullRequestsRepository.GetPullRequestById(pullRequestId)
		if err != nil {
			us.Lgr.With(
				slog.String("pull_request_id", pullRequestId),
				slog.String("error", err.Error()),
			).Error("failed to get pull request")
			return nil, err
		}

		userPullRequestDTO := &dto.UserPullRequestDTO{
			PullRequestId:   pullRequestModel.PullRequestId,
			PullRequestName: pullRequestModel.PullRequestName,
			AuthorID:        pullRequestModel.AuthorID,
			Status:          pullRequestModel.Status,
		}

		responseDTO.PullRequests = append(responseDTO.PullRequests, userPullRequestDTO)
	}

	us.Lgr.Info("user pull requests retrieved successfully")

	return responseDTO, nil
}
