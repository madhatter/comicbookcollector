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

func TestGetAllComics(t *testing.T) {
	db := setupTestDB(t)

	comic1 := &locg.ComicBookDetails{
		ID:          1,
		Title:       "Batman",
		IssueNumber: 1,
		Publisher:   "DC Comics",
		ReleaseDate: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
	}

	comic2 := &locg.ComicBookDetails{
		ID:          2,
		Title:       "Spider-Man",
		IssueNumber: 1,
		Publisher:   "Marvel Comics",
		ReleaseDate: time.Date(2024, 2, 20, 0, 0, 0, 0, time.UTC),
	}

	if err := db.SaveComic(comic1); err != nil {
		t.Fatalf("Failed to save comic1: %v", err)
	}

	if err := db.SaveComic(comic2); err != nil {
		t.Fatalf("Failed to save comic2: %v", err)
	}

	comics, err := db.GetAllComics()
	if err != nil {
		t.Fatalf("Failed to get all comics: %v", err)
	}

	if len(comics) != 2 {
		t.Fatalf("Expected 2 comics, got %d", len(comics))
	}

	// Verify the comics are returned in the correct order (newest first)
	if comics[0].ID != comic2.ID || comics[1].ID != comic1.ID {
		t.Fatalf("Expected comics to be ordered by release date descending")
	}
}

func TestGetComicByLoCGID(t *testing.T) {
	db := setupTestDB(t)

	comic1 := &locg.ComicBookDetails{
		ID:          1,
		Title:       "Batman",
		IssueNumber: 1,
		Publisher:   "DC Comics",
		ReleaseDate: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
	}

	comic2 := &locg.ComicBookDetails{
		ID:          2,
		Title:       "Spider-Man",
		IssueNumber: 1,
		Publisher:   "Marvel Comics",
		ReleaseDate: time.Date(2024, 2, 20, 0, 0, 0, 0, time.UTC),
	}
	if err := db.SaveComic(comic1); err != nil {
		t.Fatalf("Failed to save comic: %v", err)
	}

	if err := db.SaveComic(comic2); err != nil {
		t.Fatalf("Failed to save comic: %v", err)
	}

	retrievedComic, err := db.GetComicByLoCGID(comic2.ID)
	if err != nil {
		t.Fatalf("Failed to get comic by ID: %v", err)
	}

	if retrievedComic == nil {
		t.Fatalf("Expected to retrieve a comic, got nil")
	}

	if retrievedComic.ID != comic2.ID || retrievedComic.Title != comic2.Title {
		t.Fatalf("Retrieved comic does not match saved comic")
	}
}

func TestGetComicByLoCGID_NotFound(t *testing.T) {
	db := setupTestDB(t)

	comic, err := db.GetComicByLoCGID(999)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if comic != nil {
		t.Fatalf("Expected nil, got a comic")
	}
}

func TestGetComicByID(t *testing.T) {
	db := setupTestDB(t)

	// Since the ID is auto-incremented, we need to make sure there is only one comic and that it has ID 1
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

	retrievedComic, err := db.GetComicByID(1)
	if err != nil {
		t.Fatalf("Failed to get comic by ID: %v", err)
	}

	if retrievedComic == nil {
		t.Fatalf("Expected to retrieve a comic, got nil")
	}

	if retrievedComic.ID != comic.ID || retrievedComic.Title != comic.Title {
		t.Fatalf("Retrieved comic does not match saved comic")
	}
}

func TestGetComicByID_NotFound(t *testing.T) {
	db := setupTestDB(t)

	comic, err := db.GetComicByID(999)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if comic != nil {
		t.Fatalf("Expected nil, got a comic")
	}
}

func TestGetAllLoCGIDs(t *testing.T) {
	db := setupTestDB(t)

	comic1 := &locg.ComicBookDetails{
		ID:          1,
		Title:       "Batman",
		IssueNumber: 1,
		Publisher:   "DC Comics",
		ReleaseDate: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
	}

	comic2 := &locg.ComicBookDetails{
		ID:          2,
		Title:       "Spider-Man",
		IssueNumber: 1,
		Publisher:   "Marvel Comics",
		ReleaseDate: time.Date(2024, 2, 20, 0, 0, 0, 0, time.UTC),
	}

	if err := db.SaveComic(comic1); err != nil {
		t.Fatalf("Failed to save comic1: %v", err)
	}

	if err := db.SaveComic(comic2); err != nil {
		t.Fatalf("Failed to save comic2: %v", err)
	}

	ids, err := db.GetAllLoCGIDs()
	if err != nil {
		t.Fatalf("Failed to get all LoCG IDs: %v", err)
	}

	expectedIDs := []int{comic1.ID, comic2.ID}
	if len(ids) != len(expectedIDs) {
		t.Fatalf("Expected %d IDs, got %d", len(expectedIDs), len(ids))
	}

	for i, id := range ids {
		if id != expectedIDs[i] {
			t.Fatalf("Expected ID %d at index %d, got %d", expectedIDs[i], i, id)
		}
	}
}
