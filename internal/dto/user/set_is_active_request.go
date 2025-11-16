package user

type SetIsActiveRequest struct {
	UserID   string `json:"user_id" db:"user_id"`
	IsActive bool   `json:"is_active" db:"is_active"`
}
