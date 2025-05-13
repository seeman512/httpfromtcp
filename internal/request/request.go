package request

import (
	"bytes"
	"errors"
	"io"
	"regexp"
	"strings"
)

const lineSeparator = "\r\n"
const partsSeparator = " "
const bufferSize = 8

type ParseState int

var Initialized ParseState = 0
var Done ParseState = 1

var (
	ErrEmptyBody          = errors.New("empty body")
	ErrWrongPartsCnt      = errors.New("wrong parts; should be 3")
	ErrWrongVersionFormat = errors.New("wrong version format")
	ErrWrongTargetFormat  = errors.New("wrong target format")
	ErrWrongMethodFormat  = errors.New("wrong method format")
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	state       ParseState
	RequestLine RequestLine
}

func parseRequestLine(data []byte) (int, *RequestLine, error) {
	idx := bytes.Index(data, []byte(lineSeparator))

	if idx == -1 {
		return 0, nil, nil
	}

	requestLine := string(data[:idx])
	if requestLine == "" {
		return 0, nil, ErrEmptyBody
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

func (r *Request) parse(data []byte) (int, error) {
	n, requestLine, err := parseRequestLine(data)
	if err != nil {
		return 0, err
	}

	if n == 0 {
		return n, nil
	}

	r.RequestLine = *requestLine
	r.state = Done
	return n, nil
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

	r := Request{state: Initialized}
	i := 1

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
				r.state = Done
			} else {
				return nil, err
			}
		}

		readToIndex += n

		n, err = r.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}

		if n == 0 {
			continue
		}

		l := len(buf[n:])
		tmpBuff := make([]byte, l)
		copy(tmpBuff, buf)
		buf = tmpBuff
		readToIndex -= n
	}

	return &r, nil
}
