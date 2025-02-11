package handler_test

import (
	"context"
	"testing"
	"time"

	connector_v1 "github.com/iBoBoTi/connector-service/gen/proto"
	"github.com/iBoBoTi/connector-service/internal/domain"
	handler "github.com/iBoBoTi/connector-service/internal/transport/grpc"
	"github.com/iBoBoTi/connector-service/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type mockConnectorUsecase struct {
	mock.Mock
}

func (m *mockConnectorUsecase) CreateConnector(ctx context.Context, workspaceID, tenantID, defaultChannel, slackToken string) (*domain.Connector, error) {
	args := m.Called(ctx, workspaceID, tenantID, defaultChannel, slackToken)
	conn := args.Get(0)
	if conn == nil {
		return nil, args.Error(1)
	}
	return conn.(*domain.Connector), args.Error(1)
}
func (m *mockConnectorUsecase) GetConnector(ctx context.Context, connectorID string) (*domain.Connector, error) {
	args := m.Called(ctx, connectorID)
	conn := args.Get(0)
	if conn == nil {
		return nil, args.Error(1)
	}
	return conn.(*domain.Connector), args.Error(1)
}
func (m *mockConnectorUsecase) DeleteConnector(ctx context.Context, connectorID string) error {
	args := m.Called(ctx, connectorID)
	return args.Error(0)
}

func (m *mockConnectorUsecase) SendMessage(ctx context.Context, connectorID, msg string) error {
	args := m.Called(ctx, connectorID, msg)
	return args.Error(0)
}

func TestCreateConnector_Success(t *testing.T) {
	ctx := context.Background()
	mockUC := new(mockConnectorUsecase)

	mockUC.On("CreateConnector", ctx, "ws-1", "tenant-1", "#channel", "token-123").Return(&domain.Connector{
		ID: "conn-123",
	}, nil).Once()

	handler := handler.NewSlackConnectorHandler(mockUC)

	req := &connector_v1.CreateConnectorRequest{
		WorkspaceId:        "ws-1",
		TenantId:           "tenant-1",
		DefaultChannelName: "#channel",
		SlackToken:         "token-123",
	}

	resp, err := handler.CreateConnector(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "conn-123", resp.GetConnector().GetId())
	//require.Equal(t, "Connector created successfully", resp.GetConnector().String())

	mockUC.AssertExpectations(t)
}

func TestCreateConnector_ErrInvalidArgumentError(t *testing.T) {
	ctx := context.Background()
	mockUC := new(mockConnectorUsecase)

	mockUC.
		On("CreateConnector", ctx, "", "", "", "").
		Return(nil, errors.ErrInvalidArgument).
		Once()

	handler := handler.NewSlackConnectorHandler(mockUC)

	req := &connector_v1.CreateConnectorRequest{}
	resp, err := handler.CreateConnector(ctx, req)
	require.Nil(t, resp)
	require.Error(t, err)

	mockUC.AssertExpectations(t)
}

func TestGetConnector_Success(t *testing.T) {
	ctx := context.Background()
	mockUC := new(mockConnectorUsecase)

	mockUC.
		On("GetConnector", ctx, "conn-123").
		Return(&domain.Connector{
			ID:               "conn-123",
			WorkspaceID:      "ws-1",
			TenantID:         "tenant-1",
			DefaultChannelID: "C12345",
			CreatedAt:        time.Unix(1670000000, 0),
			UpdatedAt:        time.Unix(1670001000, 0),
		}, nil).
		Once()

	handler := handler.NewSlackConnectorHandler(mockUC)

	req := &connector_v1.GetConnectorRequest{ConnectorId: "conn-123"}
	resp, err := handler.GetConnector(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "conn-123", resp.GetConnector().GetId())
	require.Equal(t, "C12345", resp.GetConnector().GetDefaultChannelId())
	require.Equal(t, timestamppb.New(time.Unix(1670000000, 0)).String(), resp.GetConnector().GetCreatedAt())
	require.Equal(t, timestamppb.New(time.Unix(1670001000, 0)).String(), resp.GetConnector().GetUpdatedAt())

	mockUC.AssertExpectations(t)
}

func TestGetConnector_NotFound(t *testing.T) {
	ctx := context.Background()
	mockUC := new(mockConnectorUsecase)

	mockUC.
		On("GetConnector", ctx, "does-not-exist").
		Return(nil, errors.ErrNotFound).
		Once()

	handler := handler.NewSlackConnectorHandler(mockUC)
	req := &connector_v1.GetConnectorRequest{ConnectorId: "does-not-exist"}

	resp, err := handler.GetConnector(ctx, req)
	require.Nil(t, resp)
	require.Error(t, err)

	mockUC.AssertExpectations(t)
}

func TestDeleteConnector_Success(t *testing.T) {
	ctx := context.Background()
	mockUC := new(mockConnectorUsecase)

	mockUC.
		On("DeleteConnector", ctx, "conn-123").
		Return(nil).
		Once()

	handler := handler.NewSlackConnectorHandler(mockUC)
	req := &connector_v1.DeleteConnectorRequest{ConnectorId: "conn-123"}
	resp, err := handler.DeleteConnector(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	//require.Equal(t, "Connector deleted successfully", resp.GetMessage())

	mockUC.AssertExpectations(t)
}
