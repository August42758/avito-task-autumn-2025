package service

import (
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
}

func (ss *StatsService) GetStats() (*dto.StatsResponseDTO, error) {

	totalPRs, err := ss.PullRequestsRepository.CountAllPullRequests()
	if err != nil {
		return nil, err
	}

	prsByStatus, err := ss.PullRequestsRepository.CountPullRequestsByStatus()
	if err != nil {
		return nil, err
	}

	assignmentsByUser, err := ss.ReviewersRepository.CountAssignmentsByUser()
	if err != nil {
		return nil, err
	}

	return &dto.StatsResponseDTO{
		TotalPRs:          totalPRs,
		PRsByStatus:       prsByStatus,
		AssignmentsByUser: assignmentsByUser,
	}, nil
}
