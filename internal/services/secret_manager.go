package services

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

// SecretsManager defines the methods for storing and retrieving secrets (e.g., Slack tokens).
type AWSSecretsManager interface {
	StoreSlackToken(ctx context.Context, connectorID, token string) error
	GetSlackToken(ctx context.Context, connectorID string) (string, error)
	DeleteSlackToken(ctx context.Context, connectorID string) error
}

type awsSecretManager struct {
	sm *secretsmanager.SecretsManager
}

func NewSecretsManager(sess *session.Session) AWSSecretsManager {
	return &awsSecretManager{
		sm: secretsmanager.New(sess),
	}
}

func (s *awsSecretManager) StoreSlackToken(ctx context.Context, connectorID, token string) error {
	secretName := slackSecretName(connectorID)
	_, err := s.sm.CreateSecretWithContext(ctx, &secretsmanager.CreateSecretInput{
		Name:         aws.String(secretName),
		SecretString: aws.String(token),
	})
	if err != nil {
		_, updateErr := s.sm.UpdateSecretWithContext(ctx, &secretsmanager.UpdateSecretInput{
			SecretId:     aws.String(secretName),
			SecretString: aws.String(token),
		})
		if updateErr != nil {
			return fmt.Errorf("failed to create or update secret: %w", updateErr)
		}
	}
	return nil
}

func (s *awsSecretManager) GetSlackToken(ctx context.Context, connectorID string) (string, error) {
	secretName := slackSecretName(connectorID)
	out, err := s.sm.GetSecretValueWithContext(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	})
	if err != nil {
		return "", fmt.Errorf("failed to retrieve secret: %w", err)
	}
	return aws.StringValue(out.SecretString), nil
}

func (s *awsSecretManager) DeleteSlackToken(ctx context.Context, connectorID string) error {
	secretName := slackSecretName(connectorID)
	_, err := s.sm.DeleteSecretWithContext(ctx, &secretsmanager.DeleteSecretInput{
		SecretId:                   aws.String(secretName),
		ForceDeleteWithoutRecovery: aws.Bool(true),
	})
	if err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}
	return nil
}

func slackSecretName(connectorID string) string {
	return fmt.Sprintf("slack-connector/%s", connectorID)
}
