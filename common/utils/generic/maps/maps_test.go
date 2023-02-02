package maps

import (
	"reflect"
	"testing"
)

func TestMerge(t *testing.T) {
	ret := map[string]int{"a": 1}
	other := map[string]int{"a": 1, "b": 2}
	got := map[string]int{"a": 2, "b": 2}
	Merge(ret, other)
	t.Logf("Merge(%v, %v) == %v", map[string]int{"a": 1}, other, got)
	if !reflect.DeepEqual(ret, got) {
		t.Errorf("Merge(%v, %v) != %v", map[string]int{"a": 1}, other, got)
	}
}
