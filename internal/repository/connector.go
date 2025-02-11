package repository

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq"

	"github.com/iBoBoTi/connector-service/internal/domain"
)

type ConnectorRepository interface {
	Create(ctx context.Context, c *domain.Connector) error
	GetByID(ctx context.Context, id string) (*domain.Connector, error)
	Delete(ctx context.Context, id string) error
}

type connectorRepository struct {
	db *sql.DB
}

func NewConnectorRepository(db *sql.DB) ConnectorRepository {
	return &connectorRepository{db: db}
}

func (cr *connectorRepository) Create(ctx context.Context, c *domain.Connector) error {
	if _, err := cr.db.ExecContext(ctx, `
        INSERT INTO connectors (id, tenant_id, workspace_id, default_channel_id, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6)
    `, c.ID, c.TenantID, c.WorkspaceID, c.DefaultChannelID, c.CreatedAt, c.UpdatedAt); err != nil {
		return err
	}

	return nil
}

func (cr *connectorRepository) GetByID(ctx context.Context, id string) (*domain.Connector, error) {
	row := cr.db.QueryRowContext(ctx, `
        SELECT id, tenant_id, workspace_id, default_channel_id, created_at, updated_at
        FROM connectors WHERE id = $1
    `, id)
	var c domain.Connector
	if err := row.Scan(&c.ID, &c.TenantID, &c.WorkspaceID, &c.DefaultChannelID, &c.CreatedAt, &c.UpdatedAt); err != nil {
		return nil, err
	}
	return &c, nil
}

func (cr *connectorRepository) Delete(ctx context.Context, id string) error {
	_, err := cr.db.ExecContext(ctx, `DELETE FROM connectors WHERE id = $1`, id)
	return err
}
