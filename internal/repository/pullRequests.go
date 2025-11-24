package repository

import (
	"database/sql"
	"errors"
	"time"

	"pr-service/internal/models"

	"github.com/lib/pq"
)

type IPullRequestsRepository interface {
	GetDB() *sql.DB
	AddPullRequest(tx *sql.Tx, pullRequest *models.PullRequestModel) error
	MergePullRequest(mergedAt time.Time, id string) error
	GetPullRequestById(id string) (*models.PullRequestModel, error)
}

type PullRequestsRepository struct {
	Db *sql.DB
}

func (pr *PullRequestsRepository) AddPullRequest(tx *sql.Tx, pullRequest *models.PullRequestModel) error {
	stmt := "INSERT INTO pull_requests(pull_request_id, pull_request_name, author_id, created_at, status_id) VALUES($1, $2, $3, $4, $5)"

	var err error
	if tx != nil {
		_, err = tx.Exec(stmt, pullRequest.PullRequestId, pullRequest.PullRequestName, pullRequest.AuthorID, pullRequest.CreatedAt, 1) // 1 - OPEN статус
	} else {
		_, err = pr.Db.Exec(stmt, pullRequest.PullRequestId, pullRequest.PullRequestName, pullRequest.AuthorID, pullRequest.CreatedAt, 1)
	}

	//ошибка во время операции или из-за дубликата id pr
	if err != nil {
		var sqlError *pq.Error
		if errors.As(err, &sqlError) {
			if sqlError.Code == "23505" {
				return ErrDuplicatedPRid
			}
		}
		return err
	}

	return nil

}

func (pr *PullRequestsRepository) MergePullRequest(mergedAt time.Time, id string) error {
	stmt := "UPDATE pull_requests SET status_id = $1, merged_at = $2  WHERE pull_request_id = $3"

	//2 - статус MERGED
	if _, err := pr.Db.Exec(stmt, 2, mergedAt, id); err != nil {
		return err
	}

	return nil
}

func (pr *PullRequestsRepository) GetPullRequestById(id string) (*models.PullRequestModel, error) {
	stmt := `SELECT pull_request_id, pull_request_name, author_id, merged_at, pull_requests_status.status
	FROM pull_requests 
	JOIN pull_requests_status 
	ON pull_requests.status_id = pull_requests_status.pr_status_id 
	WHERE pull_request_id = $1`

	model := &models.PullRequestModel{}
	if err := pr.Db.QueryRow(stmt, id).Scan(&model.PullRequestId, &model.PullRequestName, &model.AuthorID, &model.MergedAt, &model.Status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		}

		return nil, err
	}

	return model, nil
}

func (pr *PullRequestsRepository) GetDB() *sql.DB {
	return pr.Db
}
