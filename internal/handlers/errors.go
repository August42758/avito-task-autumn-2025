package handlers

import "errors"

var (
	errWrongMethod    = errors.New("this method isn't acceptable")
	errInternalServer = errors.New("server error")
	errMissingParam   = errors.New("query parameter is missing")
	errWrongDataInput = errors.New("wrong format of input data")

	errTeamExists  = errors.New("team_name already exists")
	errPrExists    = errors.New("PR id already exists")
	errNoReviewrs  = errors.New("PR doesn't have reviewers")
	errPrMerged    = errors.New("cannot reassign on merged PR")
	errNotAssigned = errors.New("reviewer is not assigned to this PR")
	errNoCandidate = errors.New("no active replacement candidate in team")
	errNotFound    = errors.New("resourse not found")
	errUserExists  = errors.New("user_id already exists")
	errEmptyBody   = errors.New("Request body is empty")
)
