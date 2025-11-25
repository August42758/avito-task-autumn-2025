package service

import (
	"errors"
	"log/slog"
	"math/rand/v2"
	"time"

	"pr-service/internal/dto"
	"pr-service/internal/enums"
	"pr-service/internal/models"
	"pr-service/internal/repository"
)

type IPullRequestsService interface {
	AddPullRequest(reqPullRequest *dto.RequestPullrequestDTO) (*dto.ResponsePullrequestDTO, error)
	MergePullRequest(id string) (*dto.ResponseMergedPullRequestDTO, error)
	ReassignReviewer(requestReassignDTO *dto.RequestReassignDTO) (*dto.ResponseReassignDTO, error)
}

type PullRequestsService struct {
	PullRequestsRepository repository.IPullRequestsRepository
	UsersRepository        repository.IUsersRepository
	ReviewersRepository    repository.IReviewersRepository
	Lgr                    *slog.Logger
}

func (ps *PullRequestsService) AddPullRequest(reqPullRequest *dto.RequestPullrequestDTO) (*dto.ResponsePullrequestDTO, error) {

	ps.Lgr.Info("starting pull request creation")

	//начало транзакции
	tx, err := ps.PullRequestsRepository.GetDB().Begin()
	if err != nil {
		ps.Lgr.With(
			slog.String("error", err.Error()),
		).Error("failed to begin transaction", slog.String("error", err.Error()))
		return nil, err
	}

	//если возникла ошибка в сервисной функции
	defer func() {
		if err != nil {
			ps.Lgr.With(
				slog.String("error", err.Error()),
			).Error("rolling back transaction due to error")
			if errRollback := tx.Rollback(); errRollback != nil {
				ps.Lgr.With(
					slog.String("error", errRollback.Error()),
				).Error("rolling back was failed")
			}
		}
	}()

	//проверяем наличие автора
	author, err := ps.UsersRepository.GetUserById(tx, reqPullRequest.AuthorID)
	if err != nil {
		ps.Lgr.With(
			slog.String("author_id", reqPullRequest.AuthorID),
			slog.String("error", err.Error()),
		).Error("author not found", slog.String("author_id", reqPullRequest.AuthorID), slog.String("error", err.Error()))
		if errors.Is(err, repository.ErrNoRecord) {
			return nil, ErrNoResourse
		}
		return nil, err
	}

	//пользователи в команде автора
	users, err := ps.UsersRepository.GetUsersByTeam(tx, author.TeamName)
	if err != nil {
		ps.Lgr.With(
			slog.String("team", author.TeamName),
			slog.String("error", err.Error()),
		).Error("failed to get team users", slog.String("team", author.TeamName), slog.String("error", err.Error()))
		return nil, err
	}

	//удаляем автора из пользователей команды и всех неактивынх юзеров
	reviewerList := []*models.UserModel{}
	for i := 0; i != len(users); i++ {
		if users[i].Id == author.Id {
			continue
		}

		if users[i].IsActive {
			reviewerList = append(reviewerList, users[i])
		}
	}

	//добавляем pr
	pullRequestModel := &models.PullRequestModel{
		PullRequestId:   reqPullRequest.PullRequestId,
		PullRequestName: reqPullRequest.PullRequestName,
		AuthorID:        reqPullRequest.AuthorID,
		CreatedAt:       time.Now(),
	}

	if err := ps.PullRequestsRepository.AddPullRequest(tx, pullRequestModel); err != nil {
		ps.Lgr.With(
			slog.String("PullRequestId", pullRequestModel.PullRequestId),
			slog.String("error", err.Error()),
		).Error("failed to add pull request")
		if errors.Is(err, repository.ErrDuplicatedPRid) {
			return nil, ErrPRExists
		}
		return nil, err
	}

	//выбираем id reviewr
	reviewrIds := []string{}
	if len(reviewerList) >= 2 {

		//Выбираем первого ревьювера
		reviewerIndex := rand.IntN(len(reviewerList))
		reviewrIds = append(reviewrIds, reviewerList[reviewerIndex].Id)

		//убираем выбранного ревьювера
		reviewerList = append(reviewerList[:reviewerIndex], reviewerList[reviewerIndex+1:]...)

		//Выбираем второго ревьювера
		reviewerIndex = rand.IntN(len(reviewerList))
		reviewrIds = append(reviewrIds, reviewerList[reviewerIndex].Id)

	} else if len(reviewerList) == 1 {
		reviewerIndex := rand.IntN(len(reviewerList))
		reviewrIds = append(reviewrIds, reviewerList[reviewerIndex].Id)
	}

	//добавляем reviwers
	for _, id := range reviewrIds {
		reviwerModel := &models.ReviewerModel{
			UserId:        id,
			PullRequestId: reqPullRequest.PullRequestId,
		}

		if err := ps.ReviewersRepository.AddReviewer(tx, reviwerModel); err != nil {
			ps.Lgr.With(
				slog.String("userId", reviwerModel.UserId),
				slog.String("error", err.Error()),
			).Error("failed to add reviewers for pull request")
			return nil, err
		}
	}

	responseDTO := &dto.ResponsePullrequestDTO{
		PR: dto.NewPullRequestDTO(reqPullRequest.PullRequestId, reqPullRequest.PullRequestName, reqPullRequest.AuthorID, reviewrIds...),
	}

	//завершаем транзакцию
	err = tx.Commit()
	if err != nil {
		ps.Lgr.With(
			slog.String("error", err.Error()),
		).Error("failed to commit transaction")
		return nil, err
	}

	ps.Lgr.Info("pull request creation completed successfully")

	return responseDTO, nil
}

