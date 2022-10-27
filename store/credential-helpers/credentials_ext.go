package credhelpers

import (
	cst "github.com/DelineaXPM/dsv-cli/constants"
)

func externalUrlToInternalUrl(url string) string {
	return cst.StoreRoot + "-" + url
}

func internalUrlToExternalUrl(url string) string {
	toRemove := len(cst.StoreRoot + "-")
	return url[toRemove:]
}
