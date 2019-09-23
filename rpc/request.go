package rpc

import (
	"encoding/xml"
	"errors"
	"io"
	"strings"

	"github.com/beevik/etree"
	"golang.org/x/net/html/charset"
)

// Request for XML RPCs
type Request struct {
	Method string
	Params []interface{}
}

// MarshalXML convert request to XML
func (r Request) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	methodCall := xml.Name{Local: "methodCall"}

	err := e.EncodeToken(xml.StartElement{Name: methodCall})
	if err != nil {
		return err
	}

	err = e.EncodeElement(r.Method, xml.StartElement{
		Name: xml.Name{Local: "methodName"},
	})
	if err != nil {
		return err
	}

	if len(r.Params) > 0 {
		params := xml.Name{Local: "params"}
		err = e.EncodeToken(xml.StartElement{Name: params})
		if err != nil {
			return err
		}

		param := xml.Name{Local: "param"}
		for _, p := range r.Params {
			err = e.EncodeToken(xml.StartElement{Name: param})
			if err != nil {
				return err
			}
			err = encodeValue(p, e)
			if err != nil {
				return err
			}
			err = e.EncodeToken(xml.EndElement{Name: param})
			if err != nil {
				return err
			}
		}

		err = e.EncodeToken(xml.EndElement{Name: params})
		if err != nil {
			return err
		}
	}
	return e.EncodeToken(xml.EndElement{Name: methodCall})
}

// ParseRequest from XML
func ParseRequest(reader io.Reader) (*Request, error) {
	doc := etree.NewDocument()
	doc.ReadSettings.CharsetReader = charset.NewReaderLabel
	_, err := doc.ReadFrom(reader)
	if err != nil {
		return nil, err
	}

	// handle parameters
	elements := doc.FindElements("/methodCall/params/param/value")
	request := &Request{
		Params: make([]interface{}, len(elements)),
	}

	for idx, element := range elements {
		request.Params[idx], err = parseValue(element)
		if err != nil {
			return nil, err
		}
	}

	// handle name
	nameElement := doc.FindElement("/methodCall/methodName")
	if nameElement == nil {
		return nil, errors.New("method name is missing")
	}
	request.Method = strings.TrimSpace(nameElement.Text())

	return request, nil
}

// encodeValue convert value to XML data
func encodeValue(value interface{}, e *xml.Encoder) error {
	valueName := xml.Name{Local: "value"}

	err := e.EncodeToken(xml.StartElement{Name: valueName})
	if err != nil {
		return err
	}

	switch v := value.(type) {
	case string:
		err = e.EncodeElement(v, xml.StartElement{
			Name: xml.Name{Local: "string"},
		})
		if err != nil {
			return err
		}

	case bool:
		var i int
		if v {
			i = 1
		}
		err = e.EncodeElement(i, xml.StartElement{
			Name: xml.Name{Local: "boolean"},
		})
		if err != nil {
			return err
		}
	case float32, float64:
		err = e.EncodeElement(v, xml.StartElement{
			Name: xml.Name{Local: "double"},
		})
		if err != nil {
			return err
		}
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		err = e.EncodeElement(v, xml.StartElement{
			Name: xml.Name{Local: "int"},
		})
		if err != nil {
			return err
		}
	case []interface{}:
		arrayName := xml.Name{Local: "array"}
		err = e.EncodeToken(xml.StartElement{Name: arrayName})
		if err != nil {
			return err
		}
		dataName := xml.Name{Local: "data"}
		err = e.EncodeToken(xml.StartElement{Name: dataName})
		if err != nil {
			return err
		}

		for _, entry := range v {
			err = encodeValue(entry, e)
			if err != nil {
				return err
			}
		}

		err = e.EncodeToken(xml.EndElement{Name: dataName})
		if err != nil {
			return err
		}
		err = e.EncodeToken(xml.EndElement{Name: arrayName})
		if err != nil {
			return err
		}
	case map[string]interface{}:
		structName := xml.Name{Local: "struct"}
		err = e.EncodeToken(xml.StartElement{Name: structName})
		if err != nil {
			return err
		}

		for key, entryValue := range v {
			memberName := xml.Name{Local: "member"}
			err = e.EncodeToken(xml.StartElement{Name: memberName})
			if err != nil {
				return err
			}

			err = e.EncodeElement(key, xml.StartElement{
				Name: xml.Name{Local: "name"},
			})
			if err != nil {
				return err
			}

			err = encodeValue(entryValue, e)
			if err != nil {
				return err
			}

			err = e.EncodeToken(xml.EndElement{Name: memberName})
			if err != nil {
				return err
			}
		}

		err = e.EncodeToken(xml.EndElement{Name: structName})
		if err != nil {
			return err
		}

	default:
		return errors.New("unknown value type")
	}

	return e.EncodeToken(xml.EndElement{Name: valueName})
}
