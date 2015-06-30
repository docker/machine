package xmlrpc

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// TIME_LAYOUT defines time template defined by iso8601, used to encode/decode time values.
const TIME_LAYOUT = "20060102T15:04:05"

func parseValue(valueXml []byte) (result interface{}, err error) {
	parser := xml.NewDecoder(bytes.NewReader(valueXml))
	result, err = getValue(parser)

	return
}

func getValue(parser *xml.Decoder) (result interface{}, err error) {
	var token xml.Token

	token, err = parser.Token()

	if err != nil {
		return nil, err
	}

	for {
		switch t := token.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "boolean":
				return getBooleanValue(parser)
			case "dateTime.iso8601":
				return getDateValue(parser)
			case "double":
				return getDoubleValue(parser)
			case "int", "i4", "i8":
				return getIntValue(parser)
			case "base64":
				return getBase64Value(parser)
			case "string":
				return getStringValue(parser)
			case "struct":
				result, err = getStructValue(parser)
			case "array":
				result, err = getArrayValue(parser)
			default:
				// Move on
			}
		case xml.EndElement:
			if t.Name.Local == "value" {
				return result, nil
			}
		case xml.CharData:
			cdata := strings.TrimSpace(string(t))
			if cdata != "" {
				result = cdata
			}
		}

		token, err = parser.Token()

		if err != nil {
			return nil, err
		}
	}

	return
}

func getBooleanValue(parser *xml.Decoder) (result interface{}, err error) {
	var value string
	value, err = getElementValue(parser)

	switch value {
	case "0":
		return false, nil
	case "1":
		return true, nil
	}

	return nil, fmt.Errorf("Parse error: invalid boolean value (%s).", value)
}

func getDateValue(parser *xml.Decoder) (result interface{}, err error) {
	var value string
	value, err = getElementValue(parser)
	result, err = time.Parse(TIME_LAYOUT, value)

	return
}

func getDoubleValue(parser *xml.Decoder) (interface{}, error) {
	value, _ := getElementValue(parser)
	return strconv.ParseFloat(value, 64)
}

func getIntValue(parser *xml.Decoder) (interface{}, error) {
	value, _ := getElementValue(parser)
	var number int64
	number, err := strconv.ParseInt(value, 0, 64)

	return number, err
}

func getBase64Value(parser *xml.Decoder) (string, error) {
	return getElementValue(parser)
}

func getStringValue(parser *xml.Decoder) (string, error) {
	return getElementValue(parser)
}

func getStructValue(parser *xml.Decoder) (result interface{}, err error) {
	var token xml.Token
	token, err = parser.Token()

	result = Struct{}

	for {
		switch t := token.(type) {
		case xml.StartElement:
			member := getStructMember(parser)
			result.(Struct)[member["name"].(string)] = member["value"]
		case xml.EndElement:
			if t.Name.Local == "struct" {
				return result, err
			}
		}

		token, err = parser.Token()
	}

	return
}

func getStructMember(parser *xml.Decoder) (member Struct) {
	var token xml.Token
	token, _ = parser.Token()

	member = Struct{}

	for {
		switch t := token.(type) {
		case xml.StartElement:
			if t.Name.Local == "name" {
				member["name"], _ = getElementValue(parser)
			}

			if t.Name.Local == "value" {
				member["value"], _ = getValue(parser)
			}
		case xml.EndElement:
			if t.Name.Local == "member" {
				return member
			}
		}

		token, _ = parser.Token()
	}

	return
}

func getElementValue(parser *xml.Decoder) (value string, err error) {
	var token xml.Token
	token, err = parser.Token()

	processing := true

	for processing {
		switch token.(type) {
		case xml.CharData:
			value = strings.TrimSpace(string(token.(xml.CharData)))
			if value != "" {
				processing = false
			}
		case xml.EndElement:
			processing = false
		}
		token, err = parser.Token()
	}

	return
}

func getArrayValue(parser *xml.Decoder) (result interface{}, err error) {
	var token xml.Token
	token, err = parser.Token()

	result = []interface{}{}

	for {
		switch t := token.(type) {
		case xml.StartElement:
			if t.Name.Local == "value" {
				var value interface{}
				value, err = getValue(parser)

				result = append(result.([]interface{}), value)
			}
		case xml.EndElement:
			if t.Name.Local == "array" {
				return result, err
			}
		}

		token, err = parser.Token()
	}

	return
}
