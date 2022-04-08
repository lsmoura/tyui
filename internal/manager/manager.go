package manager

import (
	"context"
	"github.com/lsmoura/tyui/internal/model"
	"github.com/lsmoura/tyui/pkg/database"
	"github.com/pkg/errors"
)

type Manager struct {
	db *database.DB
}

func New(db *database.DB) *Manager {
	return &Manager{
		db: db,
	}
}

func (m *Manager) LinkWithToken(ctx context.Context, token string) (*model.Links, error) {
	var link model.Links
	rows, err := m.db.QueryContext(ctx, "SELECT id, token, url, created_at, clicks FROM links WHERE token = $1", token)
	if err != nil {
		return nil, errors.Wrap(err, "m.db.QueryContext")
	}
	defer rows.Close()
	if rows.Next() {
		if err := rows.Scan(&link.ID, &link.Token, &link.URL, &link.CreatedAt, &link.Clicks); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
	} else {
		return nil, errors.New("not found")
	}

	return &link, nil
}
