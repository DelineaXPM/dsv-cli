package predictor

import (
	"testing"

	"github.com/posener/complete"
	"github.com/stretchr/testify/assert"
)

func TestPrefixFilePredictor(t *testing.T) {
	args := complete.Args{}
	args.Last = "@"
	pred := NewPrefixFilePredictor("*")
	preds := pred.Predict(args)
	assert.NotEqual(t, 0, len(preds))

	args.Last = ""
	preds = pred.Predict(args)
	assert.Equal(t, 0, len(preds))

	args.Last = "@"
	pred = NewPrefixFilePredictor("*.fjksljflds")
	preds = pred.Predict(args)
	assert.Equal(t, 0, len(preds))
}
