package dto

type CreateAccountRequest struct {
	UserID  string `json:"user_id"`
	Balance int64  `json:"balance"`
}

type TopUpRequest struct {
	UserID string `json:"user_id"`
	Amount int64  `json:"amount"`
}

type BalanceResponse struct {
	UserID  string `json:"user_id"`
	Balance int64  `json:"balance"`
}
