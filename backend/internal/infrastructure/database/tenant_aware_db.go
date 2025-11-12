package database

import (
	"context"
	"database/sql"
	"fmt"

	sharedctx "easi/backend/internal/shared/context"
)

// TenantAwareDB wraps a database connection and automatically sets tenant context
// for Row-Level Security (RLS) policies
type TenantAwareDB struct {
	db *sql.DB
}

// NewTenantAwareDB creates a new tenant-aware database wrapper
func NewTenantAwareDB(db *sql.DB) *TenantAwareDB {
	return &TenantAwareDB{
		db: db,
	}
}

// DB returns the underlying database connection
func (t *TenantAwareDB) DB() *sql.DB {
	return t.db
}

// setTenantContext sets the PostgreSQL session variable for RLS
func (t *TenantAwareDB) setTenantContext(ctx context.Context, conn *sql.Conn) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return fmt.Errorf("failed to get tenant from context: %w", err)
	}

	// Set PostgreSQL session variable for RLS policies
	_, err = conn.ExecContext(ctx, "SET app.current_tenant = $1", tenantID.Value())
	if err != nil {
		return fmt.Errorf("failed to set tenant context: %w", err)
	}

	return nil
}

// WithTenantContext executes a function with a tenant-aware database connection
// The connection has the PostgreSQL session variable set for RLS
func (t *TenantAwareDB) WithTenantContext(ctx context.Context, fn func(*sql.Conn) error) error {
	// Acquire connection from pool
	conn, err := t.db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer conn.Close()

	// Set tenant context for RLS
	if err := t.setTenantContext(ctx, conn); err != nil {
		return err
	}

	// Execute function with tenant-aware connection
	return fn(conn)
}

// QueryContext executes a query with automatic tenant context
func (t *TenantAwareDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	var rows *sql.Rows
	var queryErr error

	err := t.WithTenantContext(ctx, func(conn *sql.Conn) error {
		rows, queryErr = conn.QueryContext(ctx, query, args...)
		return queryErr
	})

	if err != nil {
		return nil, err
	}
	return rows, nil
}

// QueryRowContext executes a query that returns at most one row
func (t *TenantAwareDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	var row *sql.Row

	// Note: We can't return error here due to sql.Row's interface
	// The error will be captured when Scan is called
	_ = t.WithTenantContext(ctx, func(conn *sql.Conn) error {
		row = conn.QueryRowContext(ctx, query, args...)
		return nil
	})

	return row
}

// ExecContext executes a query without returning any rows
func (t *TenantAwareDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	var result sql.Result
	var execErr error

	err := t.WithTenantContext(ctx, func(conn *sql.Conn) error {
		result, execErr = conn.ExecContext(ctx, query, args...)
		return execErr
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// BeginTxWithTenant begins a transaction with tenant context set
func (t *TenantAwareDB) BeginTxWithTenant(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	// For transactions, we need to set tenant context immediately after beginning
	tx, err := t.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Set tenant context within transaction
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to get tenant from context: %w", err)
	}

	_, err = tx.ExecContext(ctx, "SET LOCAL app.current_tenant = $1", tenantID.Value())
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to set tenant context in transaction: %w", err)
	}

	return tx, nil
}

// Ping verifies the connection to the database
func (t *TenantAwareDB) Ping(ctx context.Context) error {
	return t.db.PingContext(ctx)
}

// Close closes the database connection
func (t *TenantAwareDB) Close() error {
	return t.db.Close()
}

// PrepareContext creates a prepared statement for later queries or executions
// Note: Prepared statements don't maintain session variables, so this should be used
// within a WithTenantContext call or transaction
func (t *TenantAwareDB) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return t.db.PrepareContext(ctx, query)
}
