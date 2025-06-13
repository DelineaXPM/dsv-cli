package vault

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

// handleResponse processes the response according to the HTTP status
func handleResponse(res *http.Response, err error) ([]byte, error) {
	if err != nil { // fall-through if there was an underlying err
		return nil, err
	}

	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	// if the response was 2xx then return it, otherwise, consider it an error
	if res.StatusCode > 199 && res.StatusCode < 300 {
		return data, nil
	}
	return nil, fmt.Errorf("%s: %s", res.Status, string(data))
}
