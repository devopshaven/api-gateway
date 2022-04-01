package main

import (
	"flag"
	"net/http"

	"github.com/rs/zerolog/log"
	"go.pirat.app/api-gateway/gateway"
)

func main() {
	client := gateway.NewConfigClient()
	client.StartWatcher()

	var addr = flag.String("addr", "127.0.0.1:8080", "The addr of the application.")
	flag.Parse()

	handler := &gateway.Gateway{}

	log.Info().Msgf("Starting proxy server on %s", *addr)
	if err := http.ListenAndServe(*addr, handler); err != nil {
		log.Fatal().Err(err).Msgf("ListenAndServe: %s", err.Error())
	}
}
