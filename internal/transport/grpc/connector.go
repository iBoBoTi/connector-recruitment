package handler

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"

	connector_v1 "github.com/iBoBoTi/connector-service/gen/proto"
	"github.com/iBoBoTi/connector-service/internal/domain"
	"github.com/iBoBoTi/connector-service/usecase"
)

// SlackConnectorHandler implements connector.v1.SlackConnectorServiceServer.
type SlackConnectorHandler struct {
	connUsecase *usecase.ConnectorUsecase
	connector_v1.UnimplementedSlackConnectorServiceServer
}

// NewSlackConnectorHandler constructs a new gRPC handler instance.
func NewSlackConnectorHandler(connUC *usecase.ConnectorUsecase) *SlackConnectorHandler {
	return &SlackConnectorHandler{connUsecase: connUC}
}

func (h *SlackConnectorHandler) CreateConnector(
	ctx context.Context,
	req *connector_v1.CreateConnectorRequest,
) (*connector_v1.CreateConnectorResponse, error) {
	conn, err := h.connUsecase.CreateConnector(
		ctx,
		req.WorkspaceId,
		req.TenantId,
		req.DefaultChannelName,
		req.SlackToken,
	)
	if err != nil {
		return nil, err
	}

	return &connector_v1.CreateConnectorResponse{
		Connector: toProtoConnector(conn),
	}, nil
}

func (h *SlackConnectorHandler) GetConnector(
	ctx context.Context,
	req *connector_v1.GetConnectorRequest,
) (*connector_v1.GetConnectorResponse, error) {
	conn, err := h.connUsecase.GetConnector(ctx, req.ConnectorId)
	if err != nil {
		return nil, err
	}
	return &connector_v1.GetConnectorResponse{
		Connector: toProtoConnector(conn),
	}, nil
}

func (h *SlackConnectorHandler) DeleteConnector(
	ctx context.Context,
	req *connector_v1.DeleteConnectorRequest,
) (*connector_v1.DeleteConnectorResponse, error) {
	if err := h.connUsecase.DeleteConnector(ctx, req.ConnectorId); err != nil {
		return nil, err
	}
	return &connector_v1.DeleteConnectorResponse{
		Success: true,
	}, nil
}

func toProtoConnector(c *domain.Connector) *connector_v1.Connector {
	return &connector_v1.Connector{
		Id:               c.ID,
		WorkspaceId:      c.WorkspaceID,
		TenantId:         c.TenantID,
		DefaultChannelId: c.DefaultChannelID,
		CreatedAt:        timestamppb.New(c.CreatedAt).String(),
		UpdatedAt:        timestamppb.New(c.UpdatedAt).String(),
	}
}
