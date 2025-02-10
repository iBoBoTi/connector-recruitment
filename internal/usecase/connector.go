package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/iBoBoTi/connector-service/internal/domain"
	"github.com/iBoBoTi/connector-service/internal/repository"
	"github.com/iBoBoTi/connector-service/internal/services"
)

type ConnectorUsecase struct {
	repo    repository.ConnectorRepository
	secrets services.AWSSecretsManager
	slack   services.SlackClient
}

// NewConnectorUsecase creates a new ConnectorService.
func NewConnectorUsecase(
	repo repository.ConnectorRepository,
	secrets services.AWSSecretsManager,
	slack services.SlackClient,
) *ConnectorUsecase {
	return &ConnectorUsecase{
		repo:    repo,
		secrets: secrets,
		slack:   slack,
	}
}

// CreateConnector coordinates creating a new connector in DB and storing the Slack token in Secrets Manager.
func (s *ConnectorUsecase) CreateConnector(
	ctx context.Context,
	workspaceID, tenantID, channelName, slackToken string,
) (*domain.Connector, error) {
	connID := uuid.NewString()

	if err := s.secrets.StoreSlackToken(ctx, connID, slackToken); err != nil {
		return nil, fmt.Errorf("error storing slack token %w", err)
	}

	channelID, err := s.slack.ResolveChannelID(ctx, slackToken, channelName)
	if err != nil {
		return nil, fmt.Errorf("error resolving channel id %w", err)
	}

	now := time.Now()

	conn := &domain.Connector{
		ID:               connID,
		WorkspaceID:      workspaceID,
		TenantID:         tenantID,
		DefaultChannelID: channelID,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := s.repo.Create(ctx, conn); err != nil {
		return nil, fmt.Errorf("error creating connector %w", err)
	}

	return conn, nil
}

// GetConnector retrieves the connector data from the repository.
func (s *ConnectorUsecase) GetConnector(ctx context.Context, connectorID string) (*domain.Connector, error) {
	return s.repo.GetByID(ctx, connectorID)
}

// DeleteConnector removes the connector from DB and the Slack token from Secrets Manager.
func (s *ConnectorUsecase) DeleteConnector(ctx context.Context, connectorID string) error {
	if err := s.repo.Delete(ctx, connectorID); err != nil {
		return fmt.Errorf("error deleting connector %w", err)
	}

	return s.secrets.DeleteSlackToken(ctx, connectorID)
}
