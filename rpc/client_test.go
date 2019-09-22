package rpc

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_Call(t *testing.T) {
	ass := assert.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		bytes, err := ioutil.ReadAll(req.Body)
		ass.NoError(err)
		defer req.Body.Close()

		ass.Equal(
			"<methodCall><methodName>test</methodName><params><param><value><string>42</string></value></param></params></methodCall>",
			string(bytes))

		_, err = rw.Write([]byte(`<methodResponse>
  <params>
    <param>
        <value><i4>42</i4></value>
    </param>
  </params>
</methodResponse>`))
		ass.NoError(err)
	}))
	defer server.Close()

	c := NewClient(server.URL)
	c.(*client).client = server.Client()

	response, err := c.Call("test", []interface{}{"42"})
	ass.NoError(err)
	ass.Equal(int32(42), response.FirstParam())

	ip, err := c.LocalIP()
	ass.NoError(err)
	ass.Equal("127.0.0.1", ip)
}
