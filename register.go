package goose

import (
	"context"
	"database/sql"
	"fmt"
	"runtime"
)

// GoMigrationContext is a Go migration func that is run within a transaction and receives a
// context.
type GoMigrationContext func(ctx context.Context, tx *sql.Tx) error

// AddMigrationContext adds Go migrations.
func AddMigrationContext(up, down GoMigrationContext, opts ...MigrationOption) {
	_, filename, _, _ := runtime.Caller(1)
	AddNamedMigrationContext(filename, up, down, opts...)
}

// AddNamedMigrationContext adds named Go migrations.
func AddNamedMigrationContext(filename string, up, down GoMigrationContext, opts ...MigrationOption) {
	var mc MigrationConfig
	for _, opt := range opts {
		opt(&mc)
	}

	if err := register(
		mc.Scope,
		filename,
		true,
		&GoFunc{RunTx: up, Mode: TransactionEnabled},
		&GoFunc{RunTx: down, Mode: TransactionEnabled},
	); err != nil {
		panic(err)
	}
}

// GoMigrationNoTxContext is a Go migration func that is run outside a transaction and receives a
// context.
type GoMigrationNoTxContext func(ctx context.Context, db *sql.DB) error

// AddMigrationNoTxContext adds Go migrations that will be run outside transaction.
func AddMigrationNoTxContext(up, down GoMigrationNoTxContext, opts ...MigrationOption) {
	_, filename, _, _ := runtime.Caller(1)
	AddNamedMigrationNoTxContext(filename, up, down, opts...)
}

// AddNamedMigrationNoTxContext adds named Go migrations that will be run outside transaction.
func AddNamedMigrationNoTxContext(filename string, up, down GoMigrationNoTxContext, opts ...MigrationOption) {
	var mc MigrationConfig
	for _, opt := range opts {
		opt(&mc)
	}

	if err := register(
		mc.Scope,
		filename,
		false,
		&GoFunc{RunDB: up, Mode: TransactionDisabled},
		&GoFunc{RunDB: down, Mode: TransactionDisabled},
	); err != nil {
		panic(err)
	}
}

func register(scope, filename string, useTx bool, up, down *GoFunc) error {
	v, _ := NumericComponent(filename)
	if versionMap, ok := registeredGoMigrations[scope]; ok {
		if existing, ok := versionMap[v]; ok {
			return fmt.Errorf("failed to add migration %q: version %d conflicts with %q",
				filename,
				v,
				existing.Source,
			)
		}
	}
	// Add to global as a registered migration.
	m := NewGoMigration(v, up, down)
	m.Source = filename
	// We explicitly set transaction to maintain existing behavior. Both up and down may be nil, but
	// we know based on the register function what the user is requesting.
	m.UseTx = useTx
	if _, ok := registeredGoMigrations[scope]; !ok {
		registeredGoMigrations[scope] = make(map[int64]*Migration)
	}
	registeredGoMigrations[scope][v] = m
	return nil
}

// withContext changes the signature of a function that receives one argument to receive a context
// and the argument.
func withContext[T any](fn func(T) error) func(context.Context, T) error {
	if fn == nil {
		return nil
	}
	return func(ctx context.Context, t T) error {
		return fn(t)
	}
}

// withoutContext changes the signature of a function that receives a context and one argument to
// receive only the argument. When called the passed context is always context.Background().
func withoutContext[T any](fn func(context.Context, T) error) func(T) error {
	if fn == nil {
		return nil
	}
	return func(t T) error {
		return fn(context.Background(), t)
	}
}

// GoMigration is a Go migration func that is run within a transaction.
//
// Deprecated: Use GoMigrationContext.
type GoMigration func(tx *sql.Tx) error

// GoMigrationNoTx is a Go migration func that is run outside a transaction.
//
// Deprecated: Use GoMigrationNoTxContext.
type GoMigrationNoTx func(db *sql.DB) error

// AddMigration adds Go migrations.
//
// Deprecated: Use AddMigrationContext.
func AddMigration(up, down GoMigration, opts ...MigrationOption) {
	_, filename, _, _ := runtime.Caller(1)
	AddNamedMigrationContext(filename, withContext(up), withContext(down), opts...)
}

// AddNamedMigration adds named Go migrations.
//
// Deprecated: Use AddNamedMigrationContext.
func AddNamedMigration(filename string, up, down GoMigration, opts ...MigrationOption) {
	AddNamedMigrationContext(filename, withContext(up), withContext(down), opts...)
}

// AddMigrationNoTx adds Go migrations that will be run outside transaction.
//
// Deprecated: Use AddMigrationNoTxContext.
func AddMigrationNoTx(up, down GoMigrationNoTx, opts ...MigrationOption) {
	_, filename, _, _ := runtime.Caller(1)
	AddNamedMigrationNoTxContext(filename, withContext(up), withContext(down), opts...)
}

// AddNamedMigrationNoTx adds named Go migrations that will be run outside transaction.
//
// Deprecated: Use AddNamedMigrationNoTxContext.
func AddNamedMigrationNoTx(filename string, up, down GoMigrationNoTx, opts ...MigrationOption) {
	AddNamedMigrationNoTxContext(filename, withContext(up), withContext(down), opts...)
}
