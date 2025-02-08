package domain

import (
	"time"
)

type Connector struct {
	ID               string
	TenantID         string
	WorkspaceID      string
	DefaultChannelID string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
