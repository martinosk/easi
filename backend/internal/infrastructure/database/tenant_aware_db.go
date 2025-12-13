package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	sharedctx "easi/backend/internal/shared/context"
)

// escapeTenantID escapes single quotes in tenant ID for safe SQL interpolation
// This is defense in depth - TenantID validation already prevents special characters
func escapeTenantID(tenantID string) string {
	return strings.ReplaceAll(tenantID, "'", "''")
}

// buildSetTenantSQL builds a safe SET command for tenant context
func buildSetTenantSQL(tenantID string) string {
	return fmt.Sprintf("SET app.current_tenant = '%s'", escapeTenantID(tenantID))
}

// buildSetLocalTenantSQL builds a safe SET LOCAL command for transaction-scoped tenant context
func buildSetLocalTenantSQL(tenantID string) string {
	return fmt.Sprintf("SET LOCAL app.current_tenant = '%s'", escapeTenantID(tenantID))
}

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
	// Using safe builder with escaping for defense in depth
	_, err = conn.ExecContext(ctx, buildSetTenantSQL(tenantID.Value()))
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

// WithReadOnlyTx executes a function within a read-only transaction with tenant context
// This is the RECOMMENDED and CORRECT way to execute read queries with RLS
// It ensures proper connection lifecycle management and prevents connection leaks
func (t *TenantAwareDB) WithReadOnlyTx(ctx context.Context, fn func(*sql.Tx) error) error {
	// Begin read-only transaction with tenant context already set
	tx, err := t.BeginTxWithTenant(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return err
	}

	// Execute function
	err = fn(tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Commit to release connection back to pool
	return tx.Commit()
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

	// Using safe builder with escaping for defense in depth
	_, err = tx.ExecContext(ctx, buildSetLocalTenantSQL(tenantID.Value()))
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
