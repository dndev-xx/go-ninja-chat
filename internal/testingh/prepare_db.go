//go:build integration

package testingh

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dndev-xx/go-ninja-chat/internal/store"
)

var migrationLock sync.Mutex

func PrepareDB(ctx context.Context, t *testing.T, dbName string) (st *store.Client, cleanUp func(ctx context.Context)) {
	t.Helper()
	require.NotEmpty(t, dbName)

	pgx, err := store.NewPgxDB(store.NewPgxOptions(
		Config.PostgresAddress,
		Config.PostgresUser,
		Config.PostgresPassword,
		"postgres",
	))
	require.NoError(t, err)
	require.NoError(t, createDatabase(ctx, dbName, pgx))

	client, err := store.NewPSQLClient(store.NewPSQLOptions(
		Config.PostgresAddress,
		Config.PostgresUser,
		Config.PostgresPassword,
		dbName,
		Config.PostgresDebug,
	))
	require.NoError(t, err)

	migrationLock.Lock()
	{
		// NOTE: Schema migration is not thread-safe :(
		err = client.Schema.Create(ctx)
	}
	migrationLock.Unlock()
	require.NoError(t, err)

	return client, func(ctx2 context.Context) {
		assert.NoError(t, client.Close())
		assert.NoError(t, dropDatabaseIfExists(ctx2, dbName, pgx))
		assert.NoError(t, pgx.Close())
	}
}

func createDatabase(ctx context.Context, dbName string, db *sql.DB) error {
	if err := dropDatabaseIfExists(ctx, dbName, db); err != nil {
		return fmt.Errorf("drop database %s: %v", dbName, err)
	}

	if _, err := db.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE %q", dbName)); err != nil {
		return fmt.Errorf("create database %s: %v", dbName, err)
	}
	return nil
}

func dropDatabaseIfExists(ctx context.Context, dbName string, db *sql.DB) error {
	_, err := db.ExecContext(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %q", dbName))
	return err
}
