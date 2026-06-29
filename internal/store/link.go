package store

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/broodco/linkshort/internal/model"
)

var ErrNotFound = errors.New("link not found")

func (s *Store) CreateLink(slug, targetURL, title string, expiresAt *time.Time) (*model.Link, error) {
	if slug == "" {
		slug = GenerateSlug()
	}
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

func (s *Store) CountClicks() (int64, error) {
	var count int64
	err := s.db.QueryRow(`SELECT COUNT(*) FROM clicks`).Scan(&count)
	return count, err
}

func (s *Store) CountLinks() (int64, error) {
	var count int64
	err := s.db.QueryRow(`SELECT COUNT(*) FROM links WHERE is_active = 1`).Scan(&count)
	return count, err
}

func (s *Store) ClicksPerLink(linkID int64) (int64, error) {
	var count int64
	err := s.db.QueryRow(`SELECT COUNT(*) FROM clicks WHERE link_id = ?`, linkID).Scan(&count)
	return count, err
}

func GenerateSlug() string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	const length = 6
	b := make([]byte, length)
	_, _ = rand.Read(b)
	for i := range b {
		b[i] = charset[b[i]%byte(len(charset))]
	}
	return string(b)
}

func (s *Store) ClicksPerDay(linkID int64) ([]model.ClickStat, error) {
	rows, err := s.db.Query(`
        SELECT DATE(clicked_at) as date, COUNT(*) as count
        FROM clicks
        WHERE link_id = ?
        AND clicked_at >= DATE('now', '-30 days')
        GROUP BY DATE(clicked_at)
        ORDER BY date ASC
    `, linkID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var stats []model.ClickStat
	for rows.Next() {
		var s model.ClickStat
		if err := rows.Scan(&s.Date, &s.Count); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, nil
}
