package gateway_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.pirat.app/api-gateway/gateway"
)

func TestConfigClient(t *testing.T) {
	cc := gateway.NewConfigClient()

	// Start watcher
	cc.StartWatcher()

	// Wait for goroutine
	time.Sleep(time.Second * 2)

	cfg := cc.Config()
	assert.NotNil(t, cfg)

	err := cc.Close()
	assert.NoError(t, err)
}
