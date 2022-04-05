package predictor

import "github.com/posener/complete"

type ActionTypePredictor struct{}

func (p ActionTypePredictor) Predict(a complete.Args) (prediction []string) {
	return []string{
		"share",
		"create",
		"update",
		"delete",
		"read",
	}
}
