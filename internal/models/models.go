package models

import (
	"time"
)

type User struct {
	ID        int
	Login     string `json:"login"`
	Password  string `json:"password"`
	Cookie    string
	Balance   int
	Withdrawn int
}

type Order struct {
	UserID    int
	Number    string
	Status    string
	EventTime time.Time
}

type Balance struct {
	Current   int `json:"current"`
	Withdrawn int `json:"withdrawn"`
}

type Withdraw struct {
	Order int       `json:"order,string"`
	Sum   int       `json:"sum"`
	Time  EventTime `json:"processed_at,string"`
}

type EventTime time.Time

func (t EventTime) MarshalJSON() ([]byte, error) {
	newTime := time.Time(t)
	return newTime.MarshalJSON()
}
