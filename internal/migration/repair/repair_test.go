package repair

import (
	"context"
	"testing"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/stretchr/testify/assert"
	"github.com/supabase/cli/internal/testing/pgtest"
)

var dbConfig = pgconn.Config{
	Host:     "localhost",
	Port:     5432,
	User:     "admin",
	Password: "password",
	Database: "postgres",
}

func TestRepairCommand(t *testing.T) {
	t.Run("applies new version", func(t *testing.T) {
		// Setup mock postgres
		conn := pgtest.NewConn()
		defer conn.Close(t)
		conn.Query(CREATE_VERSION_SCHEMA).
			Reply("CREATE SCHEMA").
			Query(CREATE_VERSION_TABLE).
			Reply("CREATE TABLE").
			Query(INSERT_MIGRATION_VERSION, "0").
			Reply("INSERT 0 1")
		// Run test
		err := Run(context.Background(), dbConfig, "0", Applied, conn.Intercept)
		// Check error
		assert.NoError(t, err)
	})

	t.Run("reverts old version", func(t *testing.T) {
		// Setup mock postgres
		conn := pgtest.NewConn()
		defer conn.Close(t)
		conn.Query(CREATE_VERSION_SCHEMA).
			Reply("CREATE SCHEMA").
			Query(CREATE_VERSION_TABLE).
			Reply("CREATE TABLE").
			Query(DELETE_MIGRATION_VERSION, "0").
			Reply("DELETE 1")
		// Run test
		err := Run(context.Background(), dbConfig, "0", Reverted, conn.Intercept)
		// Check error
		assert.NoError(t, err)
	})

	t.Run("throws error on connect failure", func(t *testing.T) {
		// Run test
		err := Run(context.Background(), pgconn.Config{}, "0", Applied)
		// Check error
		assert.ErrorContains(t, err, "invalid port (outside range)")
	})

	t.Run("throws error on insert failure", func(t *testing.T) {
		// Setup mock postgres
		conn := pgtest.NewConn()
		defer conn.Close(t)
		conn.Query(CREATE_VERSION_SCHEMA).
			Reply("CREATE SCHEMA").
			Query(CREATE_VERSION_TABLE).
			Reply("CREATE TABLE").
			Query(INSERT_MIGRATION_VERSION, "0").
			ReplyError(pgerrcode.DuplicateObject, `relation "supabase_migrations.schema_migrations" does not exist`)
		// Run test
		err := Run(context.Background(), dbConfig, "0", Applied, conn.Intercept)
		// Check error
		assert.ErrorContains(t, err, `ERROR: relation "supabase_migrations.schema_migrations" does not exist (SQLSTATE 42710)`)
	})
}
