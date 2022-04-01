package gateway_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.pirat.app/api-gateway/gateway"
	"gopkg.in/yaml.v3"
)

func TestCreateConfig(t *testing.T) {
	bb, err := yaml.Marshal(&gateway.GatewayConfig{
		Version: gateway.VersionV1,
		Services: []gateway.Service{{
			Name:       "card-service",
			ServiceUrl: "card-service",
			Paths: []string{
				"/api/v1/card",
				"/api/card/v1",
			},
		}},
	})

	assert.NoError(t, err)

	fmt.Println(string(bb))
}
