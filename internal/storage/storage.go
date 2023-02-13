package storage

import (
	"context"
	"encoding/hex"
	"math/rand"

	"github.com/size12/gophermart/internal/config"
	"github.com/size12/gophermart/internal/entity"
)

// Storage
// не передаю все данные только в контекстах, так как другому программисту будет сложно понять
// что именно нужно данным методам. А так более понятно выглядит.
type Storage interface {
	GetUser(ctx context.Context, search SearchType, value string) (entity.User, error)
	AddUser(ctx context.Context, user entity.User) (int, error)
	Withdraw(ctx context.Context, user entity.User, order entity.Withdraw) error
	WithdrawalHistory(ctx context.Context, user entity.User) ([]entity.Withdraw, error)
	AddOrder(ctx context.Context, order entity.Order) error
	OrdersHistory(ctx context.Context, user entity.User) ([]entity.Order, error)
	GetOrdersForUpdate(ctx context.Context) ([]entity.Order, error)
	GetOrderForUpdate() (entity.Order, error)
	UpdateOrders(ctx context.Context, orders ...entity.Order) error
	GetConfig() config.Config
	PushFrontOrders(orders ...entity.Order) error
	PushBackOrders(orders ...entity.Order) error
}

func NewStorage(cfg config.Config) (Storage, error) {
	s, err := NewDBStorage(cfg)
	return s, err
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func GenerateRandom() string {
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(randomBytes)[:8]
}

type SearchType string

const (
	SearchByID    SearchType = "id"
	SearchByLogin SearchType = "login"
)
