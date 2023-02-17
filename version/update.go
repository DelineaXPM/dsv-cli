package version

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/DelineaXPM/dsv-cli/internal/store"
)

const (
	versionsURL          = "https://dsv.secretsvaultcloud.com/cli-version.json"
	checkFrequencyDays   = 3
	cacheFileName        = ".update"
	dateLayout           = "2006-Jan-02"
	updatePatternMessage = "Consider upgrading the CLI to version %s - download available at %s"
)

type update struct {
	latest string
	link   string
}

type latestInfo struct {
	Latest string            `json:"latest"`
	Links  map[string]string `json:"links"`
}

func CheckLatestVersion() (*update, error) {
	thyDir, err := store.GetDefaultPath()
	if err != nil {
		return nil, err
	}
	cacheFilePath := filepath.Join(thyDir, cacheFileName)
	platform := runtime.GOOS + "/" + runtime.GOARCH
	latest := readCache(cacheFilePath)
	if latest == nil {
		pageContent, err := fetchContent(versionsURL)
		if err != nil {
			return nil, err
		}
		latest = &latestInfo{}
		err = json.Unmarshal(pageContent, latest)
		if err != nil {
			return nil, err
		}
		actualLink, ok := latest.Links[platform]
		if !ok {
			return nil, errors.New("links malformed")
		}
		latest.Links = map[string]string{platform: actualLink}
		pageContent, err = json.Marshal(latest)
		if err != nil {
			return nil, err
		}
		updateCache(cacheFilePath, pageContent)
	}
	if isVersionOutdated(Version, latest.Latest) {
		actualLink, ok := latest.Links[platform]
		if !ok {
			return nil, errors.New("links malformed")
		}
		return &update{latest: latest.Latest, link: actualLink}, nil
	}
	return nil, nil
}

// readCache returns parsed info from cache file
func readCache(updateFilePath string) *latestInfo {
	_, err := os.Stat(updateFilePath)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	fileContent, err := os.ReadFile(updateFilePath)
	if err != nil {
		log.Printf("Failed to read content from file %s (%s) ", updateFilePath, err.Error())
		return nil
	}

	const fileParts int = 2
	contents := strings.SplitN(string(fileContent), "\n", fileParts)
	if len(contents) < fileParts {
		log.Printf("Wrong file %s format", updateFilePath)
		return nil
	}

	date, err := time.Parse(dateLayout, contents[0])
	if err != nil {
		log.Printf("Wrong date format: '%s'", contents[0])
		return nil
	}

	if time.Now().After(date.AddDate(0, 0, checkFrequencyDays)) {
		log.Println("Cached content too old")
		return nil
	}

	versions := &latestInfo{}
	err = json.Unmarshal([]byte(contents[1]), versions)
	if err != nil {
		log.Printf("Wrong file content: '%s'", contents[1])
		return nil
	}

	return versions
}

// updateCache updates content in the cache file
func updateCache(cacheFilePath string, content []byte) {
	fileContent := fmt.Sprintf("%s\n%s", time.Now().Format(dateLayout), string(content))
	err := os.WriteFile(cacheFilePath, []byte(fileContent), os.FileMode(0o644))
	if err != nil {
		// Only log error and continue
		log.Printf("Unsuccessfully update of the cache file %s (%s) ", cacheFilePath, err)
	}
}

// fetchContent retrieves content of the URL
func fetchContent(urlToFetch string) ([]byte, error) {
	log.Println("Attempting to query the download server for a CLI update.")
	resp, err := http.Get(urlToFetch)
	if err != nil {
		return nil, err
	}
	pageContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	_ = resp.Body.Close()

	return pageContent, nil
}

func (e update) String() string {
	return fmt.Sprintf(updatePatternMessage, e.latest, e.link)
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
		if t[i] == l[i] {
			continue
		}
		n1, _ := strconv.Atoi(t[i])
		n2, _ := strconv.Atoi(l[i])
		return n1 < n2
	}
	return false
}
