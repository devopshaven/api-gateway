package gateway

import (
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
)

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

type Gateway struct {
	config *GatewayConfig
}

func NewGateway(config *GatewayConfig) *Gateway {
	return &Gateway{
		config: config,
	}
}

func (p *Gateway) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	log.Trace().Msgf("RemoteAddr: %s Method: %s, URL: %s", req.RemoteAddr, string(req.Method), req.URL.String())

	req.URL.Scheme = "http"

	if req.URL.Scheme != "http" && req.URL.Scheme != "https" {
		msg := "unsupported protocal scheme " + req.URL.Scheme
		http.Error(wr, msg, http.StatusBadRequest)

		log.Error().Msg(msg)
		return
	}

	client := &http.Client{}

	//http: Request.RequestURI can't be set in client requests.
	//http://golang.org/src/pkg/net/http/client.go
	req.RequestURI = ""

	delHopHeaders(req.Header)

	if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
		appendHostToXForwardHeader(req.Header, clientIP)
	}

	service := "127.0.0.1:5001"

	log.Trace().Str("service", service).Msgf("setting url scheme to http and service to: %s", service)
	req.URL.Host = service
	req.URL.Scheme = "http"

	resp, err := client.Do(req)
	if err != nil {
		http.Error(wr, "Server Error", http.StatusInternalServerError)
		log.Fatal().Msgf("ServeHTTP: %s", err)
	}
	defer resp.Body.Close()

	log.Trace().Msgf("RemoteAddr: %s Status: %s", req.RemoteAddr, resp.Status)

	delHopHeaders(resp.Header)

	copyHeader(wr.Header(), resp.Header)
	wr.WriteHeader(resp.StatusCode)
	io.Copy(wr, resp.Body)
}
