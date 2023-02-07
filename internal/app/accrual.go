package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/size12/gophermart/accrualsystem"
	"github.com/size12/gophermart/internal/entity"
	"github.com/size12/gophermart/internal/storage"
)

type WorkerPool struct {
	jobs    chan entity.Order
	accrual accrualsystem.AccrualSystem
	storage storage.Storage
}

func (w *WorkerPool) StartWorker() {
	go func() {
		for {
			work := <-w.jobs

			newOrderInfo, sleep, err := w.accrual.GetOrderUpdates(work)
			_ = sleep
			if err != nil {
				log.Println("Failed get update order info:", err)
				w.storage.PushFrontOrders(work)
			}

			if newOrderInfo.Status != work.Status {
				//fmt.Println("Status changed")
				work.Accrual = newOrderInfo.Accrual
				work.Status = newOrderInfo.Status
				w.storage.UpdateOrders(context.Background(), work)
			} else {
				w.storage.PushBackOrders(work)
			}
		}
	}()
}

func NewWorkerPool(ctx context.Context, s storage.Storage, accrual accrualsystem.AccrualSystem) {
	pool := WorkerPool{
		jobs:    make(chan entity.Order),
		storage: s,
		accrual: accrual,
	}

	pool.StartWorker()

	for {
		job, err := s.GetOrderForUpdate()

		if errors.Is(err, storage.ErrEmptyQueue) {
			time.Sleep(1 * time.Second)
			continue
		}

		if err != nil {
			log.Println("Failed get order for update")
			return
		}

		select {
		case pool.jobs <- job:
			fmt.Println("Sent job to worker:", job)
		case <-ctx.Done():
			fmt.Println("Shutdown")
			return
		}
	}

}
