package headers

import (
	"bytes"
	"errors"
	"maps"
	"regexp"
	"strconv"
	"strings"
)

const lineSeparator = "\r\n"
const valuesSeparator = ","

var keyReg = regexp.MustCompile("^[a-zA-Z0-9!#$%&'*+-.^_`|]+$")

var (
	ErrWrongFormat    = errors.New("wrong format")
	ErrWrongKeyFormat = errors.New("wrong key format")
)

type Headers map[string]string

func NewHeaders() Headers {
	return map[string]string{}
}

func (h Headers) Get(key string) (string, bool) {
	val, ok := h[key]
	return val, ok
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	nlIdx := bytes.Index(data, []byte(lineSeparator))

	if nlIdx == -1 {
		return 0, false, nil
	}

	if nlIdx == 0 {
		return 2, true, nil
	}

	done = false

	colonIdx := bytes.Index(data[:nlIdx], []byte(":"))
	if colonIdx == -1 {
		return 0, done, ErrWrongFormat
	}

	left := string(data[0:colonIdx])
	right := string(data[colonIdx+1 : nlIdx])

	if strings.HasSuffix(left, " ") {
		return 0, done, ErrWrongFormat
	}

	left = strings.Trim(left, " ")
	right = strings.Trim(right, " ")

	if !keyReg.MatchString(left) {
		return 0, done, ErrWrongKeyFormat
	}

	key := strings.ToLower(left)

	oldVal, ok := h[key]
	if ok {
		h[key] = oldVal + valuesSeparator + right
	} else {
		h[key] = right
	}
	return nlIdx + 2, done, nil
}

func (h Headers) SetDefault(contentLen int, customHeaders map[string]string) {

	h["content-length"] = strconv.Itoa(contentLen)
	h["connection"] = "close"
	h["content-type"] = "text/plain"

	if customHeaders == nil {
		return
	}

	maps.Copy(h, customHeaders)
	// for k, v := range customHeaders {
	// 	h[k] = v
	// }
}

func (h Headers) Set(customHeaders map[string]string) {
	maps.Copy(h, customHeaders)
}
