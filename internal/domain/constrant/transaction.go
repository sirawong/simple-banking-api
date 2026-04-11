package constrant

type TransactionType string
type TransactionStatus string

const (
	TransactionTypeDeposit  TransactionType = "deposit"
	TransactionTypeWithdraw TransactionType = "withdraw"
	TransactionTypeTransfer TransactionType = "transfer"

	TransactionStatusPending TransactionStatus = "pending"
	TransactionStatusSuccess TransactionStatus = "success"
	TransactionStatusFailed  TransactionStatus = "failed"
)
