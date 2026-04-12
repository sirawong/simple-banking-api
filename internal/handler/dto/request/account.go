package request

// CreateAccountRequest godoc
type CreateAccountRequest struct {
	Currency string `json:"currency" binding:"required"`
}

// ListTransactionsRequest godoc
type ListTransactionsRequest struct {
	Page  int `form:"page,default=1" binding:"min=1"`
	Limit int `form:"limit,default=20" binding:"min=1,max=100"`
}
