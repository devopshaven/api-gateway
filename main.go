package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.pirat.app/api-gateway/gateway"
)

var (
	addr              *string
	authServerAddress *string
	prettyLog         *bool
)

func init() {
	addr = flag.String("addr", "127.0.0.1:8080", "The addr of the application.")
	authServerAddress = flag.String("authServer", "", "authorization server address")
	prettyLog = flag.Bool("pretty", false, "use developer friendly log")
}

func main() {
	flag.Parse()

	if *prettyLog {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout}).With().Caller().Logger()
	}

	gateway := gateway.NewGateway(*authServerAddress)

	otelHandler := otelhttp.NewHandler(gateway, "gateway")

	srv := &http.Server{
		Addr:    *addr,
		Handler: otelHandler,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Msgf("listen: %s\n", err)
		}
	}()

	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	log.Info().Msgf("Server is running with %d goroutines with allocated heap: %0.2fMB",
		runtime.NumGoroutine(),
		float64(ms.TotalAlloc)/1024/1024,
	)

	// Wait for terminate/interrupt signals
	<-done

	gateway.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Msgf("Server Shutdown Failed: %+v", err)
	}

	log.Info().Msg("Server Exited Properly")
}
