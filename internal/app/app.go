package app

import "github.com/size12/gophermart/internal/config"

type App struct {
	Cfg config.Config
}

func NewApp(cfg config.Config) App {
	return App{Cfg: cfg}
}

func (a *App) Run() error {

	//подключение middleware и хэндлеров
	//запуск сервера

	return nil
}
