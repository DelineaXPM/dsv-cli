package errors_test

import (
	serrors "errors"
	"testing"
	"thy/errors"

	"github.com/stretchr/testify/assert"
)

func TestNewS(t *testing.T) {
	err := errors.NewS("test string")
	assert.Equal(t, "test string", err.String())
}

func TestNewF(t *testing.T) {
	err := errors.NewF("test %s %d", "hi", 5)
	assert.Equal(t, "test hi 5", err.String())
}

func TestGrow(t *testing.T) {
	serr := serrors.New("A")
	err := errors.New(serr).Grow("B").Grow("C").Grow("D")
	assert.Equal(t, "D\nC\nB\nA", err.Error())
}

func TestAnd(t *testing.T) {
	serr := serrors.New("A")
	err := errors.New(serr).Grow("B")
	err2 := errors.NewS("C").Grow("D").And(err)
	assert.Equal(t, "D\nC\nand\nB\nA", err2.Error())
}

func TestOr(t *testing.T) {
	var err1 *errors.ApiError
	err2 := errors.NewS("error2")
	err3 := err1.Or(err2)
	assert.Equal(t, err2, err3)

	err1 = errors.NewS("error1")
	err2 = errors.NewS("error2")
	err3 = err1.Or(err2)
	assert.Equal(t, err1, err3)

}

func TestConvert(t *testing.T) {
	serr := serrors.New("An error")
	b := []byte{1, 2, 3}
	b2, err := errors.Convert(b, serr)
	assert.Equal(t, b, b2)
	assert.Equal(t, serr.Error(), err.Error())
}
