package script

import (
	"encoding/xml"
	"io"
	"strings"
)

// Result of a script execution
type Result map[string]string

type xmlMapEntry struct {
	XMLName xml.Name
	Value   string `xml:",chardata"`
}

// UnmarshalXML to result nap
func (s *Result) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	*s = make(map[string]string)
	for {
		var e xmlMapEntry

		err := d.Decode(&e)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		(*s)[e.XMLName.Local] = e.Value
	}
	return nil
}

// GetMap from result
func (s *Result) GetMap(key string) map[string]string {
	entry, ok := (*s)[key]
	if !ok {
		return nil
	}

	lines := strings.Split(entry, "\n")
	data := make(map[string]string, len(lines))
	for _, line := range lines {
		lineSplit := strings.SplitN(line, "=", 2)
		if len(lineSplit) > 1 {
			data[lineSplit[0]] = lineSplit[1]
		}
	}
	return data
}
