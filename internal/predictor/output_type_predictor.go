package predictor

import (
	"github.com/posener/complete"

	"github.com/DelineaXPM/dsv-cli/format"
)

type OutputTypePredictor struct{}

func (p OutputTypePredictor) Predict(a complete.Args) (prediction []string) {
	return []string{format.OutToStdout, format.OutToClip, format.OutToFilePrefix}
}
