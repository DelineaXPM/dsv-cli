package predictor

import (
	"github.com/posener/complete"
)

type AuthProviderTypePredictor struct{}

func (p AuthProviderTypePredictor) Predict(a complete.Args) (prediction []string) {
	return []string{"azure", "aws", "gcp", "thycoticone"}
}
