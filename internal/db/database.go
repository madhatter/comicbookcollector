package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/madhatter/comicbookcollector/internal/locg"
	_ "modernc.org/sqlite"
)

type Database struct {
	conn *sql.DB
}

func NewDatabase(dbFile string) (*Database, error) {
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

func (db *Database) SaveComic(comic *locg.ComicBookDetails) error {
	query := `
	INSERT INTO comics (
		upc, title, issue_number, cover_price, value, variant_info, series, publisher,
		description, release_date, locg_id, locg_url, locg_cover_image_url, storage_box
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(locg_id) DO UPDATE SET
		upc=excluded.upc,
		title=excluded.title,
		issue_number=excluded.issue_number,
		cover_price=excluded.cover_price,
		value=excluded.value,
		variant_info=excluded.variant_info,
		series=excluded.series,
		publisher=excluded.publisher,
		description=excluded.description,
		release_date=excluded.release_date,
		locg_url=excluded.locg_url,
		locg_cover_image_url=excluded.locg_cover_image_url,
		storage_box=excluded.storage_box;
	`

	_, err := db.conn.Exec(query,
		comic.UPC,
		comic.Title,
		comic.IssueNumber,
		comic.CoverPrice,
		comic.Value,
		comic.VariantInfo,
		comic.Series,
		comic.Publisher,
		comic.Description,
		comic.ReleaseDate.Format("2006-01-02"),
		comic.ID,
		comic.URL,
		comic.ImageLink,
		comic.StorageBox,
	)

	if err != nil {
		return fmt.Errorf("failed to save comic: %w", err)
	}

	return nil
}

func (db *Database) GetAllComics() ([]*locg.ComicBookDetails, error) {
	query := `SELECT upc, title, issue_number, cover_price, value, variant_info, series, publisher,
		description, release_date, locg_id, locg_url, locg_cover_image_url, storage_box FROM comics ORDER BY release_date DESC;`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query comics: %w", err)
	}
	defer rows.Close()

	var comics []*locg.ComicBookDetails
	for rows.Next() {
		var comic locg.ComicBookDetails
		var releaseDateStr string

		err := rows.Scan(
			&comic.UPC,
			&comic.Title,
			&comic.IssueNumber,
			&comic.CoverPrice,
			&comic.Value,
			&comic.VariantInfo,
			&comic.Series,
			&comic.Publisher,
			&comic.Description,
			&releaseDateStr,
			&comic.ID,
			&comic.URL,
			&comic.ImageLink,
			&comic.StorageBox,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan comic: %w", err)
		}

		comic.ReleaseDate, _ = time.Parse("2006-01-02", releaseDateStr)
		comics = append(comics, &comic)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating comics: %w", err)
	}

	return comics, nil
}

func (db *Database) GetComicByLoCGID(id int) (*locg.ComicBookDetails, error) {
	query := `SELECT upc, title, issue_number, cover_price, value, variant_info, series, publisher,
		description, release_date, locg_id, locg_url, locg_cover_image_url, storage_box FROM comics WHERE locg_id = ?;`

	row := db.conn.QueryRow(query, id)

	var comic locg.ComicBookDetails
	var releaseDateStr string

	err := row.Scan(
		&comic.UPC,
		&comic.Title,
		&comic.IssueNumber,
		&comic.CoverPrice,
		&comic.Value,
		&comic.VariantInfo,
		&comic.Series,
		&comic.Publisher,
		&comic.Description,
		&releaseDateStr,
		&comic.ID,
		&comic.URL,
		&comic.ImageLink,
		&comic.StorageBox,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to query comic by ID: %w", err)
	}

	comic.ReleaseDate, _ = time.Parse("2006-01-02", releaseDateStr)
	return &comic, nil
}

func (db *Database) GetComicByID(id int) (*locg.ComicBookDetails, error) {
	query := `SELECT upc, title, issue_number, cover_price, value, variant_info, series, publisher,
		description, release_date, locg_id, locg_url, locg_cover_image_url, storage_box FROM comics WHERE id = ?;`

	row := db.conn.QueryRow(query, id)

	var comic locg.ComicBookDetails
	var releaseDateStr string

	err := row.Scan(
		&comic.UPC,
		&comic.Title,
		&comic.IssueNumber,
		&comic.CoverPrice,
		&comic.Value,
		&comic.VariantInfo,
		&comic.Series,
		&comic.Publisher,
		&comic.Description,
		&releaseDateStr,
		&comic.ID,
		&comic.URL,
		&comic.ImageLink,
		&comic.StorageBox,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to query comic by ID: %w", err)
	}

	comic.ReleaseDate, _ = time.Parse("2006-01-02", releaseDateStr)
	return &comic, nil
}

func (db *Database) GetAllLoCGIDs() ([]int, error) {
	query := `SELECT locg_id FROM comics;`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query LoCG IDs: %w", err)
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan LoCG ID: %w", err)
		}
		ids = append(ids, id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating LoCG IDs: %w", err)
	}

	return ids, nil
}

func (db *Database) Close() error {
	return db.conn.Close()
}
