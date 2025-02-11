package usecase_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/iBoBoTi/connector-service/internal/domain"
	"github.com/iBoBoTi/connector-service/internal/usecase"
	"github.com/iBoBoTi/connector-service/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockConnectorRepository struct {
	mock.Mock
}

func (m *mockConnectorRepository) Create(ctx context.Context, c *domain.Connector) error {
	args := m.Called(ctx, c)
	return args.Error(0)
}

func (m *mockConnectorRepository) GetByID(ctx context.Context, connectorID string) (*domain.Connector, error) {
	args := m.Called(ctx, connectorID)
	conn := args.Get(0)
	if conn == nil {
		return nil, args.Error(1)
	}
	return conn.(*domain.Connector), args.Error(1)
}

func (m *mockConnectorRepository) Delete(ctx context.Context, connectorID string) error {
	args := m.Called(ctx, connectorID)
	return args.Error(0)
}

type mockSecretsManager struct {
	mock.Mock
}

func (m *mockSecretsManager) StoreSlackToken(ctx context.Context, connectorID, token string) error {
	args := m.Called(ctx, connectorID, token)
	return args.Error(0)
}

func (m *mockSecretsManager) GetSlackToken(ctx context.Context, connectorID string) (string, error) {
	args := m.Called(ctx, connectorID)
	return args.String(0), args.Error(1)
}

func (m *mockSecretsManager) DeleteSlackToken(ctx context.Context, connectorID string) error {
	args := m.Called(ctx, connectorID)
	return args.Error(0)
}

type mockSlackClient struct {
	mock.Mock
}

func (m *mockSlackClient) ResolveChannelID(ctx context.Context, token, channelName string) (string, error) {
	args := m.Called(ctx, token, channelName)
	return args.String(0), args.Error(1)
}

func (m *mockSlackClient) SendMessage(ctx context.Context, token, channelID, message string) error {
	args := m.Called(ctx, token, channelID, message)
	return args.Error(0)
}

func TestCreateConnector_Success(t *testing.T) {
	ctx := context.Background()

	mockRepo := new(mockConnectorRepository)
	mockSecrets := new(mockSecretsManager)
	mockSlack := new(mockSlackClient)

	u := usecase.NewConnectorUsecase(mockRepo, mockSecrets, mockSlack)

	mockSecrets.On("StoreSlackToken", ctx, mock.AnythingOfType("string"), "dummy-token").Return(nil).Once()

	mockSlack.On("ResolveChannelID", ctx, "dummy-token", "#general").Return("C123456", nil).Once()

	mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.Connector")).Run(func(args mock.Arguments) {
		connArg := args.Get(1).(*domain.Connector)
		connArg.ID = "123"
	}).
		Return(nil).
		Once()

	connector, err := u.CreateConnector(ctx, "workspace-1", "tenant-1", "#general", "dummy-token")

	require.NoError(t, err)
	require.Equal(t, "123", connector.ID)

	mockSlack.AssertExpectations(t)
	mockSecrets.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestCreateConnector_InvalidArguments(t *testing.T) {
	ctx := context.Background()

	u := usecase.NewConnectorUsecase(nil, nil, nil)

	id, err := u.CreateConnector(ctx, "", "", "", "")
	require.Empty(t, id)
	require.ErrorIs(t, err, errors.ErrInvalidArgument)
}

func TestGetConnector_Success(t *testing.T) {
	ctx := context.Background()

	mockRepo := new(mockConnectorRepository)
	mockSecrets := new(mockSecretsManager)
	mockSlack := new(mockSlackClient)

	u := usecase.NewConnectorUsecase(mockRepo, mockSecrets, mockSlack)

	mockRepo.On("GetByID", ctx, "conn-123").
		Return(&domain.Connector{
			ID: "conn-123",
		}, nil).
		Once()

	_, err := u.GetConnector(ctx, "conn-123")
	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestGetConnector_NotFound(t *testing.T) {
	ctx := context.Background()

	mockRepo := new(mockConnectorRepository)
	u := usecase.NewConnectorUsecase(mockRepo, nil, nil)

	mockRepo.On("GetByID", ctx, "does-not-exist").Return(nil, sql.ErrNoRows).Once()

	conn, err := u.GetConnector(ctx, "does-not-exist")
	require.Nil(t, conn)
	require.ErrorIs(t, err, errors.ErrNotFound)
	mockRepo.AssertExpectations(t)
}

func TestDeleteConnector_Success(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mockConnectorRepository)
	mockSecrets := new(mockSecretsManager)
	mockSlack := new(mockSlackClient)

	u := usecase.NewConnectorUsecase(mockRepo, mockSecrets, mockSlack)

	mockRepo.
		On("Delete", ctx, "conn-123").
		Return(nil).
		Once()

	mockSecrets.
		On("DeleteSlackToken", ctx, "conn-123").
		Return(nil).
		Once()

	err := u.DeleteConnector(ctx, "conn-123")
	require.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockSecrets.AssertExpectations(t)
}

func TestDeleteConnector_InternalError(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mockConnectorRepository)
	mockSecrets := new(mockSecretsManager)
	mockSlack := new(mockSlackClient)

	u := usecase.NewConnectorUsecase(mockRepo, mockSecrets, mockSlack)

	mockRepo.On("Delete", ctx, "conn-123").Return(fmt.Errorf("error deleting connector")).Once()

	err := u.DeleteConnector(ctx, "conn-123")
	require.ErrorIs(t, err, errors.ErrInternal)
}

func TestSendMessage_Success(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mockConnectorRepository)
	mockSecrets := new(mockSecretsManager)
	mockSlack := new(mockSlackClient)

	u := usecase.NewConnectorUsecase(mockRepo, mockSecrets, mockSlack)

	mockRepo.
		On("GetByID", ctx, "conn-123").
		Return(&domain.Connector{
			ID:               "conn-123",
			DefaultChannelID: "C123456",
		}, nil).
		Once()
	mockSecrets.
		On("GetSlackToken", ctx, "connector/conn-123").
		Return("dummy-token", nil).
		Once()
	mockSlack.
		On("SendMessage", ctx, "dummy-token", "C123456", "Hello from test").
		Return(nil).
		Once()

	err := u.SendMessage(ctx, "conn-123", "Hello from test")
	require.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockSecrets.AssertExpectations(t)
	mockSlack.AssertExpectations(t)
}
