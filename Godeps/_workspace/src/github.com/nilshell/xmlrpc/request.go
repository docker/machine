package xmlrpc

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"
)

func newRequest(url string, method string, params ...interface{}) (*http.Request, error) {
	body := buildRequestBody(method, params)
	request, err := http.NewRequest("POST", url, strings.NewReader(body))

	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "text/xml")
	request.Header.Set("Content-Length", fmt.Sprintf("%d", len(body)))

	return request, nil
}

func buildRequestBody(method string, params []interface{}) (buffer string) {
	buffer += `<?xml version="1.0" encoding="UTF-8"?><methodCall>`
	buffer += fmt.Sprintf("<methodName>%s</methodName><params>", method)

	if params != nil && len(params) > 0 {
		for _, value := range params {
			if value != nil {
				switch ps := value.(type) {
				case Params:
					for _, p := range ps.Params {
						if p != nil {
							buffer += buildParamElement(p)
						}
					}
				default:
					buffer += buildParamElement(ps)
				}
			}
		}
	}

	buffer += "</params></methodCall>"

	return
}

func buildParamElement(value interface{}) string {
	return fmt.Sprintf("<param>%s</param>", buildValueElement(value))
}

func buildValueElement(value interface{}) (buffer string) {
	buffer = `<value>`

	switch v := value.(type) {
	case Struct:
		buffer += buildStructElement(v)
	case Base64:
		escaped := escapeString(string(v))
		buffer += fmt.Sprintf("<base64>%s</base64>", escaped)
	case string:
		escaped := escapeString(value.(string))
		buffer += fmt.Sprintf("<string>%s</string>", escaped)
	case int, int8, int16, int32, int64:
		buffer += fmt.Sprintf("<int>%d</int>", v)
	case float32, float64:
		buffer += fmt.Sprintf("<double>%f</double>", v)
	case bool:
		buffer += buildBooleanElement(v)
	case time.Time:
		buffer += buildTimeElement(v)
	default:
		rv := reflect.ValueOf(value)

		if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
			buffer += buildArrayElement(v)
		} else {
			fmt.Errorf("Unsupported value type")
		}
	}

	buffer += `</value>`

	return
}

func buildStructElement(param Struct) (buffer string) {
	buffer = `<struct>`

	for name, value := range param {
		buffer += fmt.Sprintf("<member><name>%s</name>", name)
		buffer += buildValueElement(value)
		buffer += `</member>`
	}

	buffer += `</struct>`

	return
}

func buildBooleanElement(value bool) (buffer string) {
	if value {
		buffer = `<boolean>1</boolean>`
	} else {
		buffer = `<boolean>0</boolean>`
	}

	return
}

func buildTimeElement(t time.Time) string {
	return fmt.Sprintf(
		"<dateTime.iso8601>%d%d%dT%d:%d:%d</dateTime.iso8601>",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second(),
	)
}

func buildArrayElement(array interface{}) string {
	buffer := `<array><data>`

	a := reflect.ValueOf(array)
	for i := 0; i < a.Len(); i++ {
		buffer += buildValueElement(a.Index(i).Interface())
	}

	buffer += `</data></array>`

	return buffer
}

func escapeString(s string) string {
	buffer := bytes.NewBuffer([]byte{})
	xml.Escape(buffer, []byte(s))

	return fmt.Sprintf("%v", buffer)
}
