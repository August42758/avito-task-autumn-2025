package handlers

var (
	WRONG_METHOD     = "This method isn't acceptable"
	SERVER_ERROR     = "Server error"
	MISSING_PARAM    = "Query parameter is missing"
	WRONG_DATA_INPUT = "wrong format of input data"

	TEAM_EXISTS  = "team_name already exists"
	PR_EXISTS    = "PR id already exists"
	NO_REVIEWERS = "PR doesn't have reviewers"
	PR_MERGED    = "cannot reassign on merged PR"
	NOT_ASSIGNED = "reviewer is not assigned to this PR"
	NO_CANDIDATE = "no active replacement candidate in team"
	NOT_FOUND    = "resourse not found"
	USER_EXISTS  = "user_id already exists"
)
