package rpc

import (
	"encoding/xml"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequest_MarshalXML(t *testing.T) {
	ass := assert.New(t)

	request := Request{
		Method: "test",
		Params: []interface{}{
			"aaa",
		},
	}

	data, err := xml.MarshalIndent(request, "", "  ")
	ass.NoError(err)
	ass.Equal(`<methodCall>
  <methodName>test</methodName>
  <params>
    <param>
      <value>
        <string>aaa</string>
      </value>
    </param>
  </params>
</methodCall>`, string(data))
}

func TestParseRequest(t *testing.T) {
	ass := assert.New(t)

	_, err := ParseRequest(
		strings.NewReader("<<"))
	ass.EqualError(err, "XML syntax error on line 1: expected element name after <")

	_, err = ParseRequest(
		strings.NewReader("<methodCall><params><param><value><invalid>test</invalid></param></params></methodCall>"))
	ass.EqualError(err, "invalid value type invalid")

	_, err = ParseRequest(
		strings.NewReader("<methodCall><params><param><value>test</param></params></methodCall>"))
	ass.EqualError(err, "method name is missing")

	request, err := ParseRequest(
		strings.NewReader("<methodCall><methodName>func</methodName><params><param><value>test</param></params></methodCall>"))
	ass.NoError(err)
	ass.Equal("func", request.Method)
	ass.Len(request.Params, 1)
	ass.Equal("test", request.Params[0])
}

func TestRequest_encodeValue(t *testing.T) {
	ass := assert.New(t)

	request := Request{
		Method: "test",
		Params: []interface{}{
			"aaa",
			true,
			false,
			int(1),
			int8(2),
			int16(3),
			int32(4),
			int64(5),
			uint(11),
			uint8(12),
			uint16(13),
			uint32(14),
			uint64(15),
			float32(1.1),
			float64(2.2),
			[]interface{}{
				111,
				222,
			},
			map[string]interface{}{
				"aaa": 111,
			},
		},
	}

	data, err := xml.MarshalIndent(request, "", "  ")
	ass.NoError(err)
	ass.Equal(`<methodCall>
  <methodName>test</methodName>
  <params>
    <param>
      <value>
        <string>aaa</string>
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
        <int>1</int>
      </value>
    </param>
    <param>
      <value>
        <int>2</int>
      </value>
    </param>
    <param>
      <value>
        <int>3</int>
      </value>
    </param>
    <param>
      <value>
        <int>4</int>
      </value>
    </param>
    <param>
      <value>
        <int>5</int>
      </value>
    </param>
    <param>
      <value>
        <int>11</int>
      </value>
    </param>
    <param>
      <value>
        <int>12</int>
      </value>
    </param>
    <param>
      <value>
        <int>13</int>
      </value>
    </param>
    <param>
      <value>
        <int>14</int>
      </value>
    </param>
    <param>
      <value>
        <int>15</int>
      </value>
    </param>
    <param>
      <value>
        <double>1.1</double>
      </value>
    </param>
    <param>
      <value>
        <double>2.2</double>
      </value>
    </param>
    <param>
      <value>
        <array>
          <data>
            <value>
              <int>111</int>
            </value>
            <value>
              <int>222</int>
            </value>
          </data>
        </array>
      </value>
    </param>
    <param>
      <value>
        <struct>
          <member>
            <name>aaa</name>
            <value>
              <int>111</int>
            </value>
          </member>
        </struct>
      </value>
    </param>
  </params>
</methodCall>`, string(data))

	request = Request{
		Method: "test",
		Params: []interface{}{
			struct {
				AAA string
			}{
				AAA: "aaa",
			},
		},
	}

	_, err = xml.MarshalIndent(request, "", "  ")
	ass.EqualError(err, "unknown value type")
}
