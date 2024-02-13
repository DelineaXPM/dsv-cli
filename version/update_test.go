package version

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// cacheDirectory is for local files that we generate during tests so easy to review and cleanup.
var cacheDirectory = filepath.Join("../", ".artifacts", "tests") //nolint:gochecknoglobals // for testing this is ok to have fixed output location
// PermissionUserReadWriteExecute is the permissions for the cache directory.
const PermissionUserReadWriteExecute = 0o0700

func createTestDirectory(t *testing.T) {
	t.Helper()

	if err := os.MkdirAll(cacheDirectory, PermissionUserReadWriteExecute); err != nil {
		t.Fatalf("failure to create cache directory: %s: err: %v", cacheDirectory, err)
	}
	t.Logf("cacheDirectory created: %s", cacheDirectory)
}

func TestIsVersionOutdated(t *testing.T) {
	cases := []struct {
		target   string
		latest   string
		outdated bool
	}{
		{target: "1.0.0", latest: "2.0.0", outdated: true},
		{target: "1.0.0", latest: "1.1.0", outdated: true},
		{target: "1.0.0", latest: "1.0.1", outdated: true},
		{target: "1.8.1", latest: "1.9.0", outdated: true},

		{target: "1.0.0", latest: "1.0.0"},
		{target: "2.0.0", latest: "1.0.0"},
		{target: "12.0.0", latest: "2.0.0"},
		{target: "undefined", latest: "1.9.1"},
		{target: "1.38.0", latest: "1.37.5"},
	}

	for _, tc := range cases {
		if got := isVersionOutdated(tc.target, tc.latest); got != tc.outdated {
			t.Errorf("isVersionOutdated(%s, %s) = %v, want %v", tc.target, tc.latest, got, tc.outdated)
		}
	}
}

func TestReadCache(t *testing.T) {
	testCases := []struct {
		name     string
		filename string
		content  string
		result   *latestInfo
	}{
		{
			name:     "File not exists",
			filename: "",
			content:  ``,
			result:   nil,
		},
		{
			name:     "Content with one string",
			filename: "one_string_content_*.json",
			content:  time.Now().Format(dateLayout),
			result:   nil,
		},
		{
			name:     "Content with wrong date format",
			filename: "wrong_date_*.json",
			content:  "wrong_date_content\n",
			result:   nil,
		},
		{
			name:     "Content with wrong JSON data",
			filename: "wrong_json_data_*.json",
			content:  time.Now().Format(dateLayout) + "\nWrong JSON string",
			result:   nil,
		},
		{
			name:     "Old cached content",
			filename: "old_cached_content_*.json",
			// content:  time.Now().Add(checkFrequencyDays*-2*time.Hour*24).Format(dateLayout) + "\nWrong JSON string",
			content: time.Now().Add(checkFrequencyDays*-2*time.Hour*24).Format(dateLayout) +
				"\n" + `{"latest":"1.40.6","links": {"darwin/amd64":"https://dsv.secretsvaultcloud.com/downloads/cli/1.40.6/dsv-darwin-x64", "linux/amd64":"https://dsv.secretsvaultcloud.com/downloads/cli/1.40.6/dsv-linux-x64"}}` + "\n\n",
			result: nil,
		},
		{
			name:     "Correct case",
			filename: "correct_case_*.json",
			content: time.Now().Add(-1*(checkFrequencyDays-1)*time.Hour*24).Format(dateLayout) +
				"\n" + `{"latest":"1.40.6","links": {"darwin/amd64":"https://dsv.secretsvaultcloud.com/downloads/cli/1.40.6/dsv-darwin-x64"}}` + "\n\n",
			result: &latestInfo{
				Latest: "1.40.6",
				Links:  map[string]string{"darwin/amd64": "https://dsv.secretsvaultcloud.com/downloads/cli/1.40.6/dsv-darwin-x64"},
			},
		},
	}
	createTestDirectory(t)
	t.Parallel()
	for _, tc := range testCases {
		testCase := tc
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			fName := "some_not_existed_file"
			if testCase.filename != "" {
				temporaryFile, err := os.CreateTemp(cacheDirectory, testCase.filename)
				assert.NoError(t, err)
				defer func(name string) {
					if os.Getenv("KEEP_FILES") != "" {
						t.Log("KEEP_FILES was set, so local test files are kept for debugging")
					} else {
						if err := os.Remove(name); err != nil {
							t.Logf("error during delete temporary file: %s", err)
						}
					}
				}(temporaryFile.Name())

				_, err = temporaryFile.WriteString(testCase.content) // not sure why this keeps going away during fmt
				assert.NoError(t, err)

				err = temporaryFile.Close()
				assert.NoError(t, err)

				fName = temporaryFile.Name()
			}
			result := readCache(fName)
			assert.Equal(t, testCase.result, result)
		})
	}
}

func TestFetchContent(t *testing.T) {
	testCases := []struct {
		content []byte
	}{
		{content: []byte("first")},
		{content: []byte("second")},
	}

	for _, testCase := range testCases {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write(testCase.content)
			assert.NoError(t, err)
		}))
		defer srv.Close()
		result, err := fetchContent(srv.URL)
		assert.NoError(t, err)
		assert.Equal(t, testCase.content, result)
	}
}

func TestUpdateCache(t *testing.T) {
	content := []byte("some content")
	createTestDirectory(t)
	temporaryDirName, err := os.MkdirTemp(cacheDirectory, "")
	assert.NoError(t, err)
	defer os.Remove(temporaryDirName)

	temporaryFileName := filepath.Join(temporaryDirName, cacheFileName)
	defer func() {
		if os.Getenv("KEEP_FILES") != "" {
			t.Log("KEEP_FILES was set, so local test files are kept for debugging")
		} else {
			if err := os.Remove(temporaryFileName); err != nil {
				t.Logf("error during delete temporary file: %s", err)
			}
		}
	}()

	updateCache(temporaryFileName, content)

	fileContent, err := os.ReadFile(temporaryFileName)
	assert.NoError(t, err)

	expectedContent := fmt.Sprintf("%s\n%s", time.Now().Format(dateLayout), content)
	assert.Equal(t, []byte(expectedContent), fileContent)
}

func TestUpdateString(t *testing.T) {
	u := update{
		latest: "a",
		link:   "b",
	}
	expected := fmt.Sprintf(updatePatternMessage, u.latest, u.link)
	assert.Equal(t, expected, u.String())
}
