package errors

import (
	serrors "errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	apiErr := New(nil)
	assert.Nil(t, apiErr)

	stdErr := serrors.New("")
	apiErr = New(stdErr)
	assert.Equal(t, "", apiErr.String())
	assert.Equal(t, "", apiErr.Error())

	stdErr = serrors.New("test string")
	apiErr = New(stdErr)
	assert.Equal(t, "test string", apiErr.String())
	assert.Equal(t, "test string", apiErr.Error())
}

func TestNewF(t *testing.T) {
	apiErr := NewF("test %s %d", "hi", 5)
	assert.Equal(t, "test hi 5", apiErr.String())
	assert.Equal(t, "test hi 5", apiErr.Error())
}

func TestNewS(t *testing.T) {
	apiErr := NewS("")
	assert.Nil(t, apiErr)

	apiErr = NewS("test string")
	assert.Equal(t, "test string", apiErr.String())
	assert.Equal(t, "test string", apiErr.Error())
}

func TestGrow(t *testing.T) {
	var apiErr *ApiError
	apiErr = apiErr.Grow("A").Grow("B").Grow("C")
	assert.Equal(t, "", apiErr.String())
	assert.Equal(t, "", apiErr.Error())

	apiErr = NewS("A").Grow("B").Grow("C")
	assert.Equal(t, "C\nB\nA", apiErr.String())
	assert.Equal(t, "C\nB\nA", apiErr.Error())
}

func TestOr(t *testing.T) {
	var apiErr1 *ApiError
	apiErr2 := NewS("error2")
	apiErr3 := apiErr1.Or(apiErr2)
	assert.Equal(t, apiErr2, apiErr3)

	apiErr1 = NewS("error1")
	apiErr2 = NewS("error2")
	apiErr3 = apiErr1.Or(apiErr2)
	assert.Equal(t, apiErr1, apiErr3)
}

func TestAnd(t *testing.T) {
	var apiErr, apiErr2 *ApiError
	apiErr2 = apiErr.And(apiErr2)
	assert.Equal(t, "", apiErr2.String())
	assert.Equal(t, "", apiErr2.Error())

	apiErr2 = NewS("A").Grow("B")
	apiErr2 = apiErr.And(apiErr2)
	assert.Equal(t, "B\nA", apiErr2.String())
	assert.Equal(t, "B\nA", apiErr2.Error())

	apiErr = NewS("A").Grow("B")
	apiErr2 = NewS("C").Grow("D").And(apiErr)
	assert.Equal(t, "D\nC\nand\nB\nA", apiErr2.String())
	assert.Equal(t, "D\nC\nand\nB\nA", apiErr2.Error())
}

func TestConvert(t *testing.T) {
	serr := serrors.New("An error")
	b := []byte{1, 2, 3}
	b2, err := Convert(b, serr)
	assert.Equal(t, b, b2)
	assert.Equal(t, serr.Error(), err.Error())
}
