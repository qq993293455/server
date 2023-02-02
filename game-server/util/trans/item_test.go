package trans

import (
	"reflect"
	"testing"

	"coin-server/common/values"
)

func assertPanic(t *testing.T, f func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	f()
}

func TestItemSliceToMap(t *testing.T) {
	assertPanic(t, func() {
		ItemSliceToMap([]values.Integer{1})
		ItemSliceToMap([]values.Integer{1, 2, 3})
	})

	var got map[values.Integer]values.Integer
	expect := map[values.Integer]values.Integer{101: 1}
	got = ItemSliceToMap([]values.Integer{101, 1})
	if !reflect.DeepEqual(got, expect) {
		t.Errorf("expect %v got %v", expect, got)
	}

	expect = map[values.Integer]values.Integer{101: 3}
	got = ItemSliceToMap([]values.Integer{101, 1, 101, 2})
	if !reflect.DeepEqual(got, expect) {
		t.Errorf("expect %v got %v", expect, got)
	}

	expect = map[values.Integer]values.Integer{101: 3, 102: 2}
	got = ItemSliceToMap([]values.Integer{101, 1, 101, 2, 102, 2})
	if !reflect.DeepEqual(got, expect) {
		t.Errorf("expect %v got %v", expect, got)
	}
}
