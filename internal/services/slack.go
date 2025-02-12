package services

import (
	"context"
	"fmt"

	"github.com/slack-go/slack"

	"github.com/iBoBoTi/connector-service/internal/repository"
)

type SlackClient interface {
	ResolveChannelID(ctx context.Context, token, channelName string) (string, error)
	SendMessage(ctx context.Context, token, channelID, message string) error
}

type slackClient struct{}

func NewSlackClient() SlackClient {
	return &slackClient{}
}

// ResolveChannelID attempts to find a channel with the given name.
// Returns its channel ID if found, otherwise an error.
func (c *slackClient) ResolveChannelID(ctx context.Context, token, channelName string) (string, error) {
	client := slack.New(token)

	params := &slack.GetConversationsParameters{
		Limit:           200,
		ExcludeArchived: true,
		Types:           []string{"public_channel", "private_channel"},
	}

	// Iterate through all channels until we find one matching channelName
	for {
		channels, cursor, err := client.GetConversationsContext(ctx, params)

		if err != nil {
			return "", fmt.Errorf("failed to list slack channels: %w", err)
		}

		for _, ch := range channels {
			if ch.Name == channelName {
				return ch.ID, nil
			}
		}

		if cursor == "" {
			break
		}
		params.Cursor = cursor
	}

	return "", fmt.Errorf("channel '%s' not found", channelName)
}

// SendMessage posts a simple text message to the given Slack channel ID.
func (c *slackClient) SendMessage(ctx context.Context, token, channelID, message string) error {
	client := slack.New(token)

	_, _, err := client.PostMessageContext(ctx, channelID,
		slack.MsgOptionText(message, false),
	)
	if err != nil {
		return fmt.Errorf("failed to send Slack message to channelID=%s: %w", channelID, err)
	}
	return nil
}

func SendMessage(
	ctx context.Context,
	connectorID string,
	message string,
	secretsManager AWSSecretsManager,
	slackClient SlackClient,
	connectorRepository repository.ConnectorRepository,
) error {
	connector, err := connectorRepository.GetByID(ctx, connectorID)
	if err != nil {
		return fmt.Errorf("error getting connector: %w", err)
	}

	token, err := secretsManager.GetSlackToken(ctx, "connector/"+connectorID+"/slack-token")
	if err != nil {
		return fmt.Errorf("error retrieving secret: %w", err)
	}

	if err := slackClient.SendMessage(ctx, token, connector.DefaultChannelID, message); err != nil {
		return fmt.Errorf("error sending message to channel: %w", err)
	}
	
	return nil
}
