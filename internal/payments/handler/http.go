package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"HW4/internal/common/httpx"
	"HW4/internal/payments/dto"
	"HW4/internal/payments/service"
)

type Handler struct {
	svc *service.PaymentsService
}

func New(svc *service.PaymentsService) *Handler { return &Handler{svc: svc} }

func (h *Handler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid json body")
		return
	}

	err := h.svc.CreateAccount(r.Context(), req)
	if err != nil {
		if err == service.ErrBadRequest {
			httpx.Error(w, http.StatusBadRequest, "BAD_REQUEST", "user_id required and balance must be >= 0")
			return
		}
		if err == service.ErrAlreadyExists {
			httpx.Error(w, http.StatusConflict, "ALREADY_EXISTS", "account already exists")
			return
		}
		httpx.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to create account")
		return
	}

	httpx.JSON(w, http.StatusCreated, httpx.SuccessResponse[map[string]any]{Data: map[string]any{"status": "created"}})
}

func (h *Handler) TopUp(w http.ResponseWriter, r *http.Request) {
	var req dto.TopUpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid json body")
		return
	}

	err := h.svc.TopUp(r.Context(), req)
	if err != nil {
		if err == service.ErrBadRequest {
			httpx.Error(w, http.StatusBadRequest, "BAD_REQUEST", "user_id required and amount must be > 0")
			return
		}
		httpx.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to top up")
		return
	}

	httpx.JSON(w, http.StatusOK, httpx.SuccessResponse[map[string]any]{Data: map[string]any{"status": "ok"}})
}

func (h *Handler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID := strings.TrimPrefix(r.URL.Path, "/accounts/")

	resp, err := h.svc.GetBalance(r.Context(), userID)
	if err != nil {
		if err == service.ErrBadRequest {
			httpx.Error(w, http.StatusBadRequest, "BAD_REQUEST", "user_id is required")
			return
		}
		if err == service.ErrNotFound {
			httpx.Error(w, http.StatusNotFound, "NOT_FOUND", "account not found")
			return
		}
		httpx.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to get balance")
		return
	}

	httpx.JSON(w, http.StatusOK, httpx.SuccessResponse[dto.BalanceResponse]{Data: resp})
}
