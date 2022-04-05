package predictor

import "github.com/posener/complete"

type EffectTypePredictor struct{}

func (p EffectTypePredictor) Predict(a complete.Args) (prediction []string) {
	return []string{
		"allow",
		"deny",
	}
}
