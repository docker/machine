package naturalsort

import (
	"regexp"
	"strconv"
	"strings"
)

type NaturalSort []string

func (s NaturalSort) Len() int {
	return len(s)
}
func (s NaturalSort) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s NaturalSort) Less(i, j int) bool {
	r1 := regexp.MustCompilePOSIX(`^([^0-9]*)+|[0-9]+`)

	spliti := r1.FindAllString(strings.Replace(s[i], " ", "", -1), -1)
	splitj := r1.FindAllString(strings.Replace(s[j], " ", "", -1), -1)
	for index := range spliti {
		if spliti[index] != splitj[index] {
			inti, ei := strconv.Atoi(spliti[index])
			intj, ej := strconv.Atoi(splitj[index])
			if ei == nil && ej == nil { //if number
				return inti < intj
			}
			return spliti[index] < splitj[index]
		}

	}
	return s[i] < s[j]
}
