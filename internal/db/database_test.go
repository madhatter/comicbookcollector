package db

import (
	"testing"
	"time"

	"github.com/madhatter/comicbookcollector/internal/locg"
)

func setupTestDB(t *testing.T) *Database {
	t.Helper()

	// Use an in-memory SQLite database for testing
	db, err := NewDatabase(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Run migrations to set up the schema
	if err := db.Migrate(); err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	t.Cleanup(func() { db.Close() })
	return db
}

func TestDatabase_Migrate(t *testing.T) {
	db := setupTestDB(t)

	// Check if the comics table exists by querying the sqlite_master table
	var count int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='comics'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query sqlite_master: %v", err)
	}

	if count != 1 {
		t.Fatalf("Expected comics table to be created, but it was not")
	}
}

func TestSaveComic_Insert(t *testing.T) {
	db := setupTestDB(t)

	comic := &locg.ComicBookDetails{
		ID:          1,
		Title:       "Batman",
		IssueNumber: 1,
		Publisher:   "DC Comics",
		ReleaseDate: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
	}

	if err := db.SaveComic(comic); err != nil {
		t.Fatalf("Failed to save comic: %v", err)
	}

	// Verify the comic was inserted
	var title string
	err := db.conn.QueryRow("SELECT title FROM comics WHERE locg_id = ?", comic.ID).Scan(&title)
	if err != nil {
		t.Fatalf("Failed to query saved comic: %v", err)
	}

	if title != comic.Title {
		t.Fatalf("Expected title '%s', got '%s'", comic.Title, title)
	}
}

func TestSaveComic_Update(t *testing.T) {
	db := setupTestDB(t)

	comic := &locg.ComicBookDetails{
		ID:          1,
		Title:       "Batman",
		IssueNumber: 1,
		Publisher:   "DC Comics",
		ReleaseDate: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
	}

	if err := db.SaveComic(comic); err != nil {
		t.Fatalf("Failed to save comic: %v", err)
	}

	// Update the comic's value and save again
	comic.Value = 599 // $5.99
	if err := db.SaveComic(comic); err != nil {
		t.Fatalf("Failed to update comic: %v", err)
	}

	// Verify the comic was updated
	var value int64
	err := db.conn.QueryRow("SELECT value FROM comics WHERE locg_id = ?", comic.ID).Scan(&value)
	if err != nil {
		t.Fatalf("Failed to query saved comic: %v", err)
	}

	if value != comic.Value {
		t.Fatalf("Expected value '%v', got '%v'", comic.Value, value)
	}

	// Make sure that there is only one record for this comic
	var count int
	err = db.conn.QueryRow("SELECT COUNT(*) FROM comics WHERE locg_id = ?", comic.ID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query comic count: %v", err)
	}

	if count != 1 {
		t.Fatalf("Expected 1 record for comic ID %d, got %d", comic.ID, count)
	}
}
