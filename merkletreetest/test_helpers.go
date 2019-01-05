package merkletreetest

import (
	"testing"
)

type ExtendedTesting struct {
	*testing.T
}

func (et *ExtendedTesting) Assert(condition bool, msg ...interface{}) {

	if !condition {
		et.Error(msg...)
	}
}

func WrapTesting(t *testing.T) *ExtendedTesting {
	return &ExtendedTesting{t}
}
