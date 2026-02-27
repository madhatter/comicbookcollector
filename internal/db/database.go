package db

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

const dbFile = "cbc.db"

type Database struct {
	conn *sql.DB
}

func NewDatabase() (*Database, error) {
	conn, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	db := &Database{conn: conn}
	return db, nil
}

func (db *Database) Migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS comics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		upc TEXT,
		title TEXT NOT NULL,
		issue_number INTEGER NOT NULL,
		cover_price INTEGER,
		value INTEGER,
		variant_info TEXT,
		series TEXT,
		publisher TEXT NOT NULL,
		description TEXT,
		release_date TEXT,
		locg_id INTEGER UNIQUE NOT NULL,
		locg_url TEXT NOT NULL,
		locg_cover_image_url TEXT,
		storage_box TEXT
	);
	`

	_, err := db.conn.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	return nil
}

func (db *Database) Close() error {
	return db.conn.Close()
}
