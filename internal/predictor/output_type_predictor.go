package predictor

import (
	"github.com/posener/complete"

	cst "thy/constants"
	"thy/format"
)

type OutputTypePredictor struct{}

func (p OutputTypePredictor) Predict(a complete.Args) (prediction []string) {
	return []string{string(format.StdOut), string(format.ClipBoard), cst.OutFilePrefix}
}
