package network

import (
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/common/types"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/wire"
)

type Handler func(method *wire.Method, arguments []types.Value) (types.Value, error)

var errServerClosed = errors.New("the server was closed")

type Server struct {
	listener         net.Listener
	serviceInterface *wire.Interface
	calls            chan call
	done             chan struct{}
	mutex            sync.Mutex
	closed           bool
	connections      map[net.Conn]bool
	workers          sync.WaitGroup
}

type call struct {
	method    *wire.Method
	arguments []types.Value
	replies   chan reply
}

type reply struct {
	result types.Value
	err    error
}

func Listen(address common.Address, serviceInterface *wire.Interface) (*Server, error) {
	listener, err := net.Listen("tcp", address.String())
	if err != nil {
		return nil, rootCause(err)
	}
	return &Server{
		listener:         listener,
		serviceInterface: serviceInterface,
		calls:            make(chan call),
		done:             make(chan struct{}),
		connections:      map[net.Conn]bool{},
	}, nil
}

func (s *Server) Addr() net.Addr {
	return s.listener.Addr()
}

func (s *Server) Serve(handler Handler) error {
	accepted := make(chan error, 1)
	go func() { accepted <- s.accept() }()
	for {
		select {
		case call := <-s.calls:
			if s.isClosed() {
				call.replies <- reply{err: errServerClosed}
				continue
			}
			result, err := handler(call.method, call.arguments)
			call.replies <- reply{result: result, err: err}
		case err := <-accepted:
			closedByUser := s.isClosed()
			s.Close()
			s.workers.Wait()
			if closedByUser {
				return nil
			}
			return err
		}
	}
}

func (s *Server) Close() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	close(s.done)
	for conn := range s.connections {
		conn.Close()
	}
	return s.listener.Close()
}

func (s *Server) isClosed() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.closed
}

func (s *Server) accept() error {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return err
		}
		if s.track(conn) {
			go s.serveConnection(conn)
		}
	}
}

func (s *Server) track(conn net.Conn) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.closed {
		conn.Close()
		return false
	}
	s.connections[conn] = true
	s.workers.Add(1)
	return true
}

func (s *Server) untrack(conn net.Conn) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.connections, conn)
	conn.Close()
}

func (s *Server) serveConnection(conn net.Conn) {
	defer s.workers.Done()
	defer s.untrack(conn)
	if !s.handshake(conn) {
		return
	}
	for {
		payload, err := readFrame(conn)
		if err != nil {
			return
		}
		method, arguments, err := decodeRequest(payload, s.serviceInterface)
		if err != nil {
			conn.Write(encodeFailure(statusProtocolError, err.Error()))
			return
		}
		response, ok := s.execute(method, arguments)
		if !ok {
			return
		}
		if _, err := conn.Write(response); err != nil {
			return
		}
	}
}

func (s *Server) handshake(conn net.Conn) bool {
	hash, err := readHello(conn)
	switch {
	case errors.Is(err, errUnsupportedVersion):
		conn.Write(encodeRejection(err.Error()))
		return false
	case err != nil:
		return false
	case hash != s.serviceInterface.Hash:
		conn.Write(encodeRejection(fmt.Sprintf("interface mismatch with the server, which implements service '%s'", s.serviceInterface.Name)))
		return false
	}
	_, err = conn.Write(encodeWelcome())
	return err == nil
}

func (s *Server) execute(method *wire.Method, arguments []types.Value) ([]byte, bool) {
	replies := make(chan reply, 1)
	select {
	case s.calls <- call{method: method, arguments: arguments, replies: replies}:
	case <-s.done:
		return nil, false
	}
	executed := <-replies
	if executed.err != nil {
		return encodeFailure(statusServerError, executed.err.Error()), true
	}
	response, err := encodeResult(method, executed.result)
	if err != nil {
		return encodeFailure(statusServerError, fmt.Sprintf("cannot send the result of '%s': %v", method.Name, err)), true
	}
	return response, true
}
