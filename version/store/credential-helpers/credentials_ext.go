package credhelpers

import (
	cst "thy/constants"
)

func externalUrlToInternalUrl(url string) string {
	return cst.StoreRoot + "-" + url
}

func internalUrlToExternalUrl(url string) string {
	toRemove := len(cst.StoreRoot + "-")
	return url[toRemove:]
}
