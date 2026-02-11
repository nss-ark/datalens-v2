package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/complyark/datalens/internal/domain/discovery"
)

// LoadConfig loads AWS configuration from the environment or data source credentials.
func LoadConfig(ctx context.Context, ds *discovery.DataSource) (aws.Config, error) {
	// In a real implementation, we would check ds.Credentials for keys/role
	// For now, we fall back to the default chain (Env vars, profile, IAM role)
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return aws.Config{}, fmt.Errorf("load aws config: %w", err)
	}

	// If region is specified in connection strings or config, set it
	// if region := ds.Config["region"]; region != "" {
	// 	cfg.Region = region.(string)
	// }

	return cfg, nil
}
