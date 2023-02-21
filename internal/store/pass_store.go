// could do conditional linux build, but no actual hard linux dependencies
package store

import (
	ch "github.com/DelineaXPM/dsv-cli/internal/store/credential-helpers"
)

func NewPassStore() Store {
	return NewSecureStore(&ch.Pass{})
}
