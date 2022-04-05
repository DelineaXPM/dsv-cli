package format

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	cst "thy/constants"
	"thy/errors"
	"thy/utils"

	"github.com/fatih/color"
	"github.com/hokaccha/go-prettyjson"
	"github.com/savaki/jq"
	"github.com/spf13/viper"
	"github.com/tidwall/pretty"
	yaml "gopkg.in/yaml.v2"
)

type WriterType string

const (
	StdOut    = WriterType("stdout")
	File      = WriterType("file")
	ClipBoard = WriterType("clip")
)

type OutClient interface {
	WriteResponse(data []byte, err *errors.ApiError)
	Fail(err error)
	FailE(err *errors.ApiError)
	FailS(errString string)
	FailF(errFmt string, args ...interface{})
}

type outClient struct {
	outWriter io.Writer
	errWriter io.Writer
}

func NewDefaultOutClient() OutClient {
	var outWriter io.Writer
	outputType := viper.GetString(cst.Output)
	if outputType == "" {
		outputType = "stdout"
	}
	var wt WriterType
	if strings.HasPrefix(outputType, cst.OutFilePrefix) {
		wt = File
	} else {
		wt = WriterType(outputType)
	}

	if wt == StdOut {
		outWriter = os.Stdout
	} else if wt == ClipBoard {
		outWriter = clipWriter{}
	} else if wt == File {
		filePath := outputType[len(cst.OutFilePrefix):]
		outWriter = NewFileWriter(filePath)
	}
	return NewOutClient(outWriter, os.Stderr)
}

func NewOutClient(sw io.Writer, ew io.Writer) OutClient {
	c := outClient{}
	if sw == nil {
		c.outWriter = os.Stdout
	} else {
		c.outWriter = sw
	}
	if ew == nil {
		c.errWriter = os.Stderr
	} else {
		c.errWriter = ew
	}
	return &c
}

func IsJson(b []byte) bool {
	var j interface{}
	if err := json.Unmarshal(b, &j); err == nil {
		return true
	} else {
		return false
	}
}

func JsonMarshal(obj interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	// otherwise escapes '<' and '>' which we dont want
	enc.SetEscapeHTML(false)
	if err := enc.Encode(&obj); err != nil {
		return nil, errors.New(err)
	}
	b := buf.Bytes()
	if len(b) >= 2 {
		// chop newline
		return b[:len(b)-1], nil
	}
	return b, nil
}

func prettifyWindows(data []byte) ([]byte, *errors.ApiError) {
	isJson := encodingIsJson()
	if isJson {
		return pretty.Pretty(data), nil
	} else {
		return toYaml(data)
	}
}

func toYaml(data []byte) ([]byte, *errors.ApiError) {
	var obj interface{}
	err := json.Unmarshal(data, &obj)
	if err != nil {
		return nil, errors.New(err).Grow("Failed marshalling data as json prior to conversion to yaml")
	}
	return errors.Convert(yaml.Marshal(obj))
}

func encodingIsJson() bool {
	encoding := viper.GetString(cst.Encoding)
	if encoding == "" || (encoding != cst.Json && encoding != cst.Yaml && encoding != cst.YamlShort) {
		encoding = cst.Json
	}
	return encoding == cst.Json
}

func prettifyUnix(data []byte, colorify bool) ([]byte, *errors.ApiError) {
	isJson := encodingIsJson()
	if isJson {
		formatter := prettyjson.NewFormatter()
		formatter.KeyColor = color.New(color.FgMagenta, color.FgCyan)
		if !colorify {
			formatter.DisabledColor = true
		}
		return errors.Convert(formatter.Format(data))
	} else {
		return toYaml(data)
	}
}

func (c *outClient) FailE(err *errors.ApiError) {
	c.WriteResponse(nil, err)
}

func (c *outClient) Fail(err error) {
	c.WriteResponse(nil, errors.New(err))
}

func (c *outClient) FailS(errString string) {
	c.WriteResponse(nil, errors.NewS(errString))
}

func (c *outClient) FailF(errFmt string, args ...interface{}) {
	c.WriteResponse(nil, errors.NewF(errFmt, args...))
}

func (c *outClient) WriteResponse(data []byte, err *errors.ApiError) {
	dataNil := len(data) <= 0
	if dataNil && err == nil {
		return
	}
	isBeautify := viper.GetBool(cst.Beautify)

	if !dataNil {
		var errFilter *errors.ApiError
		data, errFilter = FilterResponse(data)
		err = err.Or(errFilter)
	}
	dataFmted, errFmted := FormatResponse(data, err, isBeautify)

	if _, printErr := fmt.Fprint(c.outWriter, dataFmted); printErr != nil && len(errFmted) <= 0 {
		errFmted = formatError(printErr)
	}
	fmt.Fprint(c.errWriter, errFmted)
}

func FilterResponse(data []byte) ([]byte, *errors.ApiError) {
	filter := viper.GetString(cst.Filter)
	if filter == "" {
		return data, nil
	}
	// TODO : NH - this jq library doesnt support advanced operations. would be nice to make one that does
	op, err := jq.Parse(filter)
	if err != nil {
		// TODO : Should we return data or original if filter fails? If we do should only write if stderr and stdout
		//  destinations not the same (console)
		return nil, errors.New(err).Grow(fmt.Sprintf("Invalid filter (%s) on data:\n%s", filter, string(data)))
	}
	data, err = op.Apply(data)
	return data, errors.New(err).Grow("Failed to apply the filter to the data")
}

func FormatResponse(data []byte, err *errors.ApiError, isBeautify bool) (dataStr string, errStr string) {
	if err != nil {
		if IsJson([]byte(err.Error())) {
			shouldColor := false
			if fmted, fmtErr := BeautifyBytes([]byte(err.Error()), &shouldColor); fmtErr == "" {
				errStr = fmted + "\n"
			}
		}
		if errStr == "" {
			errStr = formatError(err)
		}
	}

	data = bytes.Trim(data, `"`)
	if bytes.Equal(data, []byte("{}")) {
		data = []byte{}
	}
	if len(data) > 0 {
		if isBeautify && IsJson(data) {
			if fmted, fmtErr := BeautifyBytes(data, nil); fmtErr != "" && err == nil {
				errStr = fmtErr
			} else {
				dataStr = fmted
			}
		} else {
			dataStr = string(data)
		}
	}
	if len(dataStr) > 0 && isBeautify {
		if !strings.HasSuffix(dataStr, "\n") {
			dataStr = dataStr + "\n"
		}
	}
	return dataStr, errStr
}

func BeautifyBytes(data []byte, colorify *bool) (dataStr, errStr string) {
	isWindows := utils.GetEnvProviderFunc().GetOs() == "windows"
	outputType := viper.GetString(cst.Output)
	outputStdOut := outputType == "" || outputType == string(StdOut)
	shouldColor := outputStdOut
	if colorify != nil {
		shouldColor = *colorify
	}

	var beautifyErr *errors.ApiError
	if isWindows {
		data, beautifyErr = prettifyWindows(data)
		dataStr = string(data)
	} else {
		data, beautifyErr = prettifyUnix(data, shouldColor)
		dataStr = string(data)
	}
	if beautifyErr != nil {
		errStr = formatError(beautifyErr.Grow("Failed to present data"))
	}
	return dataStr, errStr
}

func formatError(err error) string {
	fmtErr := err.Error()
	if len(fmtErr) > 0 {
		fmtErr = fmtErr + "\n"
	}
	return fmtErr
}
