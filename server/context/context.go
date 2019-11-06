package context

import (
	gocontext "context"
)

type ErrorChannelKey struct{}

func GetErrorChannel(ctx gocontext.Context) chan<- error {
	errC := ctx.Value(ErrorChannelKey{})
	if errC == nil {
		panic("nil error channel")
	}
	return errC.(chan<- error)
}
