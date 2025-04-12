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

func TestParse(t *testing.T) {
	_, err := types.Parse[types.ChatID]("abra-cadabra")
	require.NoError(t, err)
	chatID, err := types.Parse[types.ChatID]("f0317e88-bbfe-11ed-8728-461e464ebed8")
	require.NoError(t, err)
	chatIDInt, err := types.Parse[types.ChatID](1)
	require.NoError(t, err)

	assert.Equal(t, "f0317e88-bbfe-11ed-8728-461e464ebed8", chatID.String())
	assert.Equal(t, "1", chatIDInt.String())
}

func TestMustParse(t *testing.T) {
	assert.Panics(t, func() {
		types.MustParse[types.ChatID]("")
	})

	assert.NotPanics(t, func() {
		chatID := types.MustParse[types.ChatID]("f0317e88-bbfe-11ed-8728-461e464ebed8")
		assert.Equal(t, "f0317e88-bbfe-11ed-8728-461e464ebed8", chatID.String())
	})
}

func TestChatIDNil(t *testing.T) {
	t.Log(types.ChatIDNil)
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

func TestChatIDMarshalText(t *testing.T) {
	chatID := types.MustParse[types.ChatID]("f0317e88-bbfe-11ed-8728-461e464ebed8")
	v, err := chatID.MarshalText()
	require.NoError(t, err)
	assert.Equal(t, "f0317e88-bbfe-11ed-8728-461e464ebed8", string(v))

	var chatID2 types.ChatID
	err = chatID2.UnmarshalText(v)
	require.NoError(t, err)
	assert.Equal(t, chatID.String(), chatID2.String())
}

func TestChatIDIsZero(t *testing.T) {
	id := types.NewChatID()
	assert.True(t, id.IsZero())
	assert.True(t, types.ChatIDNil.IsZero())
	assert.Equal(t, uuid.Nil.String(), types.ChatIDNil.String())
}

func TestChatIDMatches(t *testing.T) {
	id := types.NewChatID()
	id2 := types.MustParse[types.ChatID](id.String())
	assert.NotEqual(t, id, id2.String())
	assert.False(t, id.Matches(id2.String()))
}

//nolint:testifylint // not directly related single checks
func TestChatIDValidate(t *testing.T) {
	tt := types.ChatID{}
	assert.Error(t, types.NewChatID().Validate())
	assert.Error(t, tt.Validate())
	assert.Error(t, types.ChatIDNil.Validate())
}

func getValueAsString(t *testing.T, valuer driver.Valuer) string {
	t.Helper()
	v, err := valuer.Value()
	require.NoError(t, err)
	vv, ok := v.(string)
	require.True(t, ok)
	return vv
}
