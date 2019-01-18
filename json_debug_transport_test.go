package httpdebug

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testTransport struct {
	body []byte
}

func (t *testTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	res := &http.Response{
		Status:     "200 OK",
		StatusCode: http.StatusOK,
		Proto:      "HTTP/1.0",
		ProtoMajor: 1,
		ProtoMinor: 0,
		Header:     http.Header{},
	}
	if t.body != nil {
		res.Header.Set("content-type", "text/json")
		res.Body = ioutil.NopCloser(bytes.NewBuffer(t.body))
	} else {
		res.Header.Set("content-type", "text/plain")
	}
	return res, nil
}

var testJSONDebugTransportJSONOut = `****** REQUEST START ******
GET /baz HTTP/1.1
Host: foo.bar
User-Agent: Go-http-client/1.1
Content-Length: 13
Accept: text/json
Accept-Encoding: gzip


{
  "foo": "bar"
}
****** REQUEST END ******

****** RESPONSE START ******
HTTP/1.0 200 OK
Content-Type: text/plain
Content-Length: 0


****** RESPONSE END ******
`

func TestJSONDebugTransport(t *testing.T) {
	out := bytes.NewBuffer(nil)
	in := bytes.NewBufferString(`{"foo":"bar"}`)
	transport := NewJSONDebugTransport(&testTransport{}, out)
	req, err := http.NewRequest("GET", "http://foo.bar/baz", in)
	req.Header.Add("accept", "text/json")
	assert.NoError(t, err)
	_, err = transport.RoundTrip(req)
	assert.NoError(t, err)
	str := stripColors(normalizeLineBreak(out.String()))
	assert.Equal(t, testJSONDebugTransportJSONOut, str)
}

var testJSONDebugTransportNoJSONOut = `****** REQUEST START ******
GET /baz HTTP/1.1
Host: foo.bar
User-Agent: Go-http-client/1.1
Accept: text/json
Accept-Encoding: gzip


****** REQUEST END ******

****** RESPONSE START ******
HTTP/1.0 200 OK
Content-Type: text/plain
Content-Length: 0


****** RESPONSE END ******
`

func TestJSONDebugTransport_NoBody(t *testing.T) {
	out := bytes.NewBuffer(nil)
	transport := NewJSONDebugTransport(&testTransport{}, out)
	req, err := http.NewRequest("GET", "http://foo.bar/baz", nil)
	req.Header.Add("accept", "text/json")
	assert.NoError(t, err)
	_, err = transport.RoundTrip(req)
	assert.NoError(t, err)
	str := stripColors(normalizeLineBreak(out.String()))
	assert.Equal(t, testJSONDebugTransportNoJSONOut, str)
}

var testJSONDebugTransportJSONResponseOut = `****** REQUEST START ******
GET /baz HTTP/1.1
Host: foo.bar
User-Agent: Go-http-client/1.1
Accept: text/json
Accept-Encoding: gzip


****** REQUEST END ******

****** RESPONSE START ******
HTTP/1.0 200 OK
Content-Type: text/json
Content-Length: 0


{
  "message": "OK"
}
****** RESPONSE END ******
`

func TestJSONDebugTransport_JSONResponse(t *testing.T) {
	out := bytes.NewBuffer(nil)
	transport := NewJSONDebugTransport(&testTransport{body: []byte(`{"message":"OK"}`)}, out)
	req, err := http.NewRequest("GET", "http://foo.bar/baz", nil)
	req.Header.Add("accept", "text/json")
	assert.NoError(t, err)
	_, err = transport.RoundTrip(req)
	assert.NoError(t, err)
	str := stripColors(normalizeLineBreak(out.String()))
	assert.Equal(t, testJSONDebugTransportJSONResponseOut, str)
}

func normalizeLineBreak(from string) string {
	return strings.Join(strings.Split(from, "\r\n"), "\n")
}

var colorStrip = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripColors(from string) string {
	return colorStrip.ReplaceAllString(from, "")
}
