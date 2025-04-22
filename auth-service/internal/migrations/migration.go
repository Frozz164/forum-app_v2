// migrations.go в папке /internal

package migrations

import (
	"database/sql"
	"fmt"
)

// MigrateDB migrates the database
func MigrateDB(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE
		);
	`

	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	return nil
}
