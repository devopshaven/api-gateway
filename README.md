# DevopsHaven API gateway

Gateway for DevopsHaven ingress

Before you wanna use the gateway please set the RBAC permission to allow to create a configmap watch by the pod to it's own namespace. You can check the [k8s/roles.yaml](k8s/roles.yaml) file for example configuration.

Docker image: `hub.pirat.app/api-gateway`

### Tracing

The server implements tracing which uses B3 and W3C header propagators with [OpenTelemetry](https://opentelemetry.io/) standards. 

## Environment variables:

- `OTEL_EXPORTER_JAEGER_AGENT_HOST` the jaeger agent host jaeger exporter
- `OTEL_EXPORTER_JAEGER_AGENT_PORT` the jaeger agent port for jaeger exporter
