package gateway_test

import (
	_ "embed"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.pirat.app/api-gateway/gateway"
	"gopkg.in/yaml.v2"
)

//go:embed test/config_ok.yaml
var validConfig []byte

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

	var conf gateway.GatewayConfig
	err = yaml.Unmarshal(validConfig, &conf)
	assert.NoError(t, err)
	assert.Equal(t, conf.Services[0].ServiceUrl, "srv-url")

	fmt.Println(string(bb))
}
