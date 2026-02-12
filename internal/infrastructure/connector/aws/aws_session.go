package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/complyark/datalens/internal/domain/discovery"
	"github.com/complyark/datalens/internal/infrastructure/connector/shared"
)

// LoadConfig loads AWS configuration from the environment or data source credentials.
func LoadConfig(ctx context.Context, ds *discovery.DataSource) (aws.Config, error) {
	// Parse credentials
	creds, err := shared.ParseCredentials(ds.Credentials)
	if err != nil {
		return aws.Config{}, fmt.Errorf("parse credentials: %w", err)
	}

	opts := []func(*config.LoadOptions) error{}

	// If region is specified in config or credentials
	// Prioritize config, then credentials
	// If region is specified in config or credentials
	// Prioritize config, then credentials
	// But in entity definition: Config string `json:"config" db:"config"`
	// We need to parse ds.Config if we want to use it.
	// For now, let's look in credentials map for region as well, or just rely on default chain if not found.
	if r, ok := creds["region"].(string); ok && r != "" {
		opts = append(opts, config.WithRegion(r))
	}

	// Static credentials
	id, ok1 := creds["access_key_id"].(string)
	secret, ok2 := creds["secret_access_key"].(string)
	token, _ := creds["session_token"].(string)

	if ok1 && ok2 {
		opts = append(opts, config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(id, secret, token)))
	}

	return config.LoadDefaultConfig(ctx, opts...)
}
