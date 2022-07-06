// Controlled via the Makefile and git tag reference for Version
package version

import (
	"strconv"
	"time"
)

var (
	Version   string = "undefined"
	BuildDate string = "undefined"
	GitCommit string = "undefined"
)

func GetBuildDate() string {
	t, err := strconv.ParseInt(BuildDate, 10, 64)
	if err != nil {
		return BuildDate
	}
	unixTime := time.Unix(t, 0)
	return unixTime.UTC().Format(time.RFC3339)
}
