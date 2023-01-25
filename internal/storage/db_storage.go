package storage

import (
	"database/sql"

	"github.com/size12/gophermart/internal/config"
)

type DBStorage struct {
	Cfg config.Config
	DB  *sql.DB
}

func (s *DBStorage) Read(data string) string {
	//чтение из базы
	return data
}
