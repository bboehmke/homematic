package rpc

import (
	"context"
	"encoding/xml"
	"net"
	"net/http"
	"sync"
	"time"
)

// Handler for received RPC request
type Handler func(method string, params []interface{}) ([]interface{}, *Fault)

// Server for XML RPC events
type Server struct {
	srv      *http.Server
	listener net.Listener

	err     error
	running bool
	mutex   sync.RWMutex

	handler Handler
}

// NewServer creates new server
func NewServer(handler Handler) (*Server, error) {
	s := &Server{
		handler: handler,
	}
	var err error

	// listen on a random port
	s.listener, err = net.Listen("tcp4", "0.0.0.0:0")
	if err != nil {
		return nil, err
	}

	// create HTTP server
	s.srv = &http.Server{
		Handler: s,
	}

	return s, nil
}

// IsRunning returns true if server is running
func (s *Server) IsRunning() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.running
}

// Start server
func (s *Server) Start() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// already running?
	if s.running {
		return
	}

	go func() {
		if err := s.srv.Serve(s.listener); err != nil && err != http.ErrServerClosed {
			s.mutex.Lock()
			s.err = err
			s.mutex.Unlock()
		}
		s.mutex.Lock()
		s.running = false
		s.mutex.Unlock()
	}()
	s.running = true
}

// Stop server
func (s *Server) Stop() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// not running?
	if !s.running {
		return nil
	}

	// wait up to 2 seconds for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	return s.srv.Shutdown(ctx)
}

// Port of http server
func (s *Server) Port() int {
	addr := s.listener.Addr()
	return (addr.(*net.TCPAddr)).Port
}

// ServeHTTP handles received requests
func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	rpcRequest, err := ParseRequest(request.Body)
	if err != nil {
		return
	}

	response := new(Response)
	switch rpcRequest.Method {
	case "system.listMethods":
		response.Params = []interface{}{
			[]interface{}{
				"event",
			},
		}

	case "system.multicall":
		// parameters required
		if len(rpcRequest.Params) == 0 {
			response.Fault = &Fault{
				Code:   1,
				String: "parameter missing",
			}
			break
		}

		// first parameter contains list of calls
		functionCalls := rpcRequest.Params[0].([]interface{})
		response.Params = make([]interface{}, len(functionCalls))

		for idx, call := range functionCalls {
			// data must be a struct
			data, ok := call.(map[string]interface{})
			if !ok {
				response.Params[idx] = []interface{}{
					map[string]interface{}{
						"faultCode":   2,
						"faultString": "invalid function call",
					},
				}
				continue
			}

			// get method name and parameters
			methodName, ok := data["methodName"].(string)
			if !ok {
				response.Params[idx] = []interface{}{
					map[string]interface{}{
						"faultCode":   3,
						"faultString": "methodName missing",
					},
				}
				continue
			}
			params, ok := data["params"].([]interface{})
			if !ok {
				response.Params[idx] = []interface{}{
					map[string]interface{}{
						"faultCode":   3,
						"faultString": "params missing",
					},
				}
				continue
			}

			// call function
			var fault *Fault
			response.Params[idx], fault = s.handler(methodName, params)
			if fault != nil {
				response.Params[idx] = []interface{}{fault.toMap()}
			}
		}

	default:
		response.Params, response.Fault = s.handler(
			rpcRequest.Method, rpcRequest.Params)
	}

	_ = xml.NewEncoder(writer).Encode(&response)
}
