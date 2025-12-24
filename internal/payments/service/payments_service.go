package service

import (
	"context"
	"errors"

	"HW4/internal/payments/dto"
	"HW4/internal/payments/repository"
)

var (
	ErrBadRequest    = errors.New("bad_request")
	ErrAlreadyExists = errors.New("already_exists")
	ErrNotFound      = errors.New("not_found")
)

type PaymentsService struct {
	repo *repository.AccountsRepo
}

func New(repo *repository.AccountsRepo) *PaymentsService {
	return &PaymentsService{repo: repo}
}

func (s *PaymentsService) CreateAccount(ctx context.Context, req dto.CreateAccountRequest) error {
	if req.UserID == "" || req.Balance < 0 {
		return ErrBadRequest
	}
	if err := s.repo.Create(ctx, req.UserID, req.Balance); err != nil {
		return ErrAlreadyExists
	}
	return nil
}

func (s *PaymentsService) TopUp(ctx context.Context, req dto.TopUpRequest) error {
	if req.UserID == "" || req.Amount <= 0 {
		return ErrBadRequest
	}
	return s.repo.TopUp(ctx, req.UserID, req.Amount)
}

func (s *PaymentsService) GetBalance(ctx context.Context, userID string) (dto.BalanceResponse, error) {
	if userID == "" {
		return dto.BalanceResponse{}, ErrBadRequest
	}
	b, err := s.repo.GetBalance(ctx, userID)
	if err != nil {
		// тут repo отдаёт sql.ErrNoRows наверх — handler решает. Можно оставить так:
		return dto.BalanceResponse{}, ErrNotFound
	}
	return dto.BalanceResponse{UserID: userID, Balance: b}, nil
}
