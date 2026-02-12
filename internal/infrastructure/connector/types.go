package connector

import (
	"context"

	"github.com/complyark/datalens/internal/domain/discovery"
)

// ScannableConnector is an optional interface for connectors that support
// streaming or custom scanning logic (push-based) rather than the default
// pull-based (schema -> fields -> sample) model.
type ScannableConnector interface {
	Scan(ctx context.Context, ds *discovery.DataSource, callback func(discovery.PIIClassification)) error
}
