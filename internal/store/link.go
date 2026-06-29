package store

import (
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/broodco/linkshort/internal/model"
)

var ErrNotFound = errors.New("link not found")

func (s *Store) CreateLink(slug, targetURL, title string, expiresAt *time.Time) (*model.Link, error) {
	res, err := s.db.Exec(`
        INSERT INTO links (slug, target_url, title, expires_at)
        VALUES (?, ?, ?, ?)
    `, slug, targetURL, title, expiresAt)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &model.Link{
		ID:        id,
		Slug:      slug,
		TargetURL: targetURL,
		Title:     title,
		ExpiresAt: expiresAt,
		IsActive:  true,
	}, nil
}

func (s *Store) GetLinkBySlug(slug string) (*model.Link, error) {
	var link model.Link
	var expiresAt sql.NullTime

	err := s.db.QueryRow(`
        SELECT id, slug, target_url, title, expires_at, created_at, is_active
        FROM links
        WHERE slug = ? AND is_active = 1
    `, slug).Scan(
		&link.ID,
		&link.Slug,
		&link.TargetURL,
		&link.Title,
		&expiresAt,
		&link.CreatedAt,
		&link.IsActive,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	if expiresAt.Valid {
		link.ExpiresAt = &expiresAt.Time
	}

	return &link, nil
}

func (s *Store) ListLinks() ([]*model.Link, error) {
	rows, err := s.db.Query(`
        SELECT id, slug, target_url, title, expires_at, created_at, is_active
        FROM links
        ORDER BY created_at DESC
    `)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var links []*model.Link
	for rows.Next() {
		var link model.Link
		var expiresAt sql.NullTime

		err := rows.Scan(
			&link.ID,
			&link.Slug,
			&link.TargetURL,
			&link.Title,
			&expiresAt,
			&link.CreatedAt,
			&link.IsActive,
		)
		if err != nil {
			return nil, err
		}

		if expiresAt.Valid {
			link.ExpiresAt = &expiresAt.Time
		}

		links = append(links, &link)
	}

	return links, nil
}

func (s *Store) DeleteLink(slug string) error {
	_, err := s.db.Exec(`
        UPDATE links SET is_active = 0 WHERE slug = ?
    `, slug)
	return err
}

func (s *Store) RecordClick(linkID int64, referrer, userAgent string) {
	_, err := s.db.Exec(`
        INSERT INTO clicks (link_id, referrer, user_agent)
        VALUES (?, ?, ?)
    `, linkID, referrer, userAgent)
	if err != nil {
		log.Printf("error recording click: %v", err)
	}
}
