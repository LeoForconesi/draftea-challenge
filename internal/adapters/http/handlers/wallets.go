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
	topUpService interface {
		TopUp(ctx context.Context, req *wallets.TopUpRequest) (*wallets.TopUpResponse, error)
	}
	listService interface {
		ListWallets(ctx context.Context, req *wallets.ListWalletsRequest) (*wallets.ListWalletsResponse, error)
	}
	createService interface {
		CreateWallet(ctx context.Context, req *wallets.CreateWalletRequest) (*wallets.CreateWalletResponse, error)
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
	topUpService interface {
		TopUp(ctx context.Context, req *wallets.TopUpRequest) (*wallets.TopUpResponse, error)
	},
	listService interface {
		ListWallets(ctx context.Context, req *wallets.ListWalletsRequest) (*wallets.ListWalletsResponse, error)
	},
	createService interface {
		CreateWallet(ctx context.Context, req *wallets.CreateWalletRequest) (*wallets.CreateWalletResponse, error)
	},
) *WalletHandler {
	return &WalletHandler{
		balanceService:      balanceService,
		transactionsService: transactionsService,
		topUpService:        topUpService,
		listService:         listService,
		createService:       createService,
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

// TopUp handles POST /wallets/{user_id}/top-up.
func (h *WalletHandler) TopUp(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		presenter.WriteError(c, errors.NewValidationError("invalid user_id", map[string]interface{}{"user_id": c.Param("user_id")}))
		return
	}

	var body struct {
		Amount   int64  `json:"amount"`
		Currency string `json:"currency"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		presenter.WriteError(c, errors.NewValidationError("invalid request body", map[string]interface{}{"error": err.Error()}))
		return
	}

	details := make(map[string]interface{})
	if body.Amount <= 0 {
		details["amount"] = body.Amount
	}
	if body.Currency == "" {
		details["currency"] = "required"
	}
	if len(details) > 0 {
		presenter.WriteError(c, errors.NewValidationError("invalid top-up request", details))
		return
	}

	resp, err := h.topUpService.TopUp(c.Request.Context(), &wallets.TopUpRequest{
		UserID:   userID,
		Amount:   body.Amount,
		Currency: body.Currency,
	})
	if err != nil {
		presenter.WriteError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ListWallets handles GET /wallets.
func (h *WalletHandler) ListWallets(c *gin.Context) {
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

	resp, err := h.listService.ListWallets(c.Request.Context(), &wallets.ListWalletsRequest{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		presenter.WriteError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// CreateWallet handles POST /wallets.
func (h *WalletHandler) CreateWallet(c *gin.Context) {
	var body struct {
		UserID string `json:"user_id"`
		Name   string `json:"name"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		presenter.WriteError(c, errors.NewValidationError("invalid request body", map[string]interface{}{"error": err.Error()}))
		return
	}
	userID, err := uuid.Parse(body.UserID)
	if err != nil {
		presenter.WriteError(c, errors.NewValidationError("invalid user_id", map[string]interface{}{"user_id": body.UserID}))
		return
	}
	if len(body.Name) > 20 {
		presenter.WriteError(c, errors.NewValidationError("invalid wallet name", map[string]interface{}{"name": body.Name}))
		return
	}

	resp, err := h.createService.CreateWallet(c.Request.Context(), &wallets.CreateWalletRequest{
		UserID: userID,
		Name:   body.Name,
	})
	if err != nil {
		presenter.WriteError(c, err)
		return
	}

	c.JSON(http.StatusCreated, resp)
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
