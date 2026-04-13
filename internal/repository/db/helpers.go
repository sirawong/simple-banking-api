package db

import (
	"context"
	"strings"

	"gorm.io/gorm"

	pkgerrs "github.com/sirawong/simple-banking-api/pkg/errs"
)

func isDuplicateError(err error) bool {
	return strings.Contains(err.Error(), "duplicate key") ||
		strings.Contains(err.Error(), "UNIQUE constraint failed")
}

func notDeleted(db *gorm.DB) *gorm.DB {
	return db.Where("deleted_at IS NULL")
}

func requireTx(ctx context.Context) error {
	if _, ok := ctx.Value(txContextKey{}).(*gorm.DB); !ok {
		return pkgerrs.ErrInternal.New("must be called within a transaction")
	}
	return nil
}
