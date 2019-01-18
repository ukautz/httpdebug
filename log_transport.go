package httpdebug

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
)

// LogTransport logs any throughgoing request
type LogTransport struct {
	Transport http.RoundTripper
	Output    io.Writer
}

// WrapLogTransport wraps LogTransport around transport of client
func WrapLogTransport(client *http.Client, output io.Writer) {
	transport := client.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}
	client.Transport = NewLogTransport(transport, output)
}

// NewLogTransport wraps provided transport into new logging transport. Optional
// output can be provided, otherwise logging is being used
func NewLogTransport(transport http.RoundTripper, output io.Writer) *LogTransport {
	return &LogTransport{
		Transport: transport,
		Output:    output,
	}
}

// RoundTrip implements the http.RoundTripper interface
func (t *LogTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	raw, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		return nil, err
	}
	if t.Output != nil {
		fmt.Fprintf(t.Output, "***** REQUEST START ******\n%s\n***** REQUEST END ******\n", string(raw))
	} else {
		log.Printf("***** REQUEST START ******\n%s\n***** REQUEST ******\n", string(raw))

	}
	return t.Transport.RoundTrip(req)
}
