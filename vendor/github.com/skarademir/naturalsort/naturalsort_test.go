package naturalsort

import (
	"reflect"
	"sort"
	"testing"
)

func TestSortValid(t *testing.T) {
	cases := []struct {
		data, expected []string
	}{
		{
			nil,
			nil,
		},
		{
			[]string{},
			[]string{},
		},
		{
			[]string{"a"},
			[]string{"a"},
		},
		{
			[]string{"0"},
			[]string{"0"},
		},
		{
			[]string{"data", "data20", "data3"},
			[]string{"data", "data3", "data20"},
		},
		{
			[]string{"1", "2", "30", "22", "0", "00", "3"},
			[]string{"0", "00", "1", "2", "3", "22", "30"},
		},
		{
			[]string{"A1", "A0", "A21", "A11", "A111", "A2"},
			[]string{"A0", "A1", "A2", "A11", "A21", "A111"},
		},
		{
			[]string{"A1BA1", "A11AA1", "A2AB0", "B1AA1", "A1AA1"},
			[]string{"A1AA1", "A1BA1", "A2AB0", "A11AA1", "B1AA1"},
		},
		{
			[]string{"1ax10", "1a10", "1ax2", "1ax"},
			[]string{"1a10", "1ax", "1ax2", "1ax10"},
		},
		{
			[]string{"z1a10", "z1ax2", "z1ax"},
			[]string{"z1a10", "z1ax", "z1ax2"},
		},
		{
			// regression test for #8
			[]string{"a0000001", "a0001"},
			[]string{"a0001", "a0000001"},
		},
		{
			// regression test for #10 - Number sort before any symbols even if theyre lower on the ASCII table
			[]string{"#1", "1", "_1", "a"},
			[]string{"1", "#1", "_1", "a"},
		},
		{
			// regression test for #10 - Number sort before any symbols even if theyre lower on the ASCII table
			[]string{"#1", "1", "_1", "a"},
			[]string{"1", "#1", "_1", "a"},
		},
		{	// test correct handling of space-only strings 
			[]string{"1", " ", "0"},
			[]string{"0", "1", " "},
		},
		{  // test correct handling of multiple spaces being correctly ordered AFTER numbers 
			[]string{"1", " ", " 1", "  "},
			[]string{"1", " ", " 1", "  "},
		},
		{
			[]string{"1", "#1", "a#", "a1"},
			[]string{"1", "#1", "a1", "a#"},
		},
		{
			// regression test for #10
			[]string{"111111111111111111112", "111111111111111111113", "1111111111111111111120"},
			[]string{"111111111111111111112", "111111111111111111113", "1111111111111111111120"},
		},
	}

	for i, c := range cases {
		sort.Sort(NaturalSort(c.data))
		if !reflect.DeepEqual(c.data, c.expected) {
			t.Fatalf("Wrong order in test case #%d.\nExpected=%v\nGot=%v", i, c.expected, c.data)
		}
	}

}

func BenchmarkSort(b *testing.B) {
	var data = [...]string{"A1BA1", "A11AA1", "A2AB0", "B1AA1", "A1AA1"}
	for ii := 0; ii < b.N; ii++ {
		d := NaturalSort(data[:])
		sort.Sort(d)
	}
}
