package rpc

import (
	"encoding/xml"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/beevik/etree"
	"github.com/stretchr/testify/assert"
)

func TestFault_toMap(t *testing.T) {
	ass := assert.New(t)

	fault := Fault{
		Code:   42,
		String: "test",
	}

	ass.Equal(map[string]interface{}{
		"faultCode":   fault.Code,
		"faultString": fault.String,
	}, fault.toMap())
}

func TestResponse_FirstParam(t *testing.T) {
	ass := assert.New(t)

	response := Response{}
	ass.Nil(response.FirstParam())

	response.Params = []interface{}{42}
	ass.Equal(42, response.FirstParam())
}

func TestResponse_MarshalXML(t *testing.T) {
	ass := assert.New(t)

	resp := Response{
		Params: []interface{}{
			"test",
		},
	}

	data, err := xml.Marshal(&resp)
	ass.NoError(err)
	ass.Equal("<methodResponse><params><param><value><string>test</string></value></param></params></methodResponse>", string(data))
}

func TestParseResponse(t *testing.T) {
	ass := assert.New(t)

	_, err := ParseResponse(
		strings.NewReader("<<"))
	ass.EqualError(err, "XML syntax error on line 1: expected element name after <")

	_, err = ParseResponse(
		strings.NewReader("<methodResponse><params><param><value><invalid>test</invalid></param></params></methodResponse>"))
	ass.EqualError(err, "invalid value type invalid")

	_, err = ParseResponse(
		strings.NewReader("<methodResponse><fault><value><invalid>test</invalid></fault></methodResponse>"))
	ass.EqualError(err, "invalid value type invalid")

	_, err = ParseResponse(
		strings.NewReader("<methodResponse><fault><value>test</fault></methodResponse>"))
	ass.EqualError(err, "invalid fault value")

	resp, err := ParseResponse(
		strings.NewReader("<methodResponse><params><param><value>test</value></param></params><fault><value><struct><member><name>faultCode</name><value><int>4</int></value></member><member><name>faultString</name><value>faultString</value></member></struct></value></fault></methodResponse>"))
	ass.NoError(err)
	ass.Len(resp.Params, 1)
	ass.Equal("test", resp.Params[0])
	ass.Equal(int32(4), resp.Fault.Code)
	ass.Equal("faultString", resp.Fault.String)
}

var valueTestData = []struct {
	name  string
	xml   string
	value interface{}
	err   error
}{{
	"invalid_type",
	"<invalid>test</invalid>",
	nil,
	errors.New("invalid value type invalid"),
}, {
	"string",
	"<string>test</string>",
	"test",
	nil,
}, {
	"string_raw",
	"test",
	"test",
	nil,
}, {
	"int",
	"<int>42</int>",
	int32(42),
	nil,
}, {
	"i4",
	"<i4>42</i4>",
	int32(42),
	nil,
}, {
	"bool_true",
	"<boolean>1</boolean>",
	true,
	nil,
}, {
	"bool_false",
	"<boolean>0</boolean>",
	false,
	nil,
}, {
	"double",
	"<double>1.2</double>",
	float64(1.2),
	nil,
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
	nil,
}, {
	"array",
	`<array>
  <data>
    <value><invalid>111</invalid></value>
  </data>
</array>`,
	nil,
	errors.New("invalid value type invalid"),
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
	nil,
}, {
	"struct",
	`<struct>
  <member>
  </member>
</struct>`,
	nil,
	errors.New("missing struct name element"),
}, {
	"struct",
	`<struct>
  <member>
    <name>aaa</name>
  </member>
</struct>`,
	nil,
	errors.New("missing struct value element"),
}, {
	"struct",
	`<struct>
  <member>
    <name>aaa</name>
    <value><invalid>111</invalid></value>
  </member>
</struct>`,
	nil,
	errors.New("invalid value type invalid"),
}}

func TestParseValue(t *testing.T) {
	for _, d := range valueTestData {
		t.Run(d.name, func(st *testing.T) {
			ass := assert.New(st)

			doc := etree.NewDocument()
			ass.NoError(doc.ReadFromString(
				fmt.Sprintf("<value>%s</value>", d.xml)))

			value, err := parseValue(doc.SelectElement("value"))
			ass.Equal(d.err, err)
			ass.Equal(d.value, value)
		})
	}
}

var benchmarkData = `<methodResponse>
  <params>
    <param>
      <value>
        <string>test</string>
      </value>
    </param>
    <param>
      <value>
        test
      </value>
    </param>
    <param>
      <value>
        <int>42</int>
      </value>
    </param>
    <param>
      <value>
        <i4>42</i4>
      </value>
    </param>
    <param>
      <value>
        <boolean>1</boolean>
      </value>
    </param>
    <param>
      <value>
        <boolean>0</boolean>
      </value>
    </param>
    <param>
      <value>
        <double>1.2</double>
      </value>
    </param>
    <param>
      <value>
        <array>
		  <data>
			<value><i4>111</i4></value>
			<value><i4>222</i4></value>
		  </data>
		</array>
      </value>
    </param>
    <param>
      <value>
        <struct>
		  <member>
			<name>aaa</name>
			<value><i4>111</i4></value>
		  </member>
		  <member>
			<name>bbb</name>
			<value><i4>222</i4></value>
		  </member>
		</struct>
      </value>
    </param>
  </params>
  <fault>
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
  </fault>
</methodResponse>`

func BenchmarkParseResponse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		response, err := ParseResponse(strings.NewReader(benchmarkData))
		if err != nil {
			b.Fatal(err)
		}
		_ = response
	}
}
