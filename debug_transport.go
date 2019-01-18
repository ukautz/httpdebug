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

// DebugTransport logs any throughgoing request
type DebugTransport struct {
	Transport http.RoundTripper
	Output    io.Writer
	ForceJSON bool
}

// WrapDebugTransport wraps DebugTransport around transport of client
func WrapLogTransport(client *http.Client, output io.Writer) {
	client.Transport = NewDebugTransport(client.Transport, output)
}

// NewDebugTransport wraps provided transport into new logging transport. Optional
// output can be provided, otherwise logging is being used
func NewDebugTransport(transport http.RoundTripper, output io.Writer) *DebugTransport {
	if transport == nil {
		transport = http.DefaultTransport
	}
	return &DebugTransport{
		Transport: transport,
		Output:    output,
	}
}

// RoundTrip implements the http.RoundTripper interface
func (t *DebugTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	requestDump, err := dumpRequest(req, t.ForceJSON)
	if err != nil {
		return nil, err
	}
	t.print(fmt.Sprintf("****** REQUEST START ******\n%s\n****** REQUEST END ******\n", requestDump))

	res, err := t.Transport.RoundTrip(req)
	if err != nil {
		t.print(fmt.Sprintf("\n!!!!!! RESPONSE ERROR !!!!!!\n%s\n!!!!!! RESPONSE ERROR !!!!!!\n", requestDump))
		return nil, err
	}

	if res != nil {
		responseDump := dumpResponse(res, t.ForceJSON)
		t.print(fmt.Sprintf("\n****** RESPONSE START ******\n%s\n****** RESPONSE END ******\n", responseDump))
	} else {
		t.print(fmt.Sprintln("\n~~~~~~ NO RESPONSE ~~~~~~"))
	}

	return res, nil
}

func (t *DebugTransport) print(output string) {
	if t.Output != nil {
		fmt.Fprint(t.Output, output)
	} else {
		log.Print(output)
	}
}

func dumpRequest(req *http.Request, forceJSON bool) ([]byte, error) {
	var jsonBody []byte
	if req.Body != nil && (forceJSON || strings.Contains(req.Header.Get("accept"), "json")) {
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
	return raw, nil
}

func dumpResponse(res *http.Response, forceJSON bool) []byte {
	var jsonBody []byte
	if forceJSON || strings.Contains(res.Header.Get("content-type"), "json") {
		orig, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return []byte(fmt.Sprintf("FAILED to read response body: %s", err))
		}
		jsonBody = decodeJSON(orig)
		res.Body = ioutil.NopCloser(bytes.NewBuffer(orig))
	}
	raw, err := httputil.DumpResponse(res, jsonBody == nil)
	if err != nil {
		return []byte(fmt.Sprintf("FAILED to dump response: %s", err))
	}
	if jsonBody != nil {
		raw = append(raw, '\n')
		raw = append(raw, jsonBody...)
	}
	return raw
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
