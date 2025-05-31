package cursor_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dndev-xx/go-ninja-chat/internal/cursor"
	"github.com/dndev-xx/go-ninja-chat/internal/repositories/messages"
)

func TestSimpleEncode(t *testing.T) {
	c := messages.Cursor{
		LastCreatedAt: time.Unix(42, 42).UTC(),
		PageSize:      10,
	}
	rsl, err := cursor.Encode(c)
	require.NoError(t, err)
	assert.Equal(t, "eyJMYXN0Q3JlYXRlZEF0IjoiMTk3MC0wMS0wMVQwMDowMDo0Mi4wMDAwMDAwNDJaIiwiUGFnZVNpemUiOjEwfQ==", rsl)
}

func TestEncodeDecode(t *testing.T) {
	c1 := messages.Cursor{
		LastCreatedAt: time.Unix(42, 42).UTC(),
		PageSize:      10,
	}
	c, err := cursor.Encode(c1)
	require.NoError(t, err)
	assert.Equal(t, "eyJMYXN0Q3JlYXRlZEF0IjoiMTk3MC0wMS0wMVQwMDowMDo0Mi4wMDAwMDAwNDJaIiwiUGFnZVNpemUiOjEwfQ==", c)

	var c2 messages.Cursor
	require.NoError(t, cursor.Decode(c, &c2))
	assert.Equal(t, c1, c2)
}

func TestEncode_Errors(t *testing.T) {
	type StrangeCursor struct {
		Field chan struct{}
	}
	result, err := cursor.Encode(StrangeCursor{})
	require.Error(t, err)
	assert.Empty(t, result)
}

func TestDecode_Errors(t *testing.T) {
	t.Run("invalid base64", func(t *testing.T) {
		err := cursor.Decode(`{"page_size":50,"last":1670502502}`, new(messages.Cursor))
		assert.Error(t, err)
	})

	t.Run("invalid json", func(t *testing.T) {
		err := cursor.Decode("eyJwYWdlX3NpemUiOjUwLCJsYXN0IjoxNjcwNTAyNTAy", new(messages.Cursor)) // {"page_size":50,"last":1670502502
		assert.Error(t, err)
	})
}
