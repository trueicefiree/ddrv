package db

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq" // Import the PostgreSQL driver

	"github.com/forscht/ditto/pkg/migrate"
)

// Driver - fot now we only supports postgres
const Driver = "postgres"

// New creates a new database connection using the provided connection string.
// It returns the *sql.DB object representing the connection.
func New(connStr string, skipMigration bool) *sql.DB {
	// next a new database connection
	db, err := sql.Open(Driver, connStr)
	if err != nil {
		log.Fatal(err)
	}

	// Ping the database to ensure connectivity
	if err := db.Ping(); err != nil {
		log.Fatalf("db ping failed :%v", err)
	}

	// Perform database migrations
	if !skipMigration {
		if err := Migrate(db); err != nil {
			log.Fatalf("failed to execute migration :%v", err)
		}
	}

	return db
}

// Migrate performs database migrations using the provided *sql.DB connection.
func Migrate(db *sql.DB) error {
	m := migrate.NewMigrator(db)                  // Create a new migrator instance
	m.TransactionMode = migrate.SingleTransaction // Set the transaction mode to single transaction
	return m.Exec(migrate.Up, migrations...)      // Execute the migrations in the "up" direction
}
