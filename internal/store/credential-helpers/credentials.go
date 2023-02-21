package credhelpers

import cst "github.com/DelineaXPM/dsv-cli/constants"

type Credentials struct {
	ServerURL string
	Username  string
	Secret    string
}

func externalToInternal(url string) string {
	return cst.StoreRoot + "-" + url
}

func internalToExternal(url string) string {
	toRemove := len(cst.StoreRoot + "-")
	return url[toRemove:]
}
