package predictor

import (
	"github.com/posener/complete"

	"thy/auth"
)

type GcpAuthTypePredictor struct{}

func (p GcpAuthTypePredictor) Predict(a complete.Args) (prediction []string) {
	return []string{string(auth.GcpGceAuth), string(auth.GcpIamAuth)}
}
