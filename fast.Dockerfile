FROM alpine:latest

COPY bin/gateway_linux /app

ENTRYPOINT [ "/app" ]
