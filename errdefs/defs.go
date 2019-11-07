package errdefs

import "fmt"

type errWrongFormat struct {
	s string
}

func (e errWrongFormat) Error() string {
	return fmt.Sprintf("%v has wrong format", e.s)
}

func NewErrWrongFormat(s string) error {
	return errWrongFormat{s: s}
}

func IsErrWrongFormat(err error) bool {
	if _, ok := err.(errWrongFormat); ok {
		return true
	}
	return false
}
