package predictor

import (
	"github.com/posener/complete"

	cst "thy/constants"
)

type EncodingTypePredictor struct{}

func (p EncodingTypePredictor) Predict(a complete.Args) (prediction []string) {
	return []string{cst.Json, cst.YamlShort}
}
