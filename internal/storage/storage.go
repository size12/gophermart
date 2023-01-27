package storage

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"

	"github.com/size12/gophermart/internal/config"
	"github.com/size12/gophermart/internal/models"
)

// Storage
// не передаю все данные только в контекстах, так как другому программисту будет сложно понять
// что именно нужно данным методам. А так более понятно выглядит.
type Storage interface {
	GetUser(ctx context.Context, search, value string) (models.User, error)
	AddUser(ctx context.Context, user models.User) (string, error)
	Withdraw(ctx context.Context, user models.User, order models.Withdraw) error
	WithdrawalHistory(ctx context.Context, user models.User) ([]models.Withdraw, error)
	AddOrder(ctx context.Context, order models.Order) error
	OrdersHistory(ctx context.Context, user models.User) ([]models.Order, error)
}

func NewStorage(cfg config.Config) (Storage, error) {
	s, err := NewDBStorage(cfg)
	return s, err
}

func GenerateCookie() (string, error) {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		log.Println("Failed generate cookie:", err)
		return "", err
	}
	return hex.EncodeToString(b), nil
}
