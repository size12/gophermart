package storage

import "errors"

var (
	ErrNotFound          = errors.New("not found")
	ErrLoginExists       = errors.New("login already exists")
	ErrNoMoney           = errors.New("not enough money on balance")
	ErrBadOrderNum       = errors.New("wrong order number")
	ErrAlreadyLoaded     = errors.New("already loaded order number")
	ErrLoadedByOtherUser = errors.New("already loaded by another user")
	ErrEmptyQueue        = errors.New("queue is empty")
	ErrNothingChanged    = errors.New("nothing changes in order")
)
