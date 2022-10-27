package format_test

import (
	"bytes"
	"fmt"
	"testing"

	cst "github.com/DelineaXPM/dsv-cli/constants"
	"github.com/DelineaXPM/dsv-cli/errors"
	"github.com/DelineaXPM/dsv-cli/format"
	"github.com/DelineaXPM/dsv-cli/utils"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestIsJson(t *testing.T) {
	assert.False(t, format.IsJson([]byte("hi")))
	assert.False(t, format.IsJson([]byte("{")))
	assert.False(t, format.IsJson([]byte("}")))
	assert.False(t, format.IsJson([]byte(`{"A":5},`)))

	assert.True(t, format.IsJson([]byte(`1`)))
	assert.True(t, format.IsJson([]byte(`[1, "a", null]`)))
	assert.True(t, format.IsJson([]byte(`"hi"`)))
	assert.True(t, format.IsJson([]byte(`{"A":5}`)))
}

func TestJsonMarshal(t *testing.T) {
	var obj = struct {
		Name string `json:"name"`
		Val  string `json:"val"`
	}{
		Name: "property with symbols and lt/gt",
		Val:  "<prop> val!@#$%^&*(\"",
	}
	expectedString := "{\"name\":\"property with symbols and lt/gt\",\"val\":\"<prop> val!@#$%^&*(\\\"\"}"
	m, err := format.JsonMarshal(obj)
	assert.Nil(t, err)
	assert.Equal(t, []byte(expectedString), m)
}

var writerTestCases = []struct {
	Data     string
	Err      *errors.ApiError
	Beautify bool
	Encoding string
	Expected string
}{
	{
		Data:     "hi",
		Err:      nil,
		Expected: "hi",
	},
	{
		Data:     "{\"An\":invalidjson}",
		Err:      nil,
		Expected: "{\"An\":invalidjson}",
	},
	{
		Data:     "{\"A\":{\"B\":\"validjson\"}}",
		Err:      nil,
		Beautify: false,
		Expected: "{\"A\":{\"B\":\"validjson\"}}",
	},
	{
		Data:     "{\"A\":{\"B\":\"validjson\"}}",
		Err:      nil,
		Beautify: true,
		Expected: "{\n  \"A\": {\n    \"B\": \"validjson\"\n  }\n}\n",
	},
	{
		Data:     "",
		Err:      errors.NewS("but an error occurred"),
		Beautify: true,
		Expected: "but an error occurred\n",
	},
}

func TestWriterWriteResponse(t *testing.T) {
	for i, c := range writerTestCases {
		t.Run(fmt.Sprintf("case %d-B:%v-Err:%v-Data:%s", i, c.Beautify, c.Err, c.Data), func(t *testing.T) {
			viper.Set(cst.Beautify, c.Beautify)
			viper.Set(cst.Encoding, c.Encoding)
			testWriter := new(bytes.Buffer)
			w := format.NewOutClient(testWriter, testWriter)
			w.WriteResponse([]byte(c.Data), c.Err)
			written := testWriter.String()
			assert.Equal(t, c.Expected, written)
		})
	}

}

func TestFilterResponse_Prop(t *testing.T) {
	unfiltered := []byte("{\"A\":{\"B\":\"validjson\"}}")
	filter := ".A"
	viper.Set(cst.Filter, filter)
	filtered, err := format.FilterResponse(unfiltered)
	assert.Nil(t, err)
	assert.Equal(t, "{\"B\":\"validjson\"}", string(filtered))
}

func TestFilterResponse_Array(t *testing.T) {
	unfiltered := []byte("{\"A\":[{\"B\":\"validjson1\"}, {\"C\":\"validjson2\"}]}")
	filter := ".A.[1].C"
	viper.Set(cst.Filter, filter)
	filtered, err := format.FilterResponse(unfiltered)
	assert.Nil(t, err)
	assert.Equal(t, "\"validjson2\"", string(filtered))
}

func TestFormatResponse_Windows(t *testing.T) {
	utils.GetEnvProviderFunc = func() utils.EnvProvider {
		return utils.EnvFunc(func() string {
			return "windows"
		})
	}
	unformatted := "{\"A\":{\"B\":\"validjson\"}}"
	dataFmt, errFmt := format.FormatResponse([]byte(unformatted), nil, true)
	assert.Equal(t, dataFmt, "{\n  \"A\": {\n    \"B\": \"validjson\"\n  }\n}\n")
	assert.Equal(t, errFmt, "")
	utils.GetEnvProviderFunc = func() utils.EnvProvider {
		return utils.NewEnvProvider()
	}
}

func TestFormatResponse_Linux(t *testing.T) {
	utils.GetEnvProviderFunc = func() utils.EnvProvider {
		return utils.EnvFunc(func() string {
			return "linux"
		})
	}
	unformatted := "{\"A\":{\"B\":\"validjson\"}}"
	dataFmt, errFmt := format.FormatResponse([]byte(unformatted), nil, true)
	assert.Equal(t, dataFmt, "{\n  \"A\": {\n    \"B\": \"validjson\"\n  }\n}\n")
	assert.Equal(t, errFmt, "")
	utils.GetEnvProviderFunc = func() utils.EnvProvider {
		return utils.NewEnvProvider()
	}
}
