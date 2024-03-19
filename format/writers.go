package format

import (
	serrors "errors"
	"io"
	"os"

	"github.com/atotto/clipboard"
)

type clipWriter struct{}

func (c clipWriter) Write(p []byte) (n int, err error) {
	if len(p) <= 0 {
		return 0, nil
	}
	data := string(p)
	if e := clipboard.WriteAll(data); e == nil {
		n = len(p)
	} else {
		err = serrors.New("Error writing to clipboard; " + e.Error())
	}
	return n, err
}

type fileWriter struct {
	filePath string
}

func NewFileWriter(filePath string) io.Writer {
	return &fileWriter{filePath}
}

//nolint:gosec,gomnd // need to test this in a future PR
func (w fileWriter) Write(p []byte) (n int, err error) {
	err = os.WriteFile(w.filePath, p, 0o664)
	if err == nil {
		n = len(p)
	}
	return n, err
}
