package handlers

import (
	"context"
	"net/http"
	"strconv"

	"draftea-challenge/internal/adapters/http/presenter"
	"draftea-challenge/internal/application/wallets"
	"draftea-challenge/internal/domain/errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// WalletHandler handles wallet endpoints.
type WalletHandler struct {
	balanceService interface {
		GetBalance(ctx context.Context, userID uuid.UUID) (*wallets.GetBalanceResponse, error)
	}
	transactionsService interface {
		GetTransactions(ctx context.Context, req *wallets.GetTransactionsRequest) (*wallets.GetTransactionsResponse, error)
	}
}

// NewWalletHandler creates a WalletHandler.
func NewWalletHandler(
	balanceService interface {
		GetBalance(ctx context.Context, userID uuid.UUID) (*wallets.GetBalanceResponse, error)
	},
	transactionsService interface {
		GetTransactions(ctx context.Context, req *wallets.GetTransactionsRequest) (*wallets.GetTransactionsResponse, error)
	},
) *WalletHandler {
	return &WalletHandler{
		balanceService:      balanceService,
		transactionsService: transactionsService,
	}
}

// GetBalance handles GET /wallets/{user_id}/balance.
func (h *WalletHandler) GetBalance(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		presenter.WriteError(c, errors.NewValidationError("invalid user_id", map[string]interface{}{"user_id": c.Param("user_id")}))
		return
	}

	resp, err := h.balanceService.GetBalance(c.Request.Context(), userID)
	if err != nil {
		presenter.WriteError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ListTransactions handles GET /wallets/{user_id}/transactions.
func (h *WalletHandler) ListTransactions(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		presenter.WriteError(c, errors.NewValidationError("invalid user_id", map[string]interface{}{"user_id": c.Param("user_id")}))
		return
	}

	limit, limitErr := parseIntQuery(c, "limit", 20)
	offset, offsetErr := parseIntQuery(c, "offset", 0)
	if limitErr != nil || offsetErr != nil || limit < 0 || offset < 0 {
		details := make(map[string]interface{})
		if limitErr != nil || limit < 0 {
			details["limit"] = c.Query("limit")
		}
		if offsetErr != nil || offset < 0 {
			details["offset"] = c.Query("offset")
		}
		presenter.WriteError(c, errors.NewValidationError("invalid pagination params", details))
		return
	}

	req := &wallets.GetTransactionsRequest{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	}

	resp, err := h.transactionsService.GetTransactions(c.Request.Context(), req)
	if err != nil {
		presenter.WriteError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func parseIntQuery(c *gin.Context, key string, fallback int) (int, error) {
	val := c.Query(key)
	if val == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(val)
	if err != nil {
		return fallback, err
	}
	return parsed, nil
}
