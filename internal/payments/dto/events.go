package dto

type PaymentResult struct {
	MessageID string `json:"message_id"`
	OrderID   string `json:"order_id"`
	UserID    string `json:"user_id"`
	Amount    int64  `json:"amount"`
	Status    string `json:"status"`
	Reason    string `json:"reason,omitempty"`
	CreatedAt string `json:"created_at"`
}
