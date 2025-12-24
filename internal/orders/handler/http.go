package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"HW4/internal/common/httpx"
	"HW4/internal/orders/dto"
	"HW4/internal/orders/service"
)

type Handler struct {
	svc *service.OrdersService
}

func New(svc *service.OrdersService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid json body")
		return
	}
	if req.UserID == "" || req.Amount <= 0 {
		httpx.Error(w, http.StatusBadRequest, "BAD_REQUEST", "user_id required and amount must be > 0")
		return
	}

	resp, err := h.svc.CreateOrder(r.Context(), req)
	if err != nil {
		if err == service.ErrBadRequest {
			httpx.Error(w, http.StatusBadRequest, "BAD_REQUEST", "user_id required and amount must be > 0")
			return
		}
		httpx.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to create order")
		return
	}

	httpx.JSON(w, http.StatusCreated, httpx.SuccessResponse[dto.CreateOrderResponse]{Data: resp})
}

func (h *Handler) ListOrders(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		httpx.Error(w, http.StatusBadRequest, "BAD_REQUEST", "user_id is required")
		return
	}

	resp, err := h.svc.ListOrders(r.Context(), userID)
	if err != nil {
		if err == service.ErrBadRequest {
			httpx.Error(w, http.StatusBadRequest, "BAD_REQUEST", "user_id is required")
			return
		}
		httpx.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to list orders")
		return
	}

	httpx.JSON(w, http.StatusOK, httpx.SuccessResponse[dto.OrdersListResponse]{Data: resp})
}

func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/orders/")
	if id == "" {
		httpx.Error(w, http.StatusBadRequest, "BAD_REQUEST", "order_id is required")
		return
	}

	resp, err := h.svc.GetOrder(r.Context(), id)
	if err != nil {
		if err == service.ErrBadRequest {
			httpx.Error(w, http.StatusBadRequest, "BAD_REQUEST", "order_id is required")
			return
		}
		if err == service.ErrNotFound {
			httpx.Error(w, http.StatusNotFound, "NOT_FOUND", "order not found")
			return
		}
		httpx.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to get order")
		return
	}

	httpx.JSON(w, http.StatusOK, httpx.SuccessResponse[dto.OrderResponse]{Data: resp})
}
