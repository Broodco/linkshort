package store

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func New(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path+"?_parse_time=true")
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	s := &Store{db: db}

	if err := s.migrate(); err != nil {
		return nil, err
	}

	log.Printf("database ready at %s", path)
	return s, nil
}

func (s *Store) migrate() error {
	_, err := s.db.Exec(`
        CREATE TABLE IF NOT EXISTS links (
            id         INTEGER PRIMARY KEY AUTOINCREMENT,
            slug       TEXT UNIQUE NOT NULL,
            target_url TEXT NOT NULL,
            title      TEXT,
            expires_at DATETIME,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            is_active  INTEGER DEFAULT 1
        );

        CREATE TABLE IF NOT EXISTS clicks (
            id         INTEGER PRIMARY KEY AUTOINCREMENT,
            link_id    INTEGER REFERENCES links(id),
            clicked_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            referrer   TEXT,
            user_agent TEXT
        );
    `)
	return err
}
