package service

import "errors"

var (
	ErrNoReviewrs         = errors.New("pr doesn't have reviewers")
	ErrNoSuchReviewer     = errors.New("pr doesn't have such reviewer")
	ErrNoReviewrsToAssign = errors.New("no active replacement candidate in team")
	ErrPrMerged           = errors.New("cannot reassign on merged PR")
	ErrUserExists         = errors.New("user with this id exists")
	ErrTeamExists         = errors.New("team with this id exists")
	ErrPRExists           = errors.New("pr with this id exists")
	ErrNoResourse         = errors.New("resourse doesn't exist")
)
