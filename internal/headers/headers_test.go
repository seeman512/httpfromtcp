package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeadersParse(t *testing.T) {
	tests := []struct {
		input    string
		n        int
		done     bool
		err      error
		notes    string
		key      string
		value    string
		checkKey bool
	}{
		{
			input:    "test",
			n:        0,
			done:     false,
			err:      nil,
			checkKey: false,
			notes:    "No new line",
		},
		{
			input:    "Host: localhost:42069\r\n\r\n",
			n:        23,
			done:     false,
			err:      nil,
			checkKey: true,
			key:      "host",
			value:    "localhost:42069",
			notes:    "Valid header",
		},
		{
			input:    "Ho@st: localhost:42069\r\n\r\n",
			n:        0,
			done:     false,
			err:      ErrWrongKeyFormat,
			checkKey: false,
			notes:    "Wrong key symbols",
		},
		{
			input:    "       Host : localhost:42069       \r\n\r\n",
			n:        0,
			done:     false,
			err:      ErrWrongFormat,
			checkKey: false,
			notes:    "Invalid spacing header",
		},
		{
			input:    "\r\n Body Message\r\n",
			n:        2,
			done:     true,
			err:      nil,
			checkKey: false,
			notes:    "Valid done",
		},
		{
			input:    "       Host  localhost 42069       \r\n\r\n",
			n:        0,
			done:     false,
			err:      ErrWrongFormat,
			checkKey: false,
			notes:    "Invalid header format",
		},
	}

	for _, tt := range tests {
		headers := NewHeaders()
		data := []byte(tt.input)
		n, done, err := headers.Parse(data)
		assert.Equal(t, n, tt.n, tt.notes+": n")
		assert.Equal(t, done, tt.done, tt.notes+": done")
		assert.Equal(t, err, tt.err, tt.notes+": error")

		if tt.checkKey {
			assert.Equal(t, headers[tt.key], tt.value, tt.notes+": key/value")
		}
	}
}

func TestHeadersParseExistingKey(t *testing.T) {
	data1 := "Set-Person: person1\r\n"
	data2 := "Set-Person: person2\r\n"

	headers := NewHeaders()

	n, done, err := headers.Parse([]byte(data1))
	assert.Equal(t, n, 21)
	assert.Equal(t, done, false)
	assert.Equal(t, err, nil)
	assert.Equal(t, headers["set-person"], "person1")

	n, done, err = headers.Parse([]byte(data2))
	assert.Equal(t, n, 21)
	assert.Equal(t, done, false)
	assert.Equal(t, err, nil)
	assert.Equal(t, headers["set-person"], "person1,person2")
}
