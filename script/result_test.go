package script

import (
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResult_UnmarshalXML(t *testing.T) {
	ass := assert.New(t)

	var xmlData = `<xml><a>aaa</a><b>bbb</b></xml>`
	var res Result
	ass.NoError(xml.Unmarshal([]byte(xmlData), &res))
	ass.Equal(Result{
		"a": "aaa",
		"b": "bbb",
	}, res)
}

func TestResult_GetMap(t *testing.T) {
	ass := assert.New(t)

	res := Result{
		"a": "aaa",
		"b": "a=1\nb=2\n",
	}

	ass.Nil(res.GetMap("c"))
	ass.Equal(map[string]string{
		"a": "1",
		"b": "2",
	}, res.GetMap("b"))

}
