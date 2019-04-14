package rpc

import (
	"encoding/xml"
	"errors"
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

// encodeValue convert value to XML data
func encodeValue(value interface{}, e *xml.Encoder) error {
	valueName := xml.Name{Local: "value"}

	err := e.EncodeToken(xml.StartElement{Name: valueName})
	if err != nil {
		return err
	}

	switch v := value.(type) {
	case string:
		err = e.Encode(v)
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

	default:
		return errors.New("unknown value type")
	}

	return e.EncodeToken(xml.EndElement{Name: valueName})
}
