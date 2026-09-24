package service

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestForceIPv4DialContext(t *testing.T) {
	expectedErr := errors.New("dial stopped")
	var actualNetwork string
	var actualAddress string
	dial := forceIPv4DialContext(func(
		_ context.Context, network string, address string,
	) (net.Conn, error) {
		actualNetwork = network
		actualAddress = address
		return nil, expectedErr
	})

	_, err := dial(context.Background(), "tcp6", "api.dingtalk.com:443")

	require.ErrorIs(t, err, expectedErr)
	require.Equal(t, "tcp4", actualNetwork)
	require.Equal(t, "api.dingtalk.com:443", actualAddress)
}
