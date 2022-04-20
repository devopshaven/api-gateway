[![codecov](https://codecov.io/gh/devopshaven/api-gateway/branch/master/graph/badge.svg?token=ZD5AC3QTUW)](https://codecov.io/gh/devopshaven/api-gateway)
[![Go Reference](https://pkg.go.dev/badge/github.com/devopshaven/api-gateway.svg)](https://pkg.go.dev/github.com/devopshaven/api-gateway)
# DevopsHaven API gateway

Gateway for DevopsHaven ingress

Before you wanna use the gateway please set the RBAC permission to allow to create a configmap watch by the pod to it's own namespace. You can check the [k8s/roles.yaml](k8s/roles.yaml) file for example configuration.

Docker image: `hub.pirat.app/api-gateway`

### Tracing

The server implements tracing which uses B3 and W3C header propagators with [OpenTelemetry](https://opentelemetry.io/) standards. 

## Environment variables:

- `OTEL_EXPORTER_JAEGER_AGENT_HOST` the jaeger agent host jaeger exporter
- `OTEL_EXPORTER_JAEGER_AGENT_PORT` the jaeger agent port for jaeger exporter
