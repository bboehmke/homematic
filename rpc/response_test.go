package rpc

import (
	"encoding/xml"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResponse_FirstParam(t *testing.T) {
	ass := assert.New(t)

	response := Response{}
	ass.Nil(response.FirstParam())

	response.Params = []interface{}{42}
	ass.Equal(42, response.FirstParam())
}

func TestResponse_UnmarshalXML(t *testing.T) {
	ass := assert.New(t)

	data := `<methodResponse>
  <params>
    <param>
      <value>
        <string>test</string>
      </value>
    </param>
  </params>
</methodResponse>`

	var response Response
	err := xml.Unmarshal([]byte(data), &response)
	ass.NoError(err)
}

func TestFault_UnmarshalXML(t *testing.T) {
	ass := assert.New(t)

	data := `<fault>
  <value>
    <struct>
      <member>
        <name>faultCode</name>
        <value><int>4</int></value>
      </member>
      <member>
        <name>faultString</name>
        <value><string>Too many parameters</string></value>
      </member>
    </struct>
  </value>
</fault>`

	var fault Fault
	err := xml.Unmarshal([]byte(data), &fault)
	ass.NoError(err)
	ass.Equal(int32(4), fault.Code)
	ass.Equal("Too many parameters", fault.String)
}

var valueTestData = []struct {
	name  string
	xml   string
	value interface{}
}{{
	"string",
	"<string>test</string>",
	"test",
}, {
	"string_raw",
	"test",
	"test",
}, {
	"int",
	"<int>42</int>",
	int32(42),
}, {
	"i4",
	"<i4>42</i4>",
	int32(42),
}, {
	"bool_true",
	"<boolean>1</boolean>",
	true,
}, {
	"bool_false",
	"<boolean>0</boolean>",
	false,
}, {
	"double",
	"<double>1.2</double>",
	float64(1.2),
}, {
	"array",
	`<array>
  <data>
    <value><i4>111</i4></value>
    <value><i4>222</i4></value>
  </data>
</array>`,
	[]interface{}{
		int32(111), int32(222),
	},
}, {
	"struct",
	`<struct>
  <member>
    <name>aaa</name>
    <value><i4>111</i4></value>
  </member>
  <member>
    <name>bbb</name>
    <value><i4>222</i4></value>
  </member>
</struct>`,
	map[string]interface{}{
		"aaa": int32(111),
		"bbb": int32(222),
	},
}}

func TestValue_Interface(t *testing.T) {

	for _, d := range valueTestData {
		t.Run(d.name, func(st *testing.T) {
			ass := assert.New(st)

			var v value
			data := fmt.Sprintf("<value>%s</value>", d.xml)
			err := xml.Unmarshal([]byte(data), &v)
			ass.NoError(err)
			ass.Equal(d.value, v.Interface())
		})
	}
}
