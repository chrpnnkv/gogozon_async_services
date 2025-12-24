package service

import (
	"context"
	"errors"
	"time"

	"HW4/internal/orders/dto"
	"HW4/internal/orders/repository"
)

var (
	ErrBadRequest = errors.New("bad_request")
	ErrNotFound   = errors.New("not_found")
)

type OrdersService struct {
	repo *repository.OrdersRepo
}

func New(repo *repository.OrdersRepo) *OrdersService {
	return &OrdersService{repo: repo}
}

func (s *OrdersService) CreateOrder(ctx context.Context, req dto.CreateOrderRequest) (dto.CreateOrderResponse, error) {
	if req.UserID == "" || req.Amount <= 0 {
		return dto.CreateOrderResponse{}, ErrBadRequest
	}

	orderID, err := s.repo.CreateOrderWithOutbox(ctx, req.UserID, req.Amount, req.Description)
	if err != nil {
		return dto.CreateOrderResponse{}, err
	}

	return dto.CreateOrderResponse{OrderID: orderID, Status: "NEW"}, nil
}

func (s *OrdersService) ListOrders(ctx context.Context, userID string) (dto.OrdersListResponse, error) {
	if userID == "" {
		return dto.OrdersListResponse{}, ErrBadRequest
	}

	orders, err := s.repo.ListOrdersByUser(ctx, userID)
	if err != nil {
		return dto.OrdersListResponse{}, err
	}

	resp := dto.OrdersListResponse{Orders: make([]dto.OrderResponse, 0, len(orders))}
	for _, o := range orders {
		resp.Orders = append(resp.Orders, dto.OrderResponse{
			OrderID:     o.ID,
			UserID:      o.UserID,
			Amount:      o.Amount,
			Status:      o.Status,
			CreatedAt:   o.CreatedAt.UTC().Format(time.RFC3339Nano),
			Description: o.Description,
		})
	}
	return resp, nil
}

func (s *OrdersService) GetOrder(ctx context.Context, id string) (dto.OrderResponse, error) {
	if id == "" {
		return dto.OrderResponse{}, ErrBadRequest
	}

	o, err := s.repo.GetOrderByID(ctx, id)
	if err != nil {
		if err.Error() == "not_found" {
			return dto.OrderResponse{}, ErrNotFound
		}
		return dto.OrderResponse{}, err
	}

	return dto.OrderResponse{
		OrderID:     o.ID,
		UserID:      o.UserID,
		Amount:      o.Amount,
		Status:      o.Status,
		CreatedAt:   o.CreatedAt.UTC().Format(time.RFC3339Nano),
		Description: o.Description,
	}, nil
}
