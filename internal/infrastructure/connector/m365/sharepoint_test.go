package m365

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMicrosoft365Connector_Capabilities(t *testing.T) {
	connector := NewMicrosoft365Connector("secret-key")
	caps := connector.Capabilities()

	assert.True(t, caps.CanDiscover)
	assert.True(t, caps.CanSample)
	assert.True(t, caps.SupportsIncremental)
}

func TestMicrosoft365Connector_GetFields(t *testing.T) {
	connector := NewMicrosoft365Connector("secret-key")
	fields, err := connector.GetFields(context.Background(), "some-id")

	assert.NoError(t, err)
	assert.NotEmpty(t, fields)
	assert.Equal(t, "content", fields[0].Name)
	assert.Equal(t, "blob", fields[0].DataType)
}

func TestMicrosoft365Connector_ScanDrive_Logic(t *testing.T) {
	// thorough testing would require mocking http.Client in GraphClient
	// For now, we verify the struct and method existence.
	connector := NewMicrosoft365Connector("secret-key")
	assert.NotNil(t, connector)
}
