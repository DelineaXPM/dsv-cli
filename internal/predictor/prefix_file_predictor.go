package predictor

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/posener/complete"

	cst "thy/constants"
)

type PrefixFilePredictor struct {
	filePattern string
}

func NewPrefixFilePredictor(pattern string) *PrefixFilePredictor {
	return &PrefixFilePredictor{filePattern: pattern}
}

func (p PrefixFilePredictor) Predict(a complete.Args) (prediction []string) {
	last := a.Last
	if !strings.HasPrefix(last, cst.CmdFilePrefix) {
		return []string{}
	}

	allowFiles := true
	prediction = predictFiles(a, p.filePattern, allowFiles, cst.CmdFilePrefix)

	// if the number of prediction is not 1, we either have many results or
	// have no results, so we return it.
	if len(prediction) != 1 {
		return
	}

	// only try deeper, if the one item is a directory
	if stat, err := os.Stat(prediction[0]); err != nil || !stat.IsDir() {
		return
	}

	a.Last = prediction[0]
	return predictFiles(a, p.filePattern, allowFiles, cst.CmdFilePrefix)
}

func predictFiles(a complete.Args, pattern string, allowFiles bool, prefix string) []string {
	if strings.HasSuffix(a.Last, "/..") {
		return nil
	}

	argsWithRealPath := complete.Args{
		All:           a.All,
		Completed:     a.Completed,
		Last:          a.Last[len(prefix):],
		LastCompleted: a.LastCompleted,
	}
	dir := argsWithRealPath.Directory()
	files := listFiles(dir, pattern, allowFiles, prefix)

	// add dir if match
	files = append(files, dir)

	return complete.PredictFilesSet(files).Predict(a)
}

func listFiles(dir, pattern string, allowFiles bool, prefix string) []string {
	// set of all file names
	m := map[string]bool{}

	// list files
	if files, err := filepath.Glob(filepath.Join(dir, pattern)); err == nil {
		for _, f := range files {
			if stat, err := os.Stat(f); err != nil || stat.IsDir() || allowFiles {
				m[f] = true
			}
		}
	}

	// list directories
	if dirs, err := ioutil.ReadDir(dir); err == nil {
		for _, d := range dirs {
			if d.IsDir() {
				m[filepath.Join(dir, d.Name())] = true
			}
		}
	}

	list := make([]string, 0, len(m))
	for k := range m {
		list = append(list, prefix+k)
	}
	return list
}