func (ps *PullRequestsService) MergePullRequest(id string) (*dto.ResponseMergedPullRequestDTO, error) {
	ps.Lgr.Info("starting merge a pull request")

	// проверяем наличие pr
	pullRequestModel, err := ps.PullRequestsRepository.GetPullRequestById(id)
	if err != nil {
		ps.Lgr.With(
			slog.String("error", err.Error()),
		).Error("pull request not found")
		if errors.Is(err, repository.ErrNoRecord) {
			return nil, ErrNoResourse
		}
		return nil, err
	}

	// проверяем наличие ревьюверов у pr
	reviewersIds, err := ps.ReviewersRepository.GetReviewersIdByPullRequestId(id)
	if err != nil {
		ps.Lgr.With(
			slog.String("error", err.Error()),
		).Error("failed to get reviewers list")
		return nil, err
	}

	// если нет ревьювера, то не можем замержить
	if len(reviewersIds) == 0 {
		ps.Lgr.With().Warn("pr doesn't have reviewers")
		return nil, ErrNoReviewrs
	}

	//проверяем статус Pr до обращения к репозиторию
	if pullRequestModel.Status != enums.MERGED {
		mergedAt := time.Now()
		if err := ps.PullRequestsRepository.MergePullRequest(mergedAt, id); err != nil {
			ps.Lgr.With(
				slog.String("error", err.Error()),
			).Error("failed to merge pr")
			return nil, err
		}

		// обновлдяем модель для merged_at
		pullRequestModel, err = ps.PullRequestsRepository.GetPullRequestById(id)
		if err != nil {
			ps.Lgr.With(
				slog.String("error", err.Error()),
			).Error("failed to get merged pr")
			return nil, err
		}
	}

	ps.Lgr.Info("pull request merge operation completed")

	return &dto.ResponseMergedPullRequestDTO{
		PR: &dto.MergedPullRequestDTO{
			PullRequestId:     pullRequestModel.PullRequestId,
			PullRequestName:   pullRequestModel.PullRequestName,
			AuthorID:          pullRequestModel.AuthorID,
			Status:            enums.MERGED,
			AssignedReviewers: reviewersIds,
			MergedAt:          *pullRequestModel.MergedAt,
		},
	}, nil

}

