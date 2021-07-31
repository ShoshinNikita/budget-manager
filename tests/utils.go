package tests

import (
	"context"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func getFreePort(t *testing.T) (port int) {
	require := require.New(t)

	listener, err := net.Listen("tcp", "")
	require.NoError(err)
	defer listener.Close()

	tcpAddr, ok := listener.Addr().(*net.TCPAddr)
	require.True(ok)

	return tcpAddr.Port
}

func newRequest(t *testing.T, method Method, url string, body io.Reader) (req *http.Request, cancelCtx func()) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	req, err := http.NewRequestWithContext(ctx, string(method), url, body)
	require.NoError(t, err)

	return req, cancel
}

func ptrStr(v string) *string {
	return &v
}

func ptrUint(v uint) *uint {
	return &v
}

func ptrFloat(v float64) *float64 {
	return &v
}
