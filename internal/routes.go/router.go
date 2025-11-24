package routes

import (
	"net/http"

	"pr-service/internal/handlers"
)

func NewRouter(teamsHandler handlers.ITeamsHandlers, usersHandler handlers.IUsersHandlers, pullRequestsHandler handlers.IPullRequestsHandlers) *http.ServeMux {
	router := http.NewServeMux()

	router.HandleFunc("/team/add", teamsHandler.AddTeam)
	router.HandleFunc("/team/get", teamsHandler.GetTeam)

	router.HandleFunc("/pullRequest/create", pullRequestsHandler.AddPullRequest)
	router.HandleFunc("/pullRequest/merge", pullRequestsHandler.MergePullRequest)
	router.HandleFunc("/pullRequest/reassign", pullRequestsHandler.ReassignReviewer) //не рабоатет так, как должна

	router.HandleFunc("/users/getReview", usersHandler.GetReview) //неясно, что с ней
	router.HandleFunc("/users/setIsActive", usersHandler.SetIsActive)

	return router
}
