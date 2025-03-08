package types 

import (
	"database/sql/driver"
	"encoding"
	"github.com/golang/mock/gomock"
	entfield "entgo.io/ent/schema/field"
)


type TypeMashaler[T any] interface {
	encoding.TextMarshaler
	encoding.TextUnmarshaler
	entfield.ValueScanner
	entfield.Validator
	driver.Valuer
	gomock.Matcher
}