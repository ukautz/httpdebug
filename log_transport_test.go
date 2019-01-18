package httpdebug

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"net/http"
	"strings"
	"testing"
)

type noNextTransport struct{}

func (t *noNextTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return nil, nil
}

var testLogTransportOut = `****** REQUEST START ******
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
`

func TestLogTransport(t *testing.T) {
	out := bytes.NewBuffer(nil)
	in := bytes.NewBufferString(`{"foo":"bar"}`)
	transport := NewLogTransport(&noNextTransport{}, out)
	req, err := http.NewRequest("GET", "http://foo.bar/baz", in)
	req.Header.Add("accept", "text/json")
	assert.NoError(t, err)
	_, err = transport.RoundTrip(req)
	assert.NoError(t, err)
	str := strings.Replace(out.String(), "\r\n", "\n", -1)
	assert.Equal(t, testLogTransportOut, str)
}
