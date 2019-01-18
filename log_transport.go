package httpdebug

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"
)

// LogTransport logs any throughgoing request
type LogTransport struct {
	Transport http.RoundTripper
	Output    io.Writer
	ForceJSON bool
}

// WrapLogTransport wraps LogTransport around transport of client
func WrapLogTransport(client *http.Client, output io.Writer) {
	client.Transport = NewLogTransport(client.Transport, output)
}

// NewLogTransport wraps provided transport into new logging transport. Optional
// output can be provided, otherwise logging is being used
func NewLogTransport(transport http.RoundTripper, output io.Writer) *LogTransport {
	if transport == nil {
		transport = http.DefaultTransport
	}
	return &LogTransport{
		Transport: transport,
		Output:    output,
	}
}

// RoundTrip implements the http.RoundTripper interface
func (t *LogTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var jsonBody []byte
	if req.Body != nil && (t.ForceJSON || strings.Contains(req.Header.Get("accept"), "json")) {
		orig, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		jsonBody = decodeJSON(orig)
		req.Body = ioutil.NopCloser(bytes.NewBuffer(orig))
	}
	raw, err := httputil.DumpRequestOut(req, jsonBody == nil)
	if err != nil {
		return nil, err
	}
	if jsonBody != nil {
		raw = append(raw, '\n')
		raw = append(raw, jsonBody...)
	}
	if t.Output != nil {
		fmt.Fprintf(t.Output, "****** REQUEST START ******\n%s\n****** REQUEST END ******\n", raw)
	} else {
		log.Printf("****** REQUEST START ******\n%s\n****** REQUEST END ******\n", raw)
	}
	return t.Transport.RoundTrip(req)
}

func decodeJSON(raw []byte) []byte {
	attempts := []interface{}{
		[]map[string]interface{}{},
		map[string]interface{}{},
		[]map[interface{}]interface{}{},
		map[interface{}]interface{}{},
	}
	for _, attempt := range attempts {
		if err := json.Unmarshal(raw, &attempt); err != nil {
			continue
		}
		rendered, err := json.MarshalIndent(attempt, "", "  ")
		if err != nil {
			return []byte(fmt.Sprintf("invalid JSON (%s):\n--\n%s\n--\n", err, string(raw)))
		}
		return rendered
	}
	return []byte(fmt.Sprintf("unparsable JSON:\n--\n%s\n--\n", string(raw)))
}
