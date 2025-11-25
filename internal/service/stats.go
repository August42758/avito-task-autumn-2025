package service

import (
	"log/slog"
	"pr-service/internal/dto"
	"pr-service/internal/repository"
)

type IStatsService interface {
	GetStats() (*dto.StatsResponseDTO, error)
}

type StatsService struct {
	PullRequestsRepository *repository.PullRequestsRepository
	ReviewersRepository    *repository.ReviewersRepository
	UsersRepository        *repository.UsersRepository
	Lgr                    *slog.Logger
}

func (ss *StatsService) GetStats() (*dto.StatsResponseDTO, error) {
	ss.Lgr.Info("Start to collect common stats")

	totalPRs, err := ss.PullRequestsRepository.CountAllPullRequests()
	if err != nil {
		ss.Lgr.With(
			slog.String("error", err.Error()),
		).Error("failed to count all prs")
		return nil, err
	}

	prsByStatus, err := ss.PullRequestsRepository.CountPullRequestsByStatus()
	if err != nil {
		ss.Lgr.With(
			slog.String("error", err.Error()),
		).Error("failed to count pr by status")
		return nil, err
	}

	assignmentsByUser, err := ss.ReviewersRepository.CountAssignmentsByUser()
	if err != nil {
		ss.Lgr.With(
			slog.String("error", err.Error()),
		).Error("failed to count assignments by user")
		return nil, err
	}

	ss.Lgr.Info("Collecting complete")

	return &dto.StatsResponseDTO{
		TotalPRs:          totalPRs,
		PRsByStatus:       prsByStatus,
		AssignmentsByUser: assignmentsByUser,
	}, nil
}
