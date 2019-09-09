package script

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

		ass.Equal("testScript", string(bytes))

		_, err = rw.Write([]byte(`<xml><a>aaa</a><b>bbb</b></xml>`))
		ass.NoError(err)
	}))
	defer server.Close()

	c := NewClient(server.URL + "/")

	res, err := c.Call("testScript")
	ass.NoError(err)
	ass.Equal(Result{
		"a": "aaa",
		"b": "bbb",
	}, res)
}
