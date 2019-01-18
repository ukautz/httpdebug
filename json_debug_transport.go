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

	"github.com/fatih/color"
	"github.com/hokaccha/go-prettyjson"
)

// JSONDebugTransport logs any throughgoing request
type JSONDebugTransport struct {
	Transport http.RoundTripper
	Output    io.Writer
	ForceJSON bool
	Plain     bool
}

var RenderRequestDump = func(dump string) string {
	return color.CyanString("## REQUEST {") + "\n" +
		dump + "\n" +
		color.CyanString("## REQUEST }") + "\n\n"
}

var RenderResponseDump = func(dump string) string {
	return color.YellowString("## RESPONSE {") + "\n" +
		dump + "\n" +
		color.YellowString("## RESPONSE }") + "\n\n"
}

var RenderResponseErrorDump = func(dump string) string {
	return color.HiRedString("## RESPONSE ERROR {") + "\n" +
		dump + "\n" +
		color.HiRedString("## RESPONSE ERROR }") + "\n\n"
}

var RenderNoResponse = func() string {
	return color.HiYellowString("((( NO RESPONSE )))\n\n")
}

// WrapJSONDebugTransport wraps JSONDebugTransport around transport of client
func WrapJSONDebugTransport(client *http.Client, output io.Writer) {
	client.Transport = NewJSONDebugTransport(client.Transport, output)
}

// NewJSONDebugTransport wraps provided transport into new logging transport. Optional
// output can be provided, otherwise logging is being used
func NewJSONDebugTransport(transport http.RoundTripper, output io.Writer) *JSONDebugTransport {
	if transport == nil {
		transport = http.DefaultTransport
	}
	return &JSONDebugTransport{
		Transport: transport,
		Output:    output,
	}
}

// RoundTrip implements the http.RoundTripper interface
func (t *JSONDebugTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	requestDump, err := dumpRequest(req, t.ForceJSON)
	if err != nil {
		return nil, err
	}
	t.print(RenderRequestDump(string(requestDump)))

	res, err := t.Transport.RoundTrip(req)
	if err != nil {
		t.print(RenderResponseErrorDump(string(requestDump)))
		return nil, err
	} else if res != nil {
		responseDump := dumpResponse(res, t.ForceJSON)
		t.print(RenderResponseDump(string(responseDump)))
	} else {
		t.print(RenderNoResponse())
	}

	return res, nil
}

func (t *JSONDebugTransport) print(output string) {
	if t.Output != nil {
		fmt.Fprint(t.Output, output)
	} else {
		log.Print(output)
	}
}

func ForceColor() {
	color.NoColor = false
}

func formatHeaders(raw []byte) []byte {
	lines := strings.Split(string(raw), "\r\n")
	lines[0] = boldMessage(lines[0])
	for i, line := range lines[1:] {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			lines[i+1] = boldMessage(parts[0]+":") + parts[1]
		}
	}
	return []byte(strings.Join(lines, "\r\n"))
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
	raw = formatHeaders(raw)
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
			return []byte(errorMessage("FAILED to read response body: %s", err))
		}
		jsonBody = decodeJSON(orig)
		res.Body = ioutil.NopCloser(bytes.NewBuffer(orig))
	}
	raw, err := httputil.DumpResponse(res, jsonBody == nil)
	if err != nil {
		return []byte(errorMessage("FAILED to dump response: %s", err))
	}
	raw = formatHeaders(raw)
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
		//rendered, err := json.MarshalIndent(attempt, "", "  ")
		rendered, err := prettyjson.Marshal(attempt)
		if err != nil {
			return []byte(errorMessage("invalid JSON (%s):\n--\n%s\n--\n", err, string(raw)))
		}
		return rendered
	}
	return []byte(errorMessage("unparsable JSON:\n--\n%s\n--\n", string(raw)))
}

func boldMessage(msg string, args ...interface{}) string {
	return color.New(color.FgWhite, color.Bold).Sprintf(msg, args...)
}

func errorMessage(msg string, args ...interface{}) string {
	return color.New(color.FgRed, color.Bold).Sprintf(msg, args...)
}
