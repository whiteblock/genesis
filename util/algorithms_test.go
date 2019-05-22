package util

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"
)

func TestGetUniqueStrings(t *testing.T) {
	var test = []struct {
		in       []string
		expected []string
	}{
		{[]string{"0", "4", "4", "7", "9", "3", "8", "0"}, []string{"0", "4", "7", "9", "3", "8"}},
		{[]string{"3", "3", "2"}, []string{"3", "2"}},
		{[]string{"1", "1", "1"}, []string{"1"}},
		{[]string{"get", "test", "go", "test"}, []string{"get", "test", "go"}},
	}

	for i, tt := range test {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			fmt.Println(GetUniqueStrings(tt.in))
			if !reflect.DeepEqual(GetUniqueStrings(tt.in), tt.expected) {
				t.Errorf("return value from GetUniqueStrings does not match expected value")
			}
		})
	}
}
