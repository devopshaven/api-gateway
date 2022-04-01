package gateway

type Version string

const (
	VersionV1 Version = "v1"
)

type GatewayConfig struct {
	Version  Version   `yaml:"version"`
	Services []Service `yaml:"services"`
}

type Service struct {
	Name       string   `yaml:"name,omitempty"`
	ServiceUrl string   `yaml:"serviceUrl"`
	Paths      []string `yaml:"paths"`
}
