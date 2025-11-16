package entity

import "time"

type PRStatus string

const (
	StatusOpen   PRStatus = "OPEN"
	StatusMerged PRStatus = "MERGED"
)

type PullRequest struct {
	PullRequestID     string   `db:"pull_request_id"`
	Name              string   `db:"pull_request_name"`
	AuthorID          string   `db:"author_id"`
	Status            PRStatus `db:"status"`
	AssignedReviewers []string
	CreatedAt         *time.Time `db:"created_at"`
	MergedAt          *time.Time `db:"merged_at"`
}
