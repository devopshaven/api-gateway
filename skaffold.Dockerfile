FROM golang

ENV GOPROXY=goproxy.pirat.app

ENV GOTRACEBACK=all

WORKDIR /usr/src/app

COPY . .

ARG SKAFFOLD_GO_GCFLAGS
RUN go build -gcflags="${SKAFFOLD_GO_GCFLAGS}" -o /app .

ENTRYPOINT [ "/app" ]
