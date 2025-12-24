package dto

type PaymentRequested struct {
	MessageID string `json:"message_id"`
	OrderID   string `json:"order_id"`
	UserID    string `json:"user_id"`
	Amount    int64  `json:"amount"`
	CreatedAt string `json:"created_at"`
}
