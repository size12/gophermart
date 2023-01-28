package app

import (
	"context"
	"log"
	"time"

	"github.com/size12/gophermart/accrualsystem"
	"github.com/size12/gophermart/internal/models"
	"github.com/size12/gophermart/internal/storage"
)

func UpdateOrders(ctx context.Context, s storage.Storage, accrual accrualsystem.AccrualSystem) {
	errCnt := 0
	for {
		select {
		case <-ctx.Done():
			return
		default:
			time.Sleep(2 * time.Second)
			var ordersToUpdate []models.Order
			orders, err := s.GetOrdersForUpdate(ctx)
			if err != nil {
				log.Println("Can't get orders for update:", err)
				errCnt++
				continue
			}

			for _, order := range orders {
				newOrderInfo, err := accrual.GetOrderUpdates(order.Number)
				if err != nil {
					log.Println("Failed get update order info:", err)
					errCnt++
					continue
				}

				if newOrderInfo.Status != order.Status {
					order.Accrual = newOrderInfo.Accrual
					order.Status = newOrderInfo.Status
					ordersToUpdate = append(ordersToUpdate, order)
				}
			}

			err = s.UpdateOrders(ctx, ordersToUpdate)

			if err != nil {
				log.Println("Failed get update orders:", err)
				errCnt++
				continue
			}

			if errCnt > 10 {
				log.Println("Something really wrong with orders updates.")
				panic("Something really wrong with orders updates.")
			}
		}
	}
}
