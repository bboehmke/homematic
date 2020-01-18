package rpc

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/beevik/etree"
	"github.com/spf13/cast"
	"golang.org/x/net/html/charset"
)

// Fault information of response
type Fault struct {
	Code   int32
	String string
}

// toMap returns data for fault entry
func (f *Fault) toMap() map[string]interface{} {
	return map[string]interface{}{
		"faultCode":   f.Code,
		"faultString": f.String,
	}
}

// Response of XML RPCs
type Response struct {
	Params []interface{}
	Fault  *Fault
}

// FirstParam returns the first parameter or nil
func (r *Response) FirstParam() interface{} {
	if len(r.Params) > 0 {
		return r.Params[0]
	}
	return nil
}

// MarshalXML convert response to XML
func (r *Response) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	methodResponse := xml.Name{Local: "methodResponse"}

	err := e.EncodeToken(xml.StartElement{Name: methodResponse})
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

	if r.Fault != nil {
		fault := xml.Name{Local: "fault"}
		err = e.EncodeToken(xml.StartElement{Name: fault})
		if err != nil {
			return err
		}

		err = encodeValue(map[string]interface{}{
			"faultCode":   r.Fault.Code,
			"faultString": r.Fault.String,
		}, e)
		if err != nil {
			return err
		}

		err = e.EncodeToken(xml.EndElement{Name: fault})
		if err != nil {
			return err
		}
	}

	return e.EncodeToken(xml.EndElement{Name: methodResponse})
}

// ParseResponse from XML
func ParseResponse(reader io.Reader) (*Response, error) {
	doc := etree.NewDocument()
	doc.ReadSettings.CharsetReader = charset.NewReaderLabel
	_, err := doc.ReadFrom(reader)
	if err != nil {
		return nil, err
	}

	// handle parameters
	elements := doc.FindElements("/methodResponse/params/param/value")
	response := &Response{
		Params: make([]interface{}, len(elements)),
	}

	for idx, element := range elements {
		response.Params[idx], err = parseValue(element)
		if err != nil {
			return nil, err
		}
	}

	// handle faults
	faultElement := doc.FindElement("/methodResponse/fault/value")
	if faultElement != nil {
		value, err := parseValue(faultElement)
		if err != nil {
			return nil, err
		}
		data, ok := value.(map[string]interface{})
		if !ok {
			return nil, errors.New("invalid fault value")
		}
		response.Fault = &Fault{
			Code:   cast.ToInt32(data["faultCode"]),
			String: cast.ToString(data["faultString"]),
		}
	}

	return response, nil
}

// parseValue element to go interface
func parseValue(element *etree.Element) (interface{}, error) {
	// use text of element if child elements are missing
	children := element.ChildElements()
	if len(children) == 0 {
		return strings.TrimSpace(element.Text()), nil
	}
	e := children[0]

	switch e.Tag {
	case "string":
		return strings.TrimSpace(e.Text()), nil

	case "int", "i4":
		return cast.ToInt32E(strings.TrimSpace(e.Text()))

	case "boolean":
		return cast.ToBoolE(strings.TrimSpace(e.Text()))

	case "double":
		return cast.ToFloat64E(strings.TrimSpace(e.Text()))

	case "array":
		elements := e.FindElements("./data/value")
		values := make([]interface{}, len(elements))
		var err error
		for idx, valueElement := range elements {
			values[idx], err = parseValue(valueElement)
			if err != nil {
				return nil, err
			}
		}
		return values, nil

	case "struct":
		elements := e.FindElements("./member")
		values := make(map[string]interface{}, len(elements))
		var err error
		for _, memberElement := range elements {
			nameElement := memberElement.SelectElement("name")
			if nameElement == nil {
				return nil, errors.New("missing struct name element")
			}
			valueElement := memberElement.SelectElement("value")
			if valueElement == nil {
				return nil, errors.New("missing struct value element")
			}

			values[nameElement.Text()], err = parseValue(valueElement)
			if err != nil {
				return nil, err
			}
		}
		return values, nil

	default:
		return nil, fmt.Errorf("invalid value type %s", e.Tag)
	}
}