func (ps *PullRequestsService) ReassignReviewer(requestReassignDTO *dto.RequestReassignDTO) (*dto.ResponseReassignDTO, error) {
	ps.Lgr.Info("starting reviewer reassignment")

	// проверяем наличие pr
	pullRequestModel, err := ps.PullRequestsRepository.GetPullRequestById(requestReassignDTO.PullRequestId)
	if err != nil {
		ps.Lgr.With(
			slog.String("error", err.Error()),
		).Error("pull request not found")
		if errors.Is(err, repository.ErrNoRecord) {
			return nil, ErrNoResourse
		}
		return nil, err
	}

	// проверяем наличие юзера
	if _, err := ps.UsersRepository.GetUserById(nil, requestReassignDTO.OldUserId); err != nil {
		ps.Lgr.With(
			slog.String("reviewer_id", requestReassignDTO.OldUserId),
			slog.String("error", err.Error()),
		).Error("old reviewer not found")
		if errors.Is(err, repository.ErrNoRecord) {
			return nil, ErrNoResourse
		}
		return nil, err
	}

	// проверяем статус pr
	if pullRequestModel.Status == enums.MERGED {
		ps.Lgr.Warn("cannot reassign reviewer on merged PR")
		return nil, ErrPrMerged
	}

	// проверяем наличие ревьюверов у pr
	oldReviewersIds, err := ps.ReviewersRepository.GetReviewersIdByPullRequestId(requestReassignDTO.PullRequestId)
	if err != nil {
		ps.Lgr.With(
			slog.String("error", err.Error()),
		).Error("failed to get reviewers")
		return nil, err
	}

	//если у pr вообще не было ревьюверов, то пропускаем проверку старого ревьювера
	prDoesntHaveReviewers := true
	if len(oldReviewersIds) != 0 {
		// проверяем наличие конкретного ревьювера
		flag := false
		for _, reviewerId := range oldReviewersIds {
			if reviewerId == requestReassignDTO.OldUserId {
				flag = true
				break
			}
		}
		if !flag {
			ps.Lgr.With(
				slog.String("reviewer_id", requestReassignDTO.OldUserId),
			).Warn("reviewer is not assigned to this PR")
			return nil, ErrNoSuchReviewer
		}

		//у pr есть ревьюверы
		prDoesntHaveReviewers = false
	}

	// берем автора
	author, err := ps.UsersRepository.GetUserById(nil, pullRequestModel.AuthorID)
	if err != nil {
		ps.Lgr.With(
			slog.String("author_id", pullRequestModel.AuthorID),
			slog.String("error", err.Error()),
		).Error("failed to get author")
		return nil, err
	}

	// пользователи в команде автора
	temaUsers, err := ps.UsersRepository.GetUsersByTeam(nil, author.TeamName)
	if err != nil {
		ps.Lgr.With(
			slog.String("team", author.TeamName),
			slog.String("error", err.Error()),
		).Error("failed to get team users")
		return nil, err
	}

	newReviewersList := []*models.UserModel{}
	for _, teamUser := range temaUsers {

		//пропускаем автора и неактивных юзеров
		if teamUser.Id == author.Id || !teamUser.IsActive {
			continue
		}

		// пропускаем старого юзера, если у PR вообще нету ревьевров
		if !prDoesntHaveReviewers && teamUser.Id == requestReassignDTO.OldUserId {
			continue
		}

		//чтобы повторно не добавили второго ревьювера
		isAssigned := false
		for _, oldReviewerid := range oldReviewersIds {
			if teamUser.Id == oldReviewerid && oldReviewerid != requestReassignDTO.OldUserId {
				isAssigned = true
				break
			}
		}

		if !isAssigned {
			newReviewersList = append(newReviewersList, teamUser)
		}
	}

	// проверяем наличие ревьюверов
	if len(newReviewersList) == 0 {
		ps.Lgr.Warn("no available replacement candidates")
		return nil, ErrNoReviewrsToAssign
	}

	// выбираем новый id reviewer и меняем
	newReviewerID := newReviewersList[rand.IntN(len(newReviewersList))].Id

	//меняем ревьювера, если до этого кто-то да был
	if !prDoesntHaveReviewers {

		if err := ps.ReviewersRepository.ChangeReviewer(pullRequestModel.PullRequestId, requestReassignDTO.OldUserId, newReviewerID); err != nil {
			ps.Lgr.With(
				slog.String("reviewer_id", newReviewerID),
				slog.String("error", err.Error()),
			).Error("failed to change reviewer", slog.String("reviewer_id", newReviewerID), slog.String("error", err.Error()))
			return nil, err
		}

		//если ревьюверов не было
	} else {
		newReviewerModel := &models.ReviewerModel{
			UserId:        newReviewerID,
			PullRequestId: requestReassignDTO.PullRequestId,
		}

		if err := ps.ReviewersRepository.AddReviewer(nil, newReviewerModel); err != nil {
			ps.Lgr.With(
				slog.String("reviewer_id", newReviewerID),
				slog.String("error", err.Error()),
			).Error("failed to change reviewer", slog.String("reviewer_id", newReviewerID), slog.String("error", err.Error()))
			return nil, err
		}
	}

	// получаем новый список ревьюверов
	newReviewerIds, err := ps.ReviewersRepository.GetReviewersIdByPullRequestId(pullRequestModel.PullRequestId)
	if err != nil {
		ps.Lgr.With(
			slog.String("error", err.Error()),
		).Error("failed to get updated reviewers list")
		return nil, err
	}

	ps.Lgr.Info("reviewer reassignment completed successfully")

	responseDTO := &dto.ResponseReassignDTO{
		PR:         dto.NewPullRequestDTO(pullRequestModel.PullRequestId, pullRequestModel.PullRequestName, pullRequestModel.AuthorID, newReviewerIds...),
		ReplacedBy: newReviewerID,
	}

	return responseDTO, nil
}
