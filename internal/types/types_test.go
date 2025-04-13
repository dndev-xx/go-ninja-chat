package types_test

import (
	"database/sql/driver"
	"encoding"
	"testing"

	entfield "entgo.io/ent/schema/field"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dndev-xx/go-ninja-chat/internal/types"
)

var _ interface {
	encoding.TextMarshaler
	encoding.TextUnmarshaler
	entfield.ValueScanner
	entfield.Validator
	driver.Valuer
	gomock.Matcher
} = (*types.ChatID)(nil)

func TestParseChatID(t *testing.T) {
	t.Run("valid uuid string", func(t *testing.T) {
		chatID, err := types.ParseChatID("f0317e88-bbfe-11ed-8728-461e464ebed8")
		require.NoError(t, err)
		assert.Equal(t, "f0317e88-bbfe-11ed-8728-461e464ebed8", chatID.String())
	})

	t.Run("valid uuid object", func(t *testing.T) {
		uuidVal := uuid.New()
		chatID, err := types.ParseChatID(uuidVal)
		require.NoError(t, err)
		assert.Equal(t, uuidVal.String(), chatID.String())
	})

	t.Run("valid bytes", func(t *testing.T) {
		uuidVal := uuid.New()
		chatID, err := types.ParseChatID([]byte(uuidVal.String()))
		require.NoError(t, err)
		assert.Equal(t, uuidVal.String(), chatID.String())
	})

	t.Run("invalid string", func(t *testing.T) {
		_, err := types.ParseChatID("invalid-uuid")
		require.NoError(t, err) // теперь сохраняет raw string
	})

	t.Run("empty string", func(t *testing.T) {
		chatID, err := types.ParseChatID("")
		require.NoError(t, err)
		assert.Equal(t, types.ChatIDNil.String(), chatID.String())
	})

	t.Run("nil value", func(t *testing.T) {
		chatID, err := types.ParseChatID(nil)
		require.NoError(t, err)
		assert.Equal(t, types.ChatIDNil.String(), chatID.String())
	})

	t.Run("integer value", func(t *testing.T) {
		_, err := types.ParseChatID(123)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported type for ChatID")
	})
}

func TestMustParseChatID(t *testing.T) {
	t.Run("valid uuid", func(t *testing.T) {
		assert.NotPanics(t, func() {
			chatID := types.MustParseChatID("f0317e88-bbfe-11ed-8728-461e464ebed8")
			assert.Equal(t, "f0317e88-bbfe-11ed-8728-461e464ebed8", chatID.String())
		})
	})

	t.Run("invalid type", func(t *testing.T) {
		assert.Panics(t, func() {
			types.MustParseChatID(123)
		})
	})
}

func TestChatIDNil(t *testing.T) {
	assert.Equal(t, types.ChatIDNil.String(), uuid.Nil.String())
}

func TestChatIDString(t *testing.T) {
	id := types.NewChatID()
	require.NotEmpty(t, id.String())
	assert.Equal(t, uuid.MustParse(id.String()).String(), id.String())
}

func TestChatIDScan(t *testing.T) {
	const src = "5c9de646-529c-11ed-81ba-461e464ebed9"

	t.Run("from string and bytes", func(t *testing.T) {
		var id1, id2 types.ChatID
		{
			err := id1.Scan(src)
			require.NoError(t, err)
		}
		{
			err := id2.Scan([]byte(src))
			require.NoError(t, err)
		}
		assert.Equal(t, id1.String(), id2.String())
		assert.Equal(t, getValueAsString(t, &id1), getValueAsString(t, &id2))
	})

	t.Run("from NULL", func(t *testing.T) {
		for _, src := range []any{nil, []byte(nil), []byte{}, ""} {
			t.Run("", func(t *testing.T) {
				var id types.ChatID
				err := id.Scan(src)
				require.NoError(t, err)
				assert.Equal(t, types.ChatIDNil.String(), id.String())
				assert.Equal(t, types.ChatIDNil.String(), getValueAsString(t, &id))
			})
		}
	})
}

