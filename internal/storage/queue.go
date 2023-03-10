package storage

import (
	"sync"

	"github.com/size12/gophermart/internal/entity"
)

type Queue interface {
	PushFrontOrders(orders []entity.Order) error
	PushBackOrder(order entity.Order) error
	GetOrder() (entity.Order, error)
}

type SliceQueue struct {
	Orders []entity.Order
	*sync.RWMutex
}

func NewSliceQueue() *SliceQueue {
	return &SliceQueue{
		RWMutex: &sync.RWMutex{},
	}
}

func (q *SliceQueue) PushFrontOrders(orders []entity.Order) error {
	q.Lock()
	defer q.Unlock()
	q.Orders = append(orders, q.Orders...)
	return nil
}

func (q *SliceQueue) PushBackOrder(order entity.Order) error {
	q.Lock()
	defer q.Unlock()
	q.Orders = append(q.Orders, order)
	return nil
}

func (q *SliceQueue) GetOrder() (entity.Order, error) {
	q.RLock()
	defer q.RUnlock()
	if len(q.Orders) > 0 {
		order := q.Orders[0]
		q.Orders = q.Orders[1:]
		return order, nil
	}
	return entity.Order{}, ErrEmptyQueue
}
