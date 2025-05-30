package errhandler_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dndev-xx/go-ninja-chat/internal/server-client/errhandler"
	clientv1 "github.com/dndev-xx/go-ninja-chat/internal/server-client/v1/pkg"
)

func TestResponseBuilder(t *testing.T) {
	t.Run("with details", func(t *testing.T) {
		err := errhandler.ResponseBuilder(1000, "hello", "world")

		resp, ok := err.(errhandler.Response)
		require.True(t, ok)
		require.IsType(t, clientv1.Error{}, resp.Error)

		assert.Equal(t, clientv1.ErrorCode(1000), resp.Error.Code)
		assert.Equal(t, "hello", resp.Error.Message)
		require.NotNil(t, resp.Error.Details)
		assert.Equal(t, "world", *resp.Error.Details)
	})

	t.Run("without details", func(t *testing.T) {
		err := errhandler.ResponseBuilder(1001, "hello", "")

		resp, ok := err.(errhandler.Response)
		require.True(t, ok)
		require.IsType(t, clientv1.Error{}, resp.Error)

		assert.Equal(t, clientv1.ErrorCode(1001), resp.Error.Code)
		assert.Equal(t, "hello", resp.Error.Message)
		assert.Nil(t, resp.Error.Details)
	})
}
