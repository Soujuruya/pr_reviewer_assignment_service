package entity

type Team struct {
	TeamName string // уникальное имя команды
	Members  []User // список пользователей команды
}
