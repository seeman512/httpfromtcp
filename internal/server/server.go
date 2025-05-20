package server

import (
	"fmt"
	"httpfromtcp/internal/headers"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"io"
	"net"
	"strconv"
	"sync/atomic"
	"time"
)

func writeError(w io.Writer, err error) error {
	rw := response.NewWriter(w)
	writeErr := rw.WriteStatusLine(response.SERVER_ERROR)

	if writeErr != nil {
		return writeErr
	}

	errData := []byte(err.Error())
	h := headers.NewHeaders()
	h.SetDefault(len(errData), nil)
	writeErr = rw.WriteHeaders(h)
	if writeErr != nil {
		return writeErr
	}
	_, writeErr = rw.WriteBody(errData)
	return writeErr

}

type Handler func(w io.Writer, req *request.Request)

type Server struct {
	isOpen   atomic.Bool
	listener net.Listener
	Handler  Handler
}

func Serv(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(port))
	if err != nil {
		return nil, err
	}

	server := &Server{
		listener: listener,
		isOpen:   atomic.Bool{},
		Handler:  handler,
	}

	server.isOpen.Store(true)

	go server.listen()

	return server, nil
}

func (s *Server) Close() error {
	fmt.Println("Server closed")
	s.isOpen.Store(false)
	return s.listener.Close()
}

func (s *Server) listen() {
	for s.isOpen.Load() {

		conn, err := s.listener.Accept()
		if err != nil {
			fmt.Printf("Accept error: %v\n", err)
			return
		}

		fmt.Println("Connection Accepted")
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer func() {
		fmt.Println("Connection closed")
		conn.Close()
	}()

	timeoutDuration := 1 * time.Second
	conn.SetReadDeadline(time.Now().Add(timeoutDuration))

	r, err := request.RequestFromReader(conn)

	if err != nil {
		err = writeError(conn, err)
		if err != nil {
			fmt.Printf("Request error: %v\n", err)
		}
		return
	}

	s.Handler(conn, r)
}
