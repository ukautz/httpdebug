package httpdebug

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
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

var testDebugTransportJSONOut = `****** REQUEST START ******
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

func TestDebugTransport(t *testing.T) {
	out := bytes.NewBuffer(nil)
	in := bytes.NewBufferString(`{"foo":"bar"}`)
	transport := NewDebugTransport(&testTransport{}, out)
	req, err := http.NewRequest("GET", "http://foo.bar/baz", in)
	req.Header.Add("accept", "text/json")
	assert.NoError(t, err)
	_, err = transport.RoundTrip(req)
	assert.NoError(t, err)
	str := strings.Replace(out.String(), "\r\n", "\n", -1)
	assert.Equal(t, testDebugTransportJSONOut, str)
}

var testDebugTransportNoJSONOut = `****** REQUEST START ******
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

func TestDebugTransport_NoBody(t *testing.T) {
	out := bytes.NewBuffer(nil)
	transport := NewDebugTransport(&testTransport{}, out)
	req, err := http.NewRequest("GET", "http://foo.bar/baz", nil)
	req.Header.Add("accept", "text/json")
	assert.NoError(t, err)
	_, err = transport.RoundTrip(req)
	assert.NoError(t, err)
	str := strings.Replace(out.String(), "\r\n", "\n", -1)
	assert.Equal(t, testDebugTransportNoJSONOut, str)
}

var testDebugTransportJSONResponseOut = `****** REQUEST START ******
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

func TestDebugTransport_JSONResponse(t *testing.T) {
	out := bytes.NewBuffer(nil)
	in := bytes.NewBufferString(`{"foo":"bar"}`)
	transport := NewDebugTransport(&testTransport{body: []byte(`{"message":"OK"}`)}, out)
	req, err := http.NewRequest("GET", "http://foo.bar/baz", in)
	req.Header.Add("accept", "text/json")
	assert.NoError(t, err)
	_, err = transport.RoundTrip(req)
	assert.NoError(t, err)
	str := strings.Replace(out.String(), "\r\n", "\n", -1)
	assert.Equal(t, testDebugTransportJSONResponseOut, str)
}
