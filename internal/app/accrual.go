package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/size12/gophermart/internal/accrualsystem"
	"github.com/size12/gophermart/internal/entity"
	"github.com/size12/gophermart/internal/storage"
)

type Timer struct {
	Time time.Time
	*sync.RWMutex
}

type WorkerPool struct {
	jobs    chan entity.Order
	accrual accrualsystem.AccrualSystem
	storage storage.Storage
	timer   Timer
}

func (w *WorkerPool) StartWorker(ctx context.Context) {
	go func() {
		for {
			select {
			case work := <-w.jobs:
				{
					w.timer.RLock()
					timer := w.timer.Time
					t := time.Until(timer)
					w.timer.RUnlock()

					if t.Milliseconds() > 0 {
						time.Sleep(t)
					}

					newOrderInfo, sleep, err := w.accrual.GetOrderUpdates(work)
					if err != nil {
						log.Println("Failed get update order info:", err)
						err := w.storage.PushFrontOrders([]entity.Order{work})
						if err != nil {
							log.Println("Failed push order in queue: ", err)
						}
						if sleep > 0 {
							w.timer.Lock()
							w.timer.Time = time.Now().Add(time.Duration(sleep) * time.Second)
							w.timer.Unlock()
						}
						continue
					}

					if newOrderInfo.Status != work.Status {
						work.Accrual = newOrderInfo.Accrual
						work.Status = newOrderInfo.Status
						err := w.storage.UpdateOrders(context.Background(), work)
						if err != nil {
							log.Println("Failed update order: ", err)
						}
					} else {
						err := w.storage.PushBackOrder(work)
						if err != nil {
							log.Println("Failed push order in queue: ", err)
						}
					}
				}
			case <-ctx.Done():
				log.Println("Shutdown worker")
				return
			}

		}
	}()
}

func NewWorkerPool(ctx context.Context, s storage.Storage, accrual accrualsystem.AccrualSystem) {
	pool := WorkerPool{
		jobs:    make(chan entity.Order),
		storage: s,
		accrual: accrual,
		timer: Timer{
			Time:    time.Now(),
			RWMutex: &sync.RWMutex{},
		},
	}

	cfg := s.GetConfig()

	for i := 0; i < cfg.WorkersCount; i++ {
		pool.StartWorker(ctx)
	}

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
