package handler

import (
	"github.com/gin-gonic/gin"

	dtoreq "github.com/sirawong/simple-banking-api/internal/handler/dto/request"
	dtores "github.com/sirawong/simple-banking-api/internal/handler/dto/response"
	"github.com/sirawong/simple-banking-api/internal/service/account"
	"github.com/sirawong/simple-banking-api/internal/service/transaction"
	_ "github.com/sirawong/simple-banking-api/pkg/errs"
)

type AccountHandler struct {
	accountSvc account.Service
	txSvc      transaction.Service
}

// @wire:set(name=HandlerSet)
func ProvideAccountHandler(accountSvc account.Service, txSvc transaction.Service) *AccountHandler {
	return &AccountHandler{accountSvc: accountSvc, txSvc: txSvc}
}

// CreateAccount godoc
// @Summary      Create account
// @Description  Create a new bank account for a user
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Param        body  body      dtoreq.CreateAccountRequest  true  "Create account request"
// @Success      201   {object}  dtores.AccountResponse
// @Failure      400   {object}  errs.AppError
// @Failure      409   {object}  errs.AppError
// @Failure      500   {object}  errs.AppError
// @Router       /api/v1/accounts [post]
func (h *AccountHandler) CreateAccount(c *gin.Context) {
	var req dtoreq.CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dtores.HandleCreatedResponse(c, nil, err)
		return
	}

	acct, err := h.accountSvc.CreateAccount(c.Request.Context(), req.UserID, req.Currency)
	dtores.HandleCreatedResponse(c, dtores.FromEntityAccount(acct), err)
}

// GetAccount godoc
// @Summary      Get account
// @Description  Get account details including balance
// @Tags         accounts
// @Produce      json
// @Param        id   path      string  true  "Account ID"
// @Success      200  {object}  dtores.BalanceResponse
// @Failure      404  {object}  errs.AppError
// @Router       /api/v1/accounts/{id} [get]
func (h *AccountHandler) GetAccount(c *gin.Context) {
	id := c.Param("id")
	balance, err := h.accountSvc.GetBalance(c.Request.Context(), id)
	if err != nil {
		dtores.HandleResponse(c, nil, err)
		return
	}
	dtores.HandleResponse(c, dtores.BalanceResponse{AccountID: id, Balance: balance}, nil)
}

// ListTransactions godoc
// @Summary      List transactions
// @Description  List transactions for an account with pagination
// @Tags         accounts
// @Produce      json
// @Param        id     path      string  true   "Account ID"
// @Param        page   query     int     false  "Page number"
// @Param        limit  query     int     false  "Page size"
// @Success      200    {object}  dtores.TransactionListResponse
// @Failure      404    {object}  errs.AppError
// @Router       /api/v1/accounts/{id}/transactions [get]
func (h *AccountHandler) ListTransactions(c *gin.Context) {
	id := c.Param("id")
	var req dtoreq.ListTransactionsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		dtores.HandleResponse(c, nil, err)
		return
	}

	txs, total, err := h.txSvc.ListByAccount(c.Request.Context(), id, req.Page, req.Limit)
	if err != nil {
		dtores.HandleResponse(c, nil, err)
		return
	}
	dtores.HandleResponse(c, dtores.TransactionListResponse{
		Transactions: dtores.FromEntityTransactions(txs),
		Total:        total,
		Page:         req.Page,
		Limit:        req.Limit,
	}, nil)
}
