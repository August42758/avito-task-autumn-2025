package models

import "time"

type PullRequestModel struct {
	PullRequestId   string
	PullRequestName string
	AuthorID        string
	Status          string
	CreatedAt       time.Time
	MergedAt        *time.Time
}
