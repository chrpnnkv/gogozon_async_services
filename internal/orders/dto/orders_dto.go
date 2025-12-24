package dto

type CreateOrderRequest struct {
	UserID      string `json:"user_id"`
	Amount      int64  `json:"amount"`
	Description string `json:"description"`
}

type CreateOrderResponse struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
}

type OrderResponse struct {
	OrderID     string `json:"order_id"`
	UserID      string `json:"user_id"`
	Amount      int64  `json:"amount"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
	Description string `json:"description"`
}

type OrdersListResponse struct {
	Orders []OrderResponse `json:"orders"`
}
