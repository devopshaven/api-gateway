FROM golang:1.18 as build

WORKDIR /usr/src/app

LABEL org.opencontainers.image.source = "https://github.com/devopshaven/api-gateway"
LABEL maintainer="Gyula Paal <paalgyula@paalgyula.com>"

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the project
COPY . .

RUN CGO_ENABLED=0 go build -tags netgo -ldflags="-s -w" -o /app .

# Create application container from alpine (with CA certs)
FROM alpine:latest
COPY --from=build /app /app

# Do not run app in privileged mode
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

CMD [ "/app" ]
