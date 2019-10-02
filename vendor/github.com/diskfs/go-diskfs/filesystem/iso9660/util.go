package iso9660

import (
	"strings"
)

const (
	// KB represents one KB
	KB int64 = 1024
	// MB represents one MB
	MB int64 = 1024 * KB
	// GB represents one GB
	GB int64 = 1024 * MB
	// TB represents one TB
	TB int64 = 1024 * GB
)

func universalizePath(p string) (string, error) {
	// globalize the separator
	ps := strings.Replace(p, `\`, "/", -1)
	//if ps[0] != '/' {
	//return "", errors.New("Must use absolute paths")
	//}
	return ps, nil
}
func splitPath(p string) ([]string, error) {
	ps, err := universalizePath(p)
	if err != nil {
		return nil, err
	}
	// we need to split such that each one ends in "/", except possibly the last one
	parts := strings.Split(ps, "/")
	// eliminate empty parts
	ret := make([]string, 0)
	for _, sub := range parts {
		if sub != "" {
			ret = append(ret, sub)
		}
	}
	return ret, nil
}

func ucs2StringToBytes(s string) []byte {
	rs := []rune(s)
	l := len(rs)
	b := make([]byte, 0, 2*l)
	// big endian
	for _, r := range rs {
		tmpb := []byte{byte(r >> 8), byte(r & 0x00ff)}
		b = append(b, tmpb...)
	}
	return b
}

func bytesToUCS2String(b []byte) string {
	r := make([]rune, 0, 30)
	// now we can iterate
	for i := 0; i < len(b); {
		// little endian
		val := uint16(b[i])<<8 + uint16(b[i+1])
		r = append(r, rune(val))
		i += 2
	}
	return string(r)
}

// maxInt returns the larger of x or y.
func maxInt(x, y int) int {
	if x < y {
		return y
	}
	return x
}
