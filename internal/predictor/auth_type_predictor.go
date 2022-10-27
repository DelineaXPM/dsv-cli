package predictor

import (
	"github.com/DelineaXPM/dsv-cli/auth"

	"github.com/posener/complete"
)

type AuthTypePredictor struct{}

func (p AuthTypePredictor) Predict(a complete.Args) (prediction []string) {
	return []string{
		string(auth.Password),
		string(auth.ClientCredential),
		string(auth.Certificate),
		string(auth.FederatedAws),
		string(auth.FederatedAzure),
		string(auth.FederatedGcp),
	}
}
