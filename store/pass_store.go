// could do conditional linux build, but no actual hard linux dependencies
package store

import (
	ch "thy/store/credential-helpers"
)

func NewPassStore() Store {
	return NewSecureStore(&ch.Pass{})
}
