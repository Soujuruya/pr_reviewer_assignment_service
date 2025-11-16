package entity

type Team struct {
	TeamName string `db:"team_name"`
	Members  []User
}
