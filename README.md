# DevopsHaven API gateway

Gateway for DevopsHaven ingress

Before you wanna use the gateway please set the RBAC permission to allow to create a configmap watch by the pod to it's own namespace. You can check the [k8s/roles.yaml](k8s/roles.yaml) file for example configuration.

docker image: `hub.pirat.app/api-gateway`
