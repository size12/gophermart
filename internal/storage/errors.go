package storage

import "errors"

var ErrNotFound = errors.New("not found")
var ErrLoginExists = errors.New("login already exists")
var ErrNoMoney = errors.New("not enough money on balance")
var ErrBadOrderNum = errors.New("wrong order number")
