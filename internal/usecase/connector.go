package usecase

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/iBoBoTi/connector-service/internal/domain"
	"github.com/iBoBoTi/connector-service/internal/repository"
	"github.com/iBoBoTi/connector-service/internal/services"
	"github.com/iBoBoTi/connector-service/pkg/errors"
)

type ConnectorUsecase interface {
	CreateConnector(ctx context.Context, workspaceID, tenantID, defaultChannel, slackToken string) (*domain.Connector, error)
	GetConnector(ctx context.Context, connectorID string) (*domain.Connector, error)
	DeleteConnector(ctx context.Context, connectorID string) error
	SendMessage(ctx context.Context, connectorID, msg string) error
}

type connectorUsecase struct {
	repo    repository.ConnectorRepository
	secrets services.AWSSecretsManager
	slack   services.SlackClient
}

// NewConnectorUsecase creates a new ConnectorService.
func NewConnectorUsecase(
	repo repository.ConnectorRepository,
	secrets services.AWSSecretsManager,
	slack services.SlackClient,
) ConnectorUsecase {
	return &connectorUsecase{
		repo:    repo,
		secrets: secrets,
		slack:   slack,
	}
}

// CreateConnector coordinates creating a new connector in DB and storing the Slack token in Secrets Manager.
func (s *connectorUsecase) CreateConnector(
	ctx context.Context,
	workspaceID, tenantID, channelName, slackToken string,
) (*domain.Connector, error) {
	connID := uuid.NewString()

	if slackToken == "" || channelName == "" || workspaceID == "" || tenantID == "" {
		return nil, errors.ErrInvalidArgument
	}

	if err := s.secrets.StoreSlackToken(ctx, connID, slackToken); err != nil {
		slog.Error("error storing slack token", "error", err)
		return nil, errors.ErrInternal
	}

	channelID, err := s.slack.ResolveChannelID(ctx, slackToken, channelName)
	if err != nil {
		slog.Error("error resolving channel id using the channel name", "error", err)
		return nil, errors.ErrInvalidArgument
	}

	now := time.Now()

	connector := &domain.Connector{
		ID:               connID,
		WorkspaceID:      workspaceID,
		TenantID:         tenantID,
		DefaultChannelID: channelID,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := s.repo.Create(ctx, connector); err != nil {
		slog.Error("error creating connector", "error", err)
		return nil, errors.ErrInternal
	}

	return connector, nil
}

// GetConnector retrieves the connector data from the repository.
func (s *connectorUsecase) GetConnector(ctx context.Context, connectorID string) (*domain.Connector, error) {
	connector, err := s.repo.GetByID(ctx, connectorID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		slog.Error("error getting connector by id", "error", err)
		return nil, errors.ErrInternal
	}
	return connector, nil
}

// DeleteConnector removes the connector from DB and the Slack token from Secrets Manager.
func (s *connectorUsecase) DeleteConnector(ctx context.Context, connectorID string) error {
	if err := s.repo.Delete(ctx, connectorID); err != nil {
		slog.Error("error deleting connector", "error", err)
		return errors.ErrInternal
	}

	if err := s.secrets.DeleteSlackToken(ctx, connectorID); err != nil {
		slog.Error("error deleting slack token", "error", err)
		return errors.ErrInternal
	}

	return nil
}

func (u *connectorUsecase) SendMessage(ctx context.Context, connectorID, msg string) error {
	if connectorID == "" || msg == "" {
		return errors.ErrInvalidArgument
	}

	conn, err := u.repo.GetByID(ctx, connectorID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.ErrNotFound
		}
		slog.Error("error getting connector by id", "error", err)
		return errors.ErrInternal
	}

	// Retrieve secret
	secretName := "connector/" + connectorID
	token, err := u.secrets.GetSlackToken(ctx, secretName)
	if err != nil {
		slog.Error("error getting slack token from secret manager", "error", err)
		return errors.ErrInternal
	}

	if err := u.slack.SendMessage(ctx, token, conn.DefaultChannelID, msg); err != nil {
		slog.Error("error sending slack message", "error", err)
		return errors.ErrInternal
	}

	return nil
}
