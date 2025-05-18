package request

import (
	"bytes"
	"errors"
	"httpfromtcp/internal/headers"
	"io"
	"regexp"
	"strconv"
	"strings"
)

const lineSeparator = "\r\n"
const partsSeparator = " "
const bufferSize = 8

type ParseState int

var Initialized ParseState = 0
var ParsedRequestLine ParseState = 1
var ParsedHeaders ParseState = 2
var Done ParseState = 3

var (
	ErrEmptyRequestLine   = errors.New("empty request line")
	ErrWrongPartsCnt      = errors.New("wrong parts; should be 3")
	ErrWrongVersionFormat = errors.New("wrong version format")
	ErrWrongTargetFormat  = errors.New("wrong target format")
	ErrWrongMethodFormat  = errors.New("wrong method format")
	ErrWrongBodyLength    = errors.New("wrong body length")
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	state       ParseState
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
}

func parseRequestLine(data []byte) (int, *RequestLine, error) {
	idx := bytes.Index(data, []byte(lineSeparator))

	if idx == -1 {
		return 0, nil, nil
	}

	requestLine := string(data[:idx])
	if requestLine == "" {
		return 0, nil, ErrEmptyRequestLine
	}

	parts := strings.Split(requestLine, partsSeparator)
	if len(parts) != 3 {
		return 0, nil, ErrWrongPartsCnt
	}

	method, err := parseMethod(parts[0])
	if err != nil {
		return 0, nil, err
	}

	version, err := parseVersion(parts[2])
	if err != nil {
		return 0, nil, err
	}

	return idx + 2, &RequestLine{
		HttpVersion:   version,
		RequestTarget: parts[1],
		Method:        method,
	}, nil
}

func (r *Request) parse(data []byte, eof bool) (int, error) {
	if r.state == Done {
		return 0, nil
	}

	// parse request line
	if r.state == Initialized {
		n, requestLine, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}

		if n == 0 {
			return n, nil
		}

		r.RequestLine = *requestLine
		r.state = ParsedRequestLine
		return n, nil
	}

	// parse headers
	if r.state == ParsedRequestLine {
		n, done, err := r.Headers.Parse(data)

		if err != nil {
			return n, err
		}

		if done {
			r.state = ParsedHeaders
		}

		return n, nil
	}

	// parse body
	if r.state == ParsedHeaders {
		val, ok := r.Headers.Get("content-length")
		if !ok {
			r.state = Done
			return 0, nil
		}

		length, err := strconv.Atoi(val)
		if err != nil {
			return 0, err
		}

		l := len(data)
		if eof {
			if length != l {
				return 0, ErrWrongBodyLength
			}
			r.state = Done
			r.Body = data[:]
			return l, nil
		} else {
			return 0, nil
		}
	}

	return 0, errors.New("uknown parse error")
}

func parseMethod(method string) (string, error) {
	reg := regexp.MustCompile("^[A-Z]+$")
	if reg.MatchString(method) {
		return method, nil
	}
	return "", ErrWrongMethodFormat
}

func parseVersion(version string) (string, error) {
	if version != "HTTP/1.1" {
		return "", ErrWrongVersionFormat
	}

	return "1.1", nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)
	readToIndex := 0

	r := Request{
		state:   Initialized,
		Headers: headers.NewHeaders(),
	}
	i := 1

	eof := false

	for r.state != Done {
		// buffer is full, twice buffer size and copy
		if readToIndex >= len(buf)-1 {
			i++
			tmpBuff := make([]byte, bufferSize*i)
			copy(tmpBuff, buf)
			buf = tmpBuff
		}

		n, err := reader.Read(buf[readToIndex:])

		if err != nil {
			if errors.Is(err, io.EOF) {
				// r.state = Done
				eof = true
			} else {
				return nil, err
			}
		}

		readToIndex += n

		n, err = r.parse(buf[:readToIndex], eof)
		if err != nil {
			return nil, err
		}

		if n == 0 {
			continue
		}

		l := len(buf[n:])
		tmpBuff := make([]byte, l)
		copy(tmpBuff, buf[n:])
		buf = tmpBuff
		readToIndex -= n
	}

	return &r, nil
}
