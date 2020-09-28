package version

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"thy/store"
)

const url = "https://dsv.thycotic.com/cli-version.json"
const checkFrequencyDays = 3

type update struct {
	latest string
	link   string
}

func (e update) String() string {
	return fmt.Sprintf("Consider upgrading the CLI to version %s - download available at %s", e.latest, e.link)
}

// getLatestVersionInfo makes a request to the downloads site and returns information on the most recent version of the CLI.
func getLatestVersionInfo() ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	return body, err
}

func updateCache(f *os.File, data []byte, fileNew bool) error {
	date := time.Now().Format("2006-Jan-02")
	output := fmt.Sprintf("%s\n%s", date, data)
	if !fileNew {
		// Overwrite the contents of the file with current date (most recent check for update) and version info.
		f.Truncate(0)
		f.Seek(0, 0)
	}
	_, err := f.WriteString(output)
	return err
}

// CheckLatestVersion checks if the user is running the latest available version of the CLI.
// It creates a file in the default `thy` directory. In the file, it stores the date of the last check (network request),
// the newest available version and a collection of download links for all platforms.
// If the next check occurs before the given number of days passed since the last check, then CheckLatestVersion does not
// make an HTTP request to check for a new available version of the CLI. It caches info about the update and reads it
// from the file to remind the user of an update without contacting the download server.
func CheckLatestVersion() (*update, error) {
	thyDir, err := store.GetDefaultPath()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(thyDir, ".update")
	f, err := os.OpenFile(path, os.O_RDWR, 0644)
	defer f.Close()

	fileNew := err != nil
	if fileNew {
		f, err = os.Create(path)
		if err != nil {
			return nil, err
		}
		return checkWithRequest(f, true)
	} else {
		raw, err := ioutil.ReadAll(f)
		if err != nil {
			return nil, err
		}
		contents := strings.Split(string(raw), "\n")
		date, err := time.Parse("2006-Jan-02", contents[0])
		if err != nil {
			return nil, err
		}
		if time.Now().After(date.AddDate(0, 0, checkFrequencyDays)) {
			return checkWithRequest(f, false)
		}
		// Check if old update still exists.
		v, links, err := extractVersionAndLinks([]byte(contents[1]))
		if err != nil {
			return nil, err
		}
		return check(v, links)
	}
}

func extractVersionAndLinks(source []byte) (string, map[string]interface{}, error) {
	var m map[string]interface{}
	err := json.Unmarshal(source, &m)
	if err != nil {
		return "", nil, err
	}
	var errJSON = errors.New("version info is not properly formatted")
	v, ok := m["latest"].(string)
	if !ok {
		return "", nil, errJSON
	}
	links := m["links"].(map[string]interface{})
	if !ok {
		return "", nil, errJSON
	}
	return v, links, nil
}

func checkWithRequest(f *os.File, fileNew bool) (*update, error) {
	log.Println("Attempting to query the download server for a CLI update.")
	info, err := getLatestVersionInfo()
	if err != nil {
		return nil, err
	}
	latest, links, err := extractVersionAndLinks(info)
	if err != nil {
		return nil, err
	}

	if err = updateCache(f, info, fileNew); err != nil {
		return nil, err
	}
	return check(latest, links)
}

// isVersionOutdated semantically compares target and latest versions
// to check if the target version is outdated.
func isVersionOutdated(target, latest string) bool {
	// Only interested in the leading part before the first dash.
	target = strings.Split(target, "-")[0]
	latest = strings.Split(latest, "-")[0]
	if target == "undefined" || target == latest {
		return false
	}

	t := strings.Split(target, ".")
	l := strings.Split(latest, ".")
	for i := range t {
		n1, _ := strconv.Atoi(t[i])
		n2, _ := strconv.Atoi(l[i])
		if n1 == n2 {
			continue
		} else if n1 < n2 {
			return true
		} else {
			return false
		}
	}
	return false
}

func check(latest string, links map[string]interface{}) (*update, error) {
	if isVersionOutdated(Version, latest) {
		link, err := getLinkForOS(links)
		if err != nil {
			return nil, err
		}
		return &update{latest, link}, nil
	}
	return nil, nil
}

// getLinkForOS tries to find a download link for the underlying OS in a collection of links for all platforms.
func getLinkForOS(links map[string]interface{}) (string, error) {
	for k := range links {
		if k == runtime.GOOS+"/"+runtime.GOARCH {
			return links[k].(string), nil
		}
	}
	return "", errors.New("links malformed")
}
