package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/sirawong/simple-banking-api/internal/utils"

	dtoreq "github.com/sirawong/simple-banking-api/internal/handler/dto/request"
	dtores "github.com/sirawong/simple-banking-api/internal/handler/dto/response"
	"github.com/sirawong/simple-banking-api/internal/service/transaction"
	_ "github.com/sirawong/simple-banking-api/pkg/errs"
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
// @Param        body  body      dtoreq.DepositRequest  true  "Deposit request"
// @Success      200   {object}  dtores.TransactionResponse
// @Failure      400   {object}  errs.AppError
// @Failure      403   {object}  errs.AppError
// @Failure      404   {object}  errs.AppError
// @Security     BearerAuth
// @Router       /api/v1/transactions/deposit [post]
func (h *TransactionHandler) Deposit(c *gin.Context) {
	authUser, ok := utils.MustGetAuthUser(c)
	if !ok {
		return
	}

	var req dtoreq.DepositRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dtores.HandleResponse(c, nil, err)
		return
	}

	tx, err := h.txSvc.Deposit(c.Request.Context(), authUser.ID, req.AccountNumber, req.Amount)
	dtores.HandleResponse(c, dtores.FromEntityTransaction(tx), err)
}

// Withdraw godoc
// @Summary      Withdraw
// @Description  Withdraw money from an account
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Param        body  body      dtoreq.WithdrawRequest  true  "Withdraw request"
// @Success      200   {object}  dtores.TransactionResponse
// @Failure      400   {object}  errs.AppError
// @Failure      403   {object}  errs.AppError
// @Failure      404   {object}  errs.AppError
// @Failure      422   {object}  errs.AppError
// @Security     BearerAuth
// @Router       /api/v1/transactions/withdraw [post]
func (h *TransactionHandler) Withdraw(c *gin.Context) {
	authUser, ok := utils.MustGetAuthUser(c)
	if !ok {
		return
	}

	var req dtoreq.WithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dtores.HandleResponse(c, nil, err)
		return
	}

	tx, err := h.txSvc.Withdraw(c.Request.Context(), authUser.ID, req.AccountNumber, req.Amount)
	dtores.HandleResponse(c, dtores.FromEntityTransaction(tx), err)
}

// Transfer godoc
// @Summary      Transfer
// @Description  Transfer money between accounts
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Param        body  body      dtoreq.TransferRequest  true  "Transfer request"
// @Success      200   {object}  dtores.TransactionResponse
// @Failure      400   {object}  errs.AppError
// @Failure      403   {object}  errs.AppError
// @Failure      404   {object}  errs.AppError
// @Failure      422   {object}  errs.AppError
// @Security     BearerAuth
// @Router       /api/v1/transactions/transfer [post]
func (h *TransactionHandler) Transfer(c *gin.Context) {
	authUser, ok := utils.MustGetAuthUser(c)
	if !ok {
		return
	}

	var req dtoreq.TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dtores.HandleResponse(c, nil, err)
		return
	}

	tx, err := h.txSvc.Transfer(c.Request.Context(), authUser.ID, req.FromAccountNumber, req.ToAccountNumber, req.Amount)
	dtores.HandleResponse(c, dtores.FromEntityTransaction(tx), err)
}
