package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/size12/gophermart/accrualsystem"
	"github.com/size12/gophermart/internal/entity"
	"github.com/size12/gophermart/internal/storage"
)

type JobResult struct {
	Order entity.Order
	Err   error
	Sleep int64
}

type WorkerPool struct {
	jobs    chan entity.Order
	results chan JobResult
	accrual accrualsystem.AccrualSystem
}

func (w *WorkerPool) StartWorker() {
	go func() {
		for {
			work := <-w.jobs

			newOrderInfo, sleep, err := w.accrual.GetOrderUpdates(work)
			if err != nil {
				log.Println("Failed get update order info:", err)
				w.results <- JobResult{
					Order: newOrderInfo,
					Err:   err,
					Sleep: sleep,
				}
			}

			if newOrderInfo.Status != work.Status {
				fmt.Println("Status changed")
				work.Accrual = newOrderInfo.Accrual
				work.Status = newOrderInfo.Status
				w.results <- JobResult{
					Order: newOrderInfo,
					Err:   nil,
					Sleep: sleep,
				}
			} else {
				fmt.Println("Nothing changes")
				w.results <- JobResult{
					Order: newOrderInfo,
					Err:   storage.ErrNothingChanged,
					Sleep: sleep,
				}
			}
		}
	}()
}

func NewWorkerPool(ctx context.Context, s storage.Storage, accrual accrualsystem.AccrualSystem) {
	pool := WorkerPool{
		jobs:    make(chan entity.Order),
		results: make(chan JobResult),
		accrual: accrual,
	}

	pool.StartWorker()

	waitTime := time.Now().UnixMilli()
	needNewJob := true
	job := entity.Order{}
	var err error

	for {
		if needNewJob {
			job, err = s.GetOrderForUpdate()

			if errors.Is(err, storage.ErrEmptyQueue) {
				time.Sleep(1 * time.Second)
				continue
			}

			if err != nil {
				log.Println("Failed get order for update")
				return
			}
		}

		sleep := waitTime - time.Now().UnixMilli()
		if sleep > 0 {
			time.Sleep(time.Duration(sleep))
		}

		select {
		case pool.jobs <- job:
			needNewJob = true
		case result := <-pool.results:
			needNewJob = false
			if result.Err != nil && result.Sleep > 0 {
				fmt.Println("Too many requests")
				atomic.AddInt64(&waitTime, result.Sleep)
				continue
			}
			if errors.Is(result.Err, storage.ErrNothingChanged) {
				s.PushBackOrders(result.Order)
				continue
			}
			if result.Err != nil {
				log.Println("Failed do job:", result.Err)
				s.PushFrontOrders(result.Order)
				continue
			}
			err := s.UpdateOrders(ctx, []entity.Order{result.Order})
			if err != nil {
				log.Println("Failed update orders:", err)
				continue
			}
		case <-ctx.Done():
			fmt.Println("Shutdown")
			return
		}
	}

}
