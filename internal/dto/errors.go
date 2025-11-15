package dto

import "errors"

var (
	ErrNotFound    = errors.New("not found")
	ErrTeamExists  = errors.New("team already exists")
	ErrUserExists  = errors.New("user already exists")
	ErrPRExists    = errors.New("pull request already exists")
	ErrPRMerged    = errors.New("pull request already merged")
	ErrNotAssigned = errors.New("reviewer not assigned")
	ErrNoCandidate = errors.New("no candidate available")
)

type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}
