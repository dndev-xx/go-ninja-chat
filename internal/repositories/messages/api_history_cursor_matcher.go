package messages

import (
	"fmt"

	"github.com/golang/mock/gomock"
)

var _ gomock.Matcher = CursorMatcher{}

type CursorMatcher struct {
	c Cursor
}

func NewCursorMatcher(c Cursor) CursorMatcher {
	return CursorMatcher{c: c}
}

func (cm CursorMatcher) Matches(x any) bool {
	switch v := x.(type) {
	case Cursor:
		return v.PageSize == cm.c.PageSize && v.LastCreatedAt.Equal(cm.c.LastCreatedAt)
	case *Cursor:
		if v == nil {
			return false
		}
		return v.PageSize == cm.c.PageSize && v.LastCreatedAt.Equal(cm.c.LastCreatedAt)
	default:
		return false
	}
}

func (cm CursorMatcher) String() string {
	return fmt.Sprintf("{ps=%d, last_created_at=%v}", cm.c.PageSize, cm.c.LastCreatedAt)
}
