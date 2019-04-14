package rpc

import (
	"encoding/xml"

	"github.com/spf13/cast"
)

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

// UnmarshalXML convert XML to Response
func (r *Response) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var data struct {
		Params []param `xml:"params>param"`
		Fault  *Fault  `xml:"fault"`
	}

	if err := d.DecodeElement(&data, &start); err != nil {
		return err
	}

	r.Fault = data.Fault
	if data.Params != nil {
		r.Params = make([]interface{}, len(data.Params))
		for i, v := range data.Params {
			r.Params[i] = v.Value.Interface()
		}
	}

	return nil
}

// Fault information of response
type Fault struct {
	Code   int32
	String string
}

// UnmarshalXML convert XML to Fault
func (f *Fault) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var data param
	if err := d.DecodeElement(&data, &start); err != nil {
		return err
	}
	dat := data.Value.Struct.Interface()

	f.Code = cast.ToInt32(dat["faultCode"])
	f.String = cast.ToString(dat["faultString"])
	return nil
}

// param helper to parse parameters
type param struct {
	Value value `xml:"value"`
}

// value parses any value type
type value struct {
	Content string      `xml:",innerxml"`
	Array   *array      `xml:"array"`
	Struct  *structData `xml:"struct"`
	Int     *integer    `xml:"int"`
	I4      *integer    `xml:"i4"`
	Boolean *boolean    `xml:"boolean"`
	Double  *double     `xml:"double"`
	String  *str        `xml:"string"`
}

// Interface returns value as go type
func (v *value) Interface() interface{} {
	if v.Array != nil {
		return v.Array.Interface()
	}
	if v.Struct != nil {
		return v.Struct.Interface()
	}
	if v.Int != nil {
		return v.Int.Interface()
	}
	if v.I4 != nil {
		return v.I4.Interface()
	}
	if v.Boolean != nil {
		return v.Boolean.Interface()
	}
	if v.Double != nil {
		return v.Double.Interface()
	}
	if v.String != nil {
		return v.String.Interface()
	}
	return v.Content
}

// str handles string data types
type str struct {
	Content string `xml:",innerxml"`
}

// Interface returns data as go type
func (s *str) Interface() interface{} {
	return cast.ToString(s.Content)
}

// double handles double data types
type double struct {
	Content string `xml:",innerxml"`
}

// Interface returns data as go type
func (d *double) Interface() interface{} {
	return cast.ToFloat64(d.Content)
}

// boolean handles boolean data types
type boolean struct {
	Content string `xml:",innerxml"`
}

// Interface returns data as go type
func (b *boolean) Interface() interface{} {
	return cast.ToBool(b.Content)
}

// integer handles integer data types
type integer struct {
	Content string `xml:",innerxml"`
}

// Interface returns data as go type
func (i *integer) Interface() interface{} {
	return cast.ToInt32(i.Content)
}

// array handles array data types
type array struct {
	Data []value `xml:"data>value"`
}

// Interface returns data as go type
func (a *array) Interface() interface{} {
	data := make([]interface{}, len(a.Data))
	for i, v := range a.Data {
		data[i] = v.Interface()
	}
	return data
}

// structData handles struct data types
type structData struct {
	Member []struct {
		Name  string `xml:"name"`
		Value value  `xml:"value"`
	} `xml:"member"`
}

// Interface returns data as go type
func (s *structData) Interface() map[string]interface{} {
	data := make(map[string]interface{}, len(s.Member))
	for _, m := range s.Member {
		data[m.Name] = m.Value.Interface()
	}
	return data
}
