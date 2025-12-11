// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package storage // import "miniflux.app/v2/internal/storage"

import (
	"database/sql"
	"fmt"
	"time"

	"miniflux.app/v2/internal/model"
)

// GlobalUnreadStat contains aggregated unread information for a user reading list.
type GlobalUnreadStat struct {
	Count  int
	Newest time.Time
}

// FeedUnreadStat contains aggregated unread information for a feed.
type FeedUnreadStat struct {
	FeedID int64
	Count  int
	Newest time.Time
}

// CategoryUnreadStat contains aggregated unread information for a category (label).
type CategoryUnreadStat struct {
	CategoryID int64
	Title      string
	Count      int
	Newest     time.Time
}

// VisibleUnreadGlobalStat returns the number of unread entries and the newest timestamp for the reading list.
func (s *Storage) VisibleUnreadGlobalStat(userID int64) (GlobalUnreadStat, error) {
	var result GlobalUnreadStat
	var newest sql.NullTime

	query := `
        SELECT
            COALESCE(count(*), 0) as count,
            MAX(e.published_at) as newest
        FROM entries e
            JOIN feeds f ON f.id = e.feed_id
            JOIN categories c ON c.id = f.category_id
        WHERE
            e.user_id=$1 AND e.status=$2 AND
            f.hide_globally IS FALSE AND
            c.hide_globally IS FALSE
    `

	if err := s.db.QueryRow(query, userID, model.EntryStatusUnread).Scan(&result.Count, &newest); err != nil {
		return result, fmt.Errorf("store: unable to fetch unread statistics: %w", err)
	}

	if newest.Valid {
		result.Newest = newest.Time
	}

	return result, nil
}

// VisibleUnreadFeedStats returns unread counts grouped by feed for visible subscriptions.
func (s *Storage) VisibleUnreadFeedStats(userID int64) ([]FeedUnreadStat, error) {
	query := `
        SELECT
            e.feed_id,
            COUNT(*) as count,
            MAX(e.published_at) as newest
        FROM entries e
            JOIN feeds f ON f.id = e.feed_id
            JOIN categories c ON c.id = f.category_id
        WHERE
            e.user_id=$1 AND e.status=$2 AND
            f.hide_globally IS FALSE AND
            c.hide_globally IS FALSE
        GROUP BY e.feed_id
    `

	rows, err := s.db.Query(query, userID, model.EntryStatusUnread)
	if err != nil {
		return nil, fmt.Errorf("store: unable to fetch unread feed statistics: %w", err)
	}
	defer rows.Close()

	stats := make([]FeedUnreadStat, 0)
	for rows.Next() {
		var stat FeedUnreadStat
		var newest sql.NullTime

		if err := rows.Scan(&stat.FeedID, &stat.Count, &newest); err != nil {
			return nil, fmt.Errorf("store: unable to scan unread feed statistics: %w", err)
		}

		if newest.Valid {
			stat.Newest = newest.Time
		}

		stats = append(stats, stat)
	}

	return stats, nil
}

// VisibleUnreadCategoryStats returns unread counts grouped by category (label) for visible subscriptions.
func (s *Storage) VisibleUnreadCategoryStats(userID int64) ([]CategoryUnreadStat, error) {
	query := `
        SELECT
            c.id,
            c.title,
            COUNT(*) as count,
            MAX(e.published_at) as newest
        FROM entries e
            JOIN feeds f ON f.id = e.feed_id
            JOIN categories c ON c.id = f.category_id
        WHERE
            e.user_id=$1 AND e.status=$2 AND
            f.hide_globally IS FALSE AND
            c.hide_globally IS FALSE
        GROUP BY c.id, c.title
    `

	rows, err := s.db.Query(query, userID, model.EntryStatusUnread)
	if err != nil {
		return nil, fmt.Errorf("store: unable to fetch unread category statistics: %w", err)
	}
	defer rows.Close()

	stats := make([]CategoryUnreadStat, 0)
	for rows.Next() {
		var stat CategoryUnreadStat
		var newest sql.NullTime

		if err := rows.Scan(&stat.CategoryID, &stat.Title, &stat.Count, &newest); err != nil {
			return nil, fmt.Errorf("store: unable to scan unread category statistics: %w", err)
		}

		if newest.Valid {
			stat.Newest = newest.Time
		}

		stats = append(stats, stat)
	}

	return stats, nil
}
