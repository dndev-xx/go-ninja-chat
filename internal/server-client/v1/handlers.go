package v1

import (
)

//go:generate options-gen -out-filename=handlers_options.gen.go -from-struct=Options
type Options struct {
}

type Handlers struct {
	Options
}

func NewHandlers(opts Options) (*Handlers, error) {	
	return &Handlers{
		Options: opts,
	}, nil
}