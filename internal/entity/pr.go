package entity

import "time"

type PRStatus string

const (
	StatusOpen   PRStatus = "OPEN"
	StatusMerged PRStatus = "MERGED"
)

type PullRequest struct {
	PullRequestID     string // pull_request_id
	Name              string // pull_request_name
	AuthorID          string
	Status            PRStatus
	AssignedReviewers []string // список user_id
	CreatedAt         *time.Time
	MergedAt          *time.Time
}
