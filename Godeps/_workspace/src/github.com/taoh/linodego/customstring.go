package linodego

import ()

// A special class to handle marshaling string response from Linode
// As sometimes Linode returns integer instead of string from API
// https://github.com/taoh/linodego/issues/1
type CustomString struct {
	string
}

func (cs *CustomString) UnmarshalJSON(b []byte) (err error) {
	// if bytes are not surrounded by double quotes
	if b[0] != '"' && b[len(b)-1] != '"' {
		cs.string = string(append(append([]byte{'"'}, b...), '"'))
	} else {
		cs.string = string(b)
	}
	return nil
}

func (cs *CustomString) MarshalJSON() ([]byte, error) {
	return []byte(cs.string), nil
}

func (cs *CustomString) String() string {
	// remove double quotes
	return cs.string[1 : len(cs.string)-1]
}
