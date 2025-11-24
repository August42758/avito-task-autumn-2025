package repository

import "errors"

var (
	ErrDuplicatedTeamName = errors.New("duplicated team_name")
	ErrDuplicatedPRid     = errors.New("duplicated pull_request_id")
	ErrDuplicatedUserId   = errors.New("duplicated user_id")
	ErrNoRecord           = errors.New("no rows after query")
)
