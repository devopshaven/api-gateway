.PHONY: all

all: build package

clean:
	rm -Rf bin

build:
	GOOS=linux go build -a -tags netgo -ldflags "-w -s" -o bin/gateway_linux .

package: build
	docker build -t hub.pirat.app/api-gateway -f fast.Dockerfile .

deploy:
	docker push hub.pirat.app/api-gateway

test:
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...
	codecov

run:
	go run .
