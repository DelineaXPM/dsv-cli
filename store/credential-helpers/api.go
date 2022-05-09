package credhelpers

// StoreHelper is the interface a credentials store helper must implement.
type StoreHelper interface {
	// Add appends credentials to the store.
	Add(*Credentials) error
	// Delete removes credentials from the store.
	Delete(serverURL string) error
	// Get retrieves credentials from the store.
	// It returns username and secret as strings.
	Get(serverURL string) (string, string, error)
	// List returns the stored serverURLs and their associated usernames.
	List(prefix string) (map[string]string, error)
	// GetName returns the friendly name of helper.
	GetName() string
}
