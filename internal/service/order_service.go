// order_service.go
package service

import (
	"context"
	"log"
	"sync"

	"github.com/rom4k-vzlom4k/oder-service/internal/repository"
	"github.com/rom4k-vzlom4k/oder-service/pkg/models"
)

type OrderService interface {
	GetByID(ctx context.Context, id string) (*models.Order, error)
	Save(ctx context.Context, order *models.Order) error
	RestoreCache(ctx context.Context) error // восст. кэш
}

type orderService struct {
	repo  repository.OrderRepository
	cache map[string]*models.Order
	mu    sync.RWMutex
}

func NewOrderService(repo repository.OrderRepository) OrderService {
	return &orderService{
		repo:  repo,
		cache: make(map[string]*models.Order),
	}
}

func (s *orderService) GetByID(ctx context.Context, id string) (*models.Order, error) {
	s.mu.RLock()
	order, found := s.cache[id]
	s.mu.RUnlock()

	if found {
		return order, nil
	}

	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if order != nil {
		s.mu.Lock()
		s.cache[order.OrderUID] = order
		s.mu.Unlock()
	}
	return order, nil
}

func (s *orderService) Save(ctx context.Context, order *models.Order) error {
	err := s.repo.Save(ctx, order)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.cache[order.OrderUID] = order
	s.mu.Unlock()
	return nil
}

func (s *orderService) RestoreCache(ctx context.Context) error {
	log.Println("Восстановление кэш из БД loading..")
	orders, err := s.repo.GetAllOrders(ctx)
	if err != nil {
		return err
	}
	s.mu.Lock()
	for _, order := range orders {
		s.cache[order.OrderUID] = order
	}
	s.mu.Unlock()
	log.Printf("Кэш восстановлен. Загружено %d заказов.", len(orders))
	return nil
}
