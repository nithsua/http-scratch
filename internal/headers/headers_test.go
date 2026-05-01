package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeadersParse(t *testing.T) {
	// Test: Single Header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Single Header with extra whitespace
	headers = NewHeaders()
	data = []byte("Host:        localhost:42069        \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 38, n)
	assert.False(t, done)

	// Test: 2 Headers with existing headers
	headers = NewHeaders()
	data = []byte("Host: localhost:42069\r\nContent-Type: application/json\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)
	n, done, err = headers.Parse(data[n:])
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "application/json", headers["content-type"])
	assert.Equal(t, 32, n)
	assert.False(t, done)

	// Test: Invalid space between key and colon
	// headers = NewHeaders()
	data = []byte("Host : localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Valid done
	headers = NewHeaders()
	data = []byte("\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, 2, n)
	assert.True(t, done)

	// Test: Mixed-case header keys are normalized to lowercase
	headers = NewHeaders()
	data = []byte("HOST: localhost:42069\r\nContent-TYPE: application/json\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.False(t, done)
	n, done, err = headers.Parse(data[n:])
	require.NoError(t, err)
	assert.Equal(t, "application/json", headers["content-type"])
	assert.False(t, done)

	// Test: Invalid character in header key
	headers = NewHeaders()
	data = []byte("H©st: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Pre-existing header key matches incoming header (values are appended)
	headers = NewHeaders()
	headers["set-person"] = "lane-loves-go"
	data = []byte("Set-Person: prime-loves-zig\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "lane-loves-go, prime-loves-zig", headers["set-person"])
	assert.False(t, done)
}
