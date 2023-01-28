package accrual_system

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/size12/gophermart/internal/config"
	"github.com/size12/gophermart/internal/models"
)

type ExAccrualSystem struct {
	BaseURL string
}

func NewExAccrualSystem(cfg config.Config) *ExAccrualSystem {
	return &ExAccrualSystem{BaseURL: cfg.AccrualSystemAddress}
}

func (s *ExAccrualSystem) GetOrderUpdates(number int) (models.Order, error) {
	order := models.Order{}

	path := "/api/orders/"
	url := fmt.Sprintf("%s%s%v", s.BaseURL, path, number)

	r, err := http.Get(url)
	if err != nil {
		log.Println("Can't get order updates from external API:", err)
		return order, err
	}

	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		log.Println("Can't read response body:", err)
		return order, err
	}

	err = json.Unmarshal(body, &order)

	if err != nil {
		log.Println("Can't unmarshal response body:", err)
		return order, err
	}

	return order, nil
}
