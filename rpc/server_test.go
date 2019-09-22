package rpc

import (
	"bytes"
	"encoding/xml"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestServer_StartStop(t *testing.T) {
	ass := assert.New(t)

	var handler Handler = func(_ string, _ []interface{}) ([]interface{}, *Fault) {
		return nil, nil
	}

	// create server
	server, err := NewServer(handler)
	ass.NoError(err)
	ass.NotNil(server.handler)

	// check port
	ass.Equal((server.listener.Addr().(*net.TCPAddr)).Port, server.Port())

	// multiple starts should work
	server.Start()
	server.Start()

	time.Sleep(time.Millisecond * 5)
	ass.True(server.IsRunning())

	// multiple starts should work
	ass.NoError(server.Stop())
	ass.NoError(server.Stop())

	time.Sleep(time.Millisecond * 5)
	ass.False(server.IsRunning())
}

func testRequest(request Request) (*http.Request, error) {
	buf := new(bytes.Buffer)
	err := xml.NewEncoder(buf).Encode(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("", "", buf)
	return req, err
}

func TestServer_ServeHTTP_listMethods(t *testing.T) {
	ass := assert.New(t)

	server := new(Server)
	request, err := testRequest(Request{
		Method: "system.listMethods",
	})
	ass.NoError(err)

	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, request)

	response, err := ParseResponse(recorder.Body)
	ass.NoError(err)
	ass.Equal(&Response{
		Params: []interface{}{
			[]interface{}{
				"event",
			},
		},
	}, response)
}

func TestServer_ServeHTTP_event(t *testing.T) {
	ass := assert.New(t)

	server := new(Server)

	server.handler = func(method string, params []interface{}) ([]interface{}, *Fault) {
		ass.Equal("event", method)
		ass.Equal([]interface{}{"aaa", "bbb"}, params)
		return []interface{}{"111", 222}, nil
	}

	request, err := testRequest(Request{
		Method: "event",
		Params: []interface{}{"aaa", "bbb"},
	})
	ass.NoError(err)

	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, request)

	response, err := ParseResponse(recorder.Body)
	ass.NoError(err)
	ass.Equal(&Response{
		Params: []interface{}{"111", int32(222)},
	}, response)
}

func TestServer_ServeHTTP_multicall(t *testing.T) {
	ass := assert.New(t)

	server := new(Server)

	server.handler = func(method string, params []interface{}) ([]interface{}, *Fault) {
		ass.Equal("event", method)
		ass.Equal([]interface{}{"aaa", "bbb"}, params)
		return []interface{}{"111", 222}, nil
	}

	request, err := testRequest(Request{
		Method: "system.multicall",
		Params: []interface{}{
			[]interface{}{
				map[string]interface{}{
					"methodName": "event",
					"params":     []interface{}{"aaa", "bbb"},
				},
				"aaa",
				map[string]interface{}{},
				map[string]interface{}{
					"methodName": "event",
				},
			},
		},
	})
	ass.NoError(err)

	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, request)

	response, err := ParseResponse(recorder.Body)
	ass.NoError(err)
	ass.Equal(&Response{
		Params: []interface{}{
			[]interface{}{
				"111", int32(222),
			},
			[]interface{}{
				map[string]interface{}{
					"faultCode":   int32(2),
					"faultString": "invalid function call",
				},
			},
			[]interface{}{
				map[string]interface{}{
					"faultCode":   int32(3),
					"faultString": "methodName missing",
				},
			},
			[]interface{}{
				map[string]interface{}{
					"faultCode":   int32(3),
					"faultString": "params missing",
				},
			},
		},
	}, response)
}
