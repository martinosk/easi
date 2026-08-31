package database

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTxFromContext_MissingReturnsFalse(t *testing.T) {
	_, ok := TxFromContext(context.Background())

	assert.False(t, ok)
}

func TestWithTx_TxFromContext_RoundTrip(t *testing.T) {
	var tx *sql.Tx
	ctx := WithTx(context.Background(), tx)

	got, ok := TxFromContext(ctx)

	assert.True(t, ok)
	assert.Same(t, tx, got)
}
