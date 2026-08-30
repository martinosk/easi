package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"sync"
	"testing"

	sharedctx "easi/backend/internal/shared/context"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeTxDriver struct{}

func (fakeTxDriver) Open(string) (driver.Conn, error) { return &fakeTxConn{}, nil }

type fakeTxConn struct{}

func (c *fakeTxConn) Prepare(query string) (driver.Stmt, error) { return fakeTxStmt{}, nil }
func (c *fakeTxConn) Close() error                              { return nil }
func (c *fakeTxConn) Begin() (driver.Tx, error)                 { return fakeTx{}, nil }
func (c *fakeTxConn) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	return fakeTx{}, nil
}

type fakeTx struct{}

func (fakeTx) Commit() error   { return nil }
func (fakeTx) Rollback() error { return nil }

type fakeTxStmt struct{}

func (fakeTxStmt) Close() error  { return nil }
func (fakeTxStmt) NumInput() int { return -1 }
func (fakeTxStmt) Exec(args []driver.Value) (driver.Result, error) {
	return driver.RowsAffected(1), nil
}
func (fakeTxStmt) Query([]driver.Value) (driver.Rows, error) {
	return nil, errors.New("query not supported by fake driver")
}

var registerFakeTxDriverOnce sync.Once

func newFakeTxDB(t *testing.T) *sql.DB {
	t.Helper()
	registerFakeTxDriverOnce.Do(func() {
		sql.Register("fake-tx-driver", fakeTxDriver{})
	})
	db, err := sql.Open("fake-tx-driver", "fake")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func testTenantCtx(t *testing.T) context.Context {
	t.Helper()
	tid, err := sharedvo.NewTenantID("tx-routing-test")
	require.NoError(t, err)
	return sharedctx.WithTenant(context.Background(), tid)
}

func newJoinedTenantDB(t *testing.T) (*TenantAwareDB, context.Context, *sql.Tx) {
	t.Helper()
	db := newFakeTxDB(t)
	tx, err := db.Begin()
	require.NoError(t, err)
	return NewTenantAwareDB(db), WithTx(context.Background(), tx), tx
}

func newOwningTenantDB(t *testing.T) (*TenantAwareDB, context.Context) {
	t.Helper()
	return NewTenantAwareDB(newFakeTxDB(t)), testTenantCtx(t)
}

func TestExecContext_JoinsTxFromContext_WithoutRequiringTenant(t *testing.T) {
	tenantDB, ctx, tx := newJoinedTenantDB(t)
	defer func() { _ = tx.Rollback() }()

	_, err := tenantDB.ExecContext(ctx, "INSERT INTO test.whatever VALUES (1)")

	assert.NoError(t, err)
}

func TestExecContext_WithoutTxInContext_StillRequiresTenant(t *testing.T) {
	tenantDB := NewTenantAwareDB(newFakeTxDB(t))

	_, err := tenantDB.ExecContext(context.Background(), "INSERT INTO test.whatever VALUES (1)")

	require.Error(t, err)
}

func TestWithReadOnlyTx_JoinsExistingTx_LeavesCommitToOwner(t *testing.T) {
	tenantDB, ctx, tx := newJoinedTenantDB(t)

	var sawTx *sql.Tx
	err := tenantDB.WithReadOnlyTx(ctx, func(inner *sql.Tx) error {
		sawTx = inner
		return nil
	})
	require.NoError(t, err)
	assert.Same(t, tx, sawTx)

	assert.NoError(t, tx.Commit())
}

func TestWithReadOnlyTx_JoinsExistingTx_PropagatesErrorWithoutTouchingOwnerTx(t *testing.T) {
	tenantDB, ctx, tx := newJoinedTenantDB(t)

	fnErr := errors.New("fn failed")
	err := tenantDB.WithReadOnlyTx(ctx, func(*sql.Tx) error {
		return fnErr
	})
	require.ErrorIs(t, err, fnErr)

	assert.NoError(t, tx.Rollback())
}

func TestWithReadOnlyTx_NoTxInContext_BeginsOwnAndCommits(t *testing.T) {
	tenantDB, ctx := newOwningTenantDB(t)

	called := false
	err := tenantDB.WithReadOnlyTx(ctx, func(*sql.Tx) error {
		called = true
		return nil
	})

	require.NoError(t, err)
	assert.True(t, called)
}

func TestRunInTx_JoinsExistingTx_DoesNotBeginOrCommitANewOne(t *testing.T) {
	tenantDB, ctx, tx := newJoinedTenantDB(t)

	var sawTx *sql.Tx
	err := tenantDB.RunInTx(ctx, func(fnCtx context.Context) error {
		got, ok := TxFromContext(fnCtx)
		require.True(t, ok)
		sawTx = got
		return nil
	})

	require.NoError(t, err)
	assert.Same(t, tx, sawTx)
	assert.NoError(t, tx.Commit())
}

func TestRunInTx_NoTxInContext(t *testing.T) {
	fnErr := errors.New("boom")

	tests := []struct {
		name       string
		fnResult   error
		verifyDone func(t *testing.T, tx *sql.Tx)
	}{
		{
			name:     "commits on success",
			fnResult: nil,
			verifyDone: func(t *testing.T, tx *sql.Tx) {
				assert.ErrorIs(t, tx.Commit(), sql.ErrTxDone)
			},
		},
		{
			name:     "rolls back on error",
			fnResult: fnErr,
			verifyDone: func(t *testing.T, tx *sql.Tx) {
				assert.ErrorIs(t, tx.Rollback(), sql.ErrTxDone)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tenantDB, ctx := newOwningTenantDB(t)

			var joinedTx *sql.Tx
			err := tenantDB.RunInTx(ctx, func(fnCtx context.Context) error {
				joinedTx, _ = TxFromContext(fnCtx)
				return tt.fnResult
			})

			require.ErrorIs(t, err, tt.fnResult)
			require.NotNil(t, joinedTx)
			tt.verifyDone(t, joinedTx)
		})
	}
}

func TestRunInTx_NoTxInContext_RollsBackOnPanic(t *testing.T) {
	tenantDB, ctx := newOwningTenantDB(t)

	var joinedTx *sql.Tx
	func() {
		defer func() { _ = recover() }()
		_ = tenantDB.RunInTx(ctx, func(fnCtx context.Context) error {
			tx, _ := TxFromContext(fnCtx)
			joinedTx = tx
			panic("simulated panic")
		})
	}()

	require.NotNil(t, joinedTx)
	assert.ErrorIs(t, joinedTx.Rollback(), sql.ErrTxDone)
}
