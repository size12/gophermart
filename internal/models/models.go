package models

import "time"

type User struct {
	ID        int
	Login     string `json:"login"`
	Password  string `json:"password"`
	Cookie    string
	Balance   int
	Withdrawn int
}

type Order struct {
	UserID int
	Number string
	Status string
	Time   time.Time
}
