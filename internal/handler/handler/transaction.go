package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sirawong/simple-banking-api/internal/handler/dto"
	"github.com/sirawong/simple-banking-api/internal/handler/response"
	transaction "github.com/sirawong/simple-banking-api/internal/service/transaction"
)

type TransactionHandler struct {
	txSvc transaction.Service
}

// @wire:set(name=HandlerSet)
func ProvideTransactionHandler(txSvc transaction.Service) *TransactionHandler {
	return &TransactionHandler{txSvc: txSvc}
}

// Deposit godoc
// @Summary      Deposit
// @Description  Deposit money into an account
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Param        body  body      dto.DepositRequest  true  "Deposit request"
// @Success      200   {object}  errs.AppError
// @Failure      400   {object}  errs.AppError
// @Failure      404   {object}  errs.AppError
// @Router       /api/v1/transactions/deposit [post]
func (h *TransactionHandler) Deposit(c *gin.Context) {
	var req dto.DepositRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleResponse(c, nil, err)
		return
	}

	tx, err := h.txSvc.Deposit(c.Request.Context(), req.AccountID, req.Amount)
	response.HandleResponse(c, tx, err)
}

// Withdraw godoc
// @Summary      Withdraw
// @Description  Withdraw money from an account
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Param        body  body      dto.WithdrawRequest  true  "Withdraw request"
// @Success      200   {object}  errs.AppError
// @Failure      400   {object}  errs.AppError
// @Failure      404   {object}  errs.AppError
// @Failure      422   {object}  errs.AppError
// @Router       /api/v1/transactions/withdraw [post]
func (h *TransactionHandler) Withdraw(c *gin.Context) {
	var req dto.WithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleResponse(c, nil, err)
		return
	}

	tx, err := h.txSvc.Withdraw(c.Request.Context(), req.AccountID, req.Amount)
	response.HandleResponse(c, tx, err)
}

// Transfer godoc
// @Summary      Transfer
// @Description  Transfer money between accounts
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Param        body  body      dto.TransferRequest  true  "Transfer request"
// @Success      200   {object}  errs.AppError
// @Failure      400   {object}  errs.AppError
// @Failure      404   {object}  errs.AppError
// @Failure      422   {object}  errs.AppError
// @Router       /api/v1/transactions/transfer [post]
func (h *TransactionHandler) Transfer(c *gin.Context) {
	var req dto.TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleResponse(c, nil, err)
		return
	}

	tx, err := h.txSvc.Transfer(c.Request.Context(), req.FromAccountID, req.ToAccountID, req.Amount)
	response.HandleResponse(c, tx, err)
}
