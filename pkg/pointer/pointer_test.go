package pointer_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dndev-xx/go-ninja-chat/pkg/pointer"
)

func TestIndirect(t *testing.T) {
	{
		i := 42
		s := "42"
		assert.Equal(t, 42, pointer.Indirect(&i))
		assert.Equal(t, "42", pointer.Indirect(&s))
	}

	{
		assert.Equal(t, 0, pointer.Indirect((*int)(nil)))
		assert.Equal(t, "", pointer.Indirect((*string)(nil)))
	}

	{
		var iPtr *int
		var sPtr *string
		assert.Equal(t, 0, pointer.Indirect(iPtr))
		assert.Equal(t, "", pointer.Indirect(sPtr))
	}

	{
		iPtr := new(int)
		sPtr := new(string)
		assert.Equal(t, 0, pointer.Indirect(iPtr))
		assert.Equal(t, "", pointer.Indirect(sPtr))
	}
}

func TestPtr(t *testing.T) {
	assert.NotNil(t, pointer.Ptr(""))
	assert.NotNil(t, pointer.Ptr(0))
	assert.NotNil(t, pointer.Ptr([3]int{}))
	assert.NotNil(t, pointer.Ptr([]int(nil)))

	assert.NotNil(t, pointer.Ptr("42"))
	assert.NotNil(t, pointer.Ptr(42))
	assert.NotNil(t, pointer.Ptr(new(struct{ string })))

	var local string
	assert.Equal(t, &local, pointer.Ptr(local))
}

func TestPtrWithZeroAsNil(t *testing.T) {
	assert.Nil(t, pointer.PtrWithZeroAsNil(""))
	assert.Nil(t, pointer.PtrWithZeroAsNil(0))
	assert.Nil(t, pointer.PtrWithZeroAsNil([3]int{}))
	assert.Nil(t, pointer.PtrWithZeroAsNil((*struct{ string })(nil)))

	assert.NotNil(t, pointer.PtrWithZeroAsNil("42"))
	assert.NotNil(t, pointer.PtrWithZeroAsNil(42))
	assert.NotNil(t, pointer.PtrWithZeroAsNil([3]int{1}))
	assert.NotNil(t, pointer.PtrWithZeroAsNil(new(struct{ string })))

	local := "42"
	assert.Equal(t, &local, pointer.PtrWithZeroAsNil(local))
}
