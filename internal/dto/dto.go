package dto

import (
	"encoding/json"
	"pr-service/internal/enums"
	"time"
)

type EmbendedError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponseDTO struct {
	Error EmbendedError `json:"error"`
}

func NewErrorResponseDTO(code, message string) *ErrorResponseDTO {
	return &ErrorResponseDTO{
		Error: EmbendedError{
			Code:    code,
			Message: message,
		},
	}
}

// TODO: убрать панику на нормальную проверку
func (e *ErrorResponseDTO) ToString() string {
	b, err := json.MarshalIndent(e, "", "    ")
	if err != nil {
		panic(err)
	}

	return string(b)
}

type TeamMemberDTO struct {
	UserId   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type TeamDTO struct {
	TeamName string           `json:"team_name"`
	Members  []*TeamMemberDTO `json:"members"`
}

type ResponseTeamDTO struct {
	Team *TeamDTO `json:"team"`
}

type User struct {
	UserId   string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

type UserDTO struct {
	User *User `json:"user"`
}

type IsActiveUserDTO struct {
	Id       string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

type RequestPullrequestDTO struct {
	PullRequestId   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
}

type PullrequestDTO struct {
	PullRequestId     string   `json:"pull_request_id"`
	PullRequestName   string   `json:"pull_request_name"`
	AuthorID          string   `json:"author_id"`
	Status            string   `json:"status"`
	AssignedReviewers []string `json:"assigned_reviewers"`
}

type ResponsePullrequestDTO struct {
	PR *PullrequestDTO `json:"pr"`
}

func NewPullRequestDTO(pullRequestId string, pullRequestName string, authorID string, assignedReviewers ...string) *PullrequestDTO {
	dto := &PullrequestDTO{
		PullRequestId:   pullRequestId,
		PullRequestName: pullRequestName,
		AuthorID:        authorID,
		Status:          enums.OPEN,
	}

	//максимум два ревьювера
	dto.AssignedReviewers = make([]string, 0, 2)

	dto.AssignedReviewers = append(dto.AssignedReviewers, assignedReviewers...)

	return dto
}

type MergedPullRequestDTO struct {
	PullRequestId     string    `json:"pull_request_id"`
	PullRequestName   string    `json:"pull_request_name"`
	AuthorID          string    `json:"author_id"`
	Status            string    `json:"status"`
	AssignedReviewers []string  `json:"assigned_reviewers"`
	MergedAt          time.Time `json:"mergedAt"`
}

type ResponseMergedPullRequestDTO struct {
	PR *MergedPullRequestDTO `json:"pr"`
}

type PullRequestIdDTO struct {
	PullRequestId string `json:"pull_request_id"`
}

type RequestReassignDTO struct {
	PullRequestId string `json:"pull_request_id"`
	OldUserId     string `json:"old_reviewer_id"`
}

type ResponseReassignDTO struct {
	PR         *PullrequestDTO `json:"pr"`
	ReplacedBy string          `json:"replaced_by"`
}

type UserPullRequestDTO struct {
	PullRequestId   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
	Status          string `json:"status"`
}

type UserPullRequestsDTO struct {
	UserId       string                `json:"user_id"`
	PullRequests []*UserPullRequestDTO `json:"pull_requests"`
}

type StatsResponseDTO struct {
	TotalPRs          int            `json:"total_prs"`
	PRsByStatus       map[string]int `json:"pr_by_status"`
	AssignmentsByUser map[string]int `json:"assignments_by_user"`
}
