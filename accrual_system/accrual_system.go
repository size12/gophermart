package accrual_system

import (
	"github.com/size12/gophermart/internal/config"
	"github.com/size12/gophermart/internal/models"
)

type AccrualSystem interface {
	GetOrderUpdates(number int) (models.Order, error)
}

func NewAccrualSystem(cfg config.Config) AccrualSystem {
	return NewExAccrualSystem(cfg)
}
