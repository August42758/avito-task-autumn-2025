package repository

import (
	"database/sql"
	"fmt"

	"pr-service/internal/models"
)

type IReviewersRepository interface {
	AddReviewer(tx *sql.Tx, reviewer *models.ReviewerModel) error
	GetReviewersIdByPullRequestId(id string) ([]string, error)
	ChangeReviewer(pullRequestId, oldReviewerId, newReviewerId string) error
	GetPullRequestsByUserId(id string) ([]string, error)
	CountAssignmentsByUser() (map[string]int, error)
}

type ReviewersRepository struct {
	Db *sql.DB
}

func (rr *ReviewersRepository) AddReviewer(tx *sql.Tx, reviewer *models.ReviewerModel) error {
	stmt := "INSERT INTO reviewers(user_id, pull_request_id) VALUES($1, $2)"

	var err error
	if tx != nil {
		_, err = tx.Exec(stmt, reviewer.UserId, reviewer.PullRequestId)
	} else {
		_, err = rr.Db.Exec(stmt, reviewer.UserId, reviewer.PullRequestId)
	}

	if err != nil {
		return err
	}

	return nil
}

func (rr *ReviewersRepository) GetReviewersIdByPullRequestId(id string) ([]string, error) {
	stmt := "SELECT user_id FROM reviewers WHERE pull_request_id = $1"

	rows, err := rr.Db.Query(stmt, id)
	if err != nil {
		return nil, err
	}

	defer func() {
		if errClose := rows.Close(); errClose != nil {
			if err == nil {
				err = fmt.Errorf("failed to close database rows: %w", errClose)
			}
		}
	}()

	reviewrIds := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		reviewrIds = append(reviewrIds, id)
	}

	return reviewrIds, nil
}

func (rr *ReviewersRepository) ChangeReviewer(pullRequestId, oldReviewerId, newReviewerId string) error {
	stmt := "UPDATE reviewers SET user_id = $1 WHERE pull_request_id = $2 and user_id = $3"

	if _, err := rr.Db.Exec(stmt, newReviewerId, pullRequestId, oldReviewerId); err != nil {
		return err
	}

	return nil
}

func (rr *ReviewersRepository) GetPullRequestsByUserId(id string) ([]string, error) {
	stmt := `SELECT reviewers.pull_request_id 
    FROM reviewers 
    JOIN pull_requests ON reviewers.pull_request_id = pull_requests.pull_request_id
    WHERE user_id = $1 AND pull_requests.status_id = $2`

	rows, err := rr.Db.Query(stmt, id, 1)
	if err != nil {
		return nil, err
	}

	defer func() {
		if errClose := rows.Close(); errClose != nil {
			if err == nil {
				err = fmt.Errorf("failed to close database rows: %w", errClose)
			}
		}
	}()

	pullRequestIds := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		pullRequestIds = append(pullRequestIds, id)
	}

	return pullRequestIds, nil
}

func (rr *ReviewersRepository) CountAssignmentsByUser() (map[string]int, error) {
	stmt := "SELECT user_id, COUNT(*) FROM reviewers GROUP BY user_id"

	rows, err := rr.Db.Query(stmt)
	if err != nil {
		return nil, err
	}

	defer func() {
		if errClose := rows.Close(); errClose != nil {
			if err == nil {
				err = fmt.Errorf("failed to close database rows: %w", errClose)
			}
		}
	}()

	result := make(map[string]int)
	for rows.Next() {
		var userId string
		var count int
		if err := rows.Scan(&userId, &count); err != nil {
			return nil, err
		}
		result[userId] = count
	}

	return result, nil
}
