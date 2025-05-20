package response

import (
	"errors"
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
)

type StatusCode int

var (
	OK           StatusCode = 200
	BAD_REQUEST  StatusCode = 400
	SERVER_ERROR StatusCode = 500
)

type writeState int

var (
	initState       writeState = 0
	statusLineState writeState = 1
	headersState    writeState = 2
	bodyState       writeState = 3
)

var ErrWrongWriteOrder = errors.New("wrong write order")

type Writer struct {
	w     io.Writer
	state writeState
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		w:     w,
		state: initState,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.state != initState {
		return ErrWrongWriteOrder
	}

	var err error = nil

	switch statusCode {
	case OK:
		_, err = w.w.Write([]byte("HTTP/1.1 200 OK\r\n"))
	case BAD_REQUEST:
		_, err = w.w.Write([]byte("HTTP/1.1 400 Bad Request\r\n"))
	case SERVER_ERROR:
		_, err = w.w.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n"))
	default:
		_, err = w.w.Write([]byte(fmt.Sprintf("HTTP/1.1 %v\r\n", statusCode)))
	}

	if err == nil {
		w.state = statusLineState
	}

	return err
}

func (w *Writer) WriteHeaders(h headers.Headers) error {
	if w.state != statusLineState {
		return ErrWrongWriteOrder
	}
	format := "%s: %v\r\n"
	for key, value := range h {
		_, err := w.w.Write([]byte(fmt.Sprintf(format, key, value)))
		if err != nil {
			return err
		}
	}

	_, err := w.w.Write([]byte("\r\n"))
	if err == nil {
		w.state = headersState
	}
	return err
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	if w.state != bodyState {
		return ErrWrongWriteOrder
	}

	format := "%s: %v\r\n"
	for key, value := range h {
		_, err := w.w.Write([]byte(fmt.Sprintf(format, key, value)))
		if err != nil {
			return err
		}
	}

	_, err := w.w.Write([]byte("\r\n"))
	return err
}

func (w *Writer) WriteBody(data []byte) (int, error) {
	if w.state != headersState {
		return 0, ErrWrongWriteOrder
	}

	w.state = bodyState

	return w.w.Write(data)
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.state != headersState {
		return 0, ErrWrongWriteOrder
	}

	bodyStr := fmt.Sprintf("%x\r\n%s\r\n", len(p), string(p))

	return w.w.Write([]byte(bodyStr))
}

func (w *Writer) WriteChunkedBodyDone(useTrailers bool) (int, error) {
	w.state = bodyState
	if useTrailers {
		return w.w.Write([]byte("0\r\n"))
	}
	return w.w.Write([]byte("0\r\n\r\n"))
}