func TestChatIDIsZero(t *testing.T) {
	assert.True(t, types.ChatIDNil.IsZero())
	
	id := types.NewChatID()
	assert.True(t, id.IsZero())

	parsedID := types.MustParseChatID("f0317e88-bbfe-11ed-8728-461e464ebed8")
	assert.False(t, parsedID.IsZero())
}

func TestChatIDValidate(t *testing.T) {
	t.Run("valid uuid", func(t *testing.T) {
		id := types.MustParseChatID("f0317e88-bbfe-11ed-8728-461e464ebed8")
		assert.NoError(t, id.Validate())
	})

	t.Run("nil value", func(t *testing.T) {
		assert.Error(t, types.ChatIDNil.Validate())
	})

	t.Run("invalid string format", func(t *testing.T) {
		// Используем ParseChatID вместо прямого доступа к полю value
		id, err := types.ParseChatID("invalid-uuid")
		require.NoError(t, err)
		assert.Error(t, id.Validate())
	})

	t.Run("empty value", func(t *testing.T) {
		id := types.NewChatID()
		assert.Error(t, id.Validate())
	})
}

func TestChatIDMatches(t *testing.T) {
	validUUID := "f0317e88-bbfe-11ed-8728-461e464ebed8"
	
	t.Run("matches string", func(t *testing.T) {
		id := types.MustParseChatID(validUUID)
		assert.True(t, id.Matches(validUUID))
	})

	t.Run("matches uuid.UUID", func(t *testing.T) {
		id := types.MustParseChatID(validUUID)
		uuidVal := uuid.MustParse(validUUID)
		assert.True(t, id.Matches(uuidVal))
	})

	t.Run("matches bytes", func(t *testing.T) {
		id := types.MustParseChatID(validUUID)
		assert.True(t, id.Matches([]byte(validUUID)))
	})

	t.Run("does not match different value", func(t *testing.T) {
		id := types.MustParseChatID(validUUID)
		assert.False(t, id.Matches("different-value"))
	})

	t.Run("nil receiver", func(t *testing.T) {
		var id *types.ChatID
		assert.False(t, id.Matches(validUUID))
	})
}

func TestChatIDMarshalText(t *testing.T) {
	validUUID := "f0317e88-bbfe-11ed-8728-461e464ebed8"
	
	t.Run("marshal uuid value", func(t *testing.T) {
		id := types.MustParseChatID(validUUID)
		v, err := id.MarshalText()
		require.NoError(t, err)
		assert.Equal(t, validUUID, string(v))
	})

	t.Run("marshal string value", func(t *testing.T) {
		id, err := types.ParseChatID(validUUID)
		require.NoError(t, err)
		v, err := id.MarshalText()
		require.NoError(t, err)
		assert.Equal(t, validUUID, string(v))
	})

	t.Run("marshal nil value", func(t *testing.T) {
		id := types.ChatIDNil
		v, err := id.MarshalText()
		require.NoError(t, err)
		assert.Equal(t, uuid.Nil.String(), string(v))
	})
}

func TestChatIDValue(t *testing.T) {
	t.Run("valid uuid", func(t *testing.T) {
		id := types.MustParseChatID("f0317e88-bbfe-11ed-8728-461e464ebed8")
		val, err := id.Value()
		require.NoError(t, err)
		assert.Equal(t, "f0317e88-bbfe-11ed-8728-461e464ebed8", val)
	})

	t.Run("nil value", func(t *testing.T) {
		val, err := types.ChatIDNil.Value()
		require.NoError(t, err)
		assert.Equal(t, uuid.Nil.String(), val)
	})
}

func getValueAsString(t *testing.T, valuer driver.Valuer) string {
	t.Helper()
	v, err := valuer.Value()
	require.NoError(t, err)
	vv, ok := v.(string)
	require.True(t, ok)
	return vv
}