package gateway

import (
	_ "embed"
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"
	"strings"

	authservice "github.com/devopshaven/gateway-auth-service"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
)

//go:embed htmlerror.html
var errorTemplate string

// Hop-by-hop headers. These are removed when sent to the backend.
// http://www.w3.org/Protocols/rfc2616/rfc2616-sec13.html
var hopHeaders = []string{
	"Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te", // canonicalized version of "TE"
	"Trailers",
	"Transfer-Encoding",
	"Upgrade",
}

type Gateway struct {
	configClient *ConfigClient
	authClient   *authservice.Client
}

func NewGateway() *Gateway {
	client := NewConfigClient()
	client.StartWatcher()

	authClient := authservice.NewClient("127.0.0.1:5009")

	InitTelemetry()

	return &Gateway{
		configClient: client,
		authClient:   authClient,
	}
}

func (g *Gateway) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	log.Trace().Msgf("RemoteAddr: %s Method: %s, URL: %s", req.RemoteAddr, string(req.Method), req.URL.String())

	req.URL.Scheme = "http"

	if req.URL.Scheme != "http" && req.URL.Scheme != "https" {
		msg := "unsupported protocal scheme " + req.URL.Scheme
		http.Error(wr, msg, http.StatusBadRequest)

		log.Error().Msg(msg)
		return
	}

	client := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}

	//http: Request.RequestURI can't be set in client requests.
	//http://golang.org/src/pkg/net/http/client.go
	req.RequestURI = ""

	delHopHeaders(req.Header)

	if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
		appendHostToXForwardHeader(req.Header, clientIP)
	}

	if !strings.HasPrefix(req.URL.Path, "/api") {
		http.Error(wr, "the api only forwards /api requests", http.StatusBadRequest)

		return
	}

	service := "no-service"
	serviceName := "no-service"

SERVICES:
	for _, s := range g.configClient.config.Services {
		for _, p := range s.Paths {
			if strings.HasPrefix(req.URL.Path, p) {
				log.Trace().Interface("srv", s).Str("url", req.URL.Path).Msgf("service hit: %s", s.Name)
				service = s.ServiceUrl
				serviceName = s.Name

				break SERVICES
			}
		}
	}

	if service == "no-service" {
		http.Error(wr, "no service defined", http.StatusInternalServerError)

		return
	}

	authResult, err := g.authClient.Authorize(
		req.Context(),
		req.Method,
		req.Host,
		req.URL.Path,
		req.Header,
	)

	if err != nil {
		renderError(wr, fmt.Sprintf("auth error: %s", err.Error()))

		return
	}

	if authResult.Block {
		authResult.RenderError(wr)

		return
	} else {
		authResult.AddHeaders(req.Header)
	}

	log.Trace().Str("service", service).Msgf("setting url scheme to http and service to: %s", service)

	req.URL.Host = service
	req.URL.Scheme = "http"

	_, span := otel.Tracer(serviceName).Start(
		req.Context(),
		fmt.Sprintf("%s %s", strings.ToUpper(req.Method), req.URL.Path),
	)

	span.SetAttributes(
		attribute.String("service", service),
	)
	defer span.End()

	wr.Header().Set("x-service", serviceName)

	req = req.WithContext(baggage.ContextWithoutBaggage(req.Context()))

	// Make a call to the upstream
	resp, err := client.Do(req)
	if err != nil {
		renderError(wr, fmt.Sprintf("error while calling upstream: %s", err.Error()))
		log.Error().Err(err).Msg("upstream call failed")
		span.RecordError(err)
		span.SetAttributes(
			attribute.Bool("error", true),
		)

		return
	}
	defer resp.Body.Close()

	log.Trace().Msgf("RemoteAddr: %s Status: %s", req.RemoteAddr, resp.Status)
	span.SetAttributes(
		attribute.String("status", resp.Status),
		attribute.String("remote_address", req.RemoteAddr),
	)

	delHopHeaders(resp.Header)

	copyHeader(wr.Header(), resp.Header)
	wr.WriteHeader(resp.StatusCode)
	io.Copy(wr, resp.Body)
}

func renderError(w http.ResponseWriter, message string, err ...error) {
	w.WriteHeader(http.StatusBadGateway)
	w.Header().Set("Content-Type", "text/html")

	if message == "" {
		message = "error while completing request"
	}

	templates, _ := template.New("error").Parse(errorTemplate)
	context := map[string]interface{}{
		"Error":     message,
		"Version":   "v1.0.1",
		"RequestID": uuid.NewString(),
	}
	templates.Execute(w, context)
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func delHopHeaders(header http.Header) {
	for _, h := range hopHeaders {
		header.Del(h)
	}
}

func appendHostToXForwardHeader(header http.Header, host string) {
	// If we aren't the first proxy retain prior
	// X-Forwarded-For information as a comma+space
	// separated list and fold multiple headers into one.
	if prior, ok := header["X-Forwarded-For"]; ok {
		host = strings.Join(prior, ", ") + ", " + host
	}
	header.Set("X-Forwarded-For", host)
}
