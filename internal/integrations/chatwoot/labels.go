package chatwoot

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"

	"wzap/internal/logger"
)

var (
	cwDBPools   = make(map[string]*pgxpool.Pool)
	cwDBPoolsMu sync.Mutex
)

func getCWPool(ctx context.Context, dbURI string) (*pgxpool.Pool, error) {
	cwDBPoolsMu.Lock()
	defer cwDBPoolsMu.Unlock()
	if pool, ok := cwDBPools[dbURI]; ok {
		return pool, nil
	}
	pool, err := pgxpool.New(ctx, dbURI)
	if err != nil {
		return nil, fmt.Errorf("failed to create chatwoot db pool: %w", err)
	}
	cwDBPools[dbURI] = pool
	return pool, nil
}

func addLabelToContact(ctx context.Context, dbURI, inboxName string, contactID int) error {
	if dbURI == "" || inboxName == "" {
		return nil
	}

	label := strings.ToLower(strings.TrimSpace(inboxName))
	if label == "" {
		return nil
	}

	pool, err := getCWPool(ctx, dbURI)
	if err != nil {
		return err
	}

	var tagID int
	err = pool.QueryRow(ctx,
		`INSERT INTO tags (name) VALUES ($1) ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name RETURNING id`,
		label).Scan(&tagID)
	if err != nil {
		return fmt.Errorf("failed to upsert tag: %w", err)
	}

	var exists bool
	err = pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM taggings WHERE tag_id = $1 AND taggable_type = 'Contact' AND taggable_id = $2 AND context = 'labels')`,
		tagID, contactID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check tagging existence: %w", err)
	}

	if exists {
		return nil
	}

	_, err = pool.Exec(ctx,
		`INSERT INTO taggings (tag_id, taggable_type, taggable_id, context, created_at) VALUES ($1, 'Contact', $2, 'labels', NOW())`,
		tagID, contactID)
	if err != nil {
		return fmt.Errorf("failed to insert tagging: %w", err)
	}

	logger.Debug().Int("contactID", contactID).Str("label", label).Msg("[CW] label added to contact")
	return nil
}
