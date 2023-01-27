package storage

import "errors"

var ErrNotFound = errors.New("not found")
var ErrLoginExists = errors.New("login already exists")
