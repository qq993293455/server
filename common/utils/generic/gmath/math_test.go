package gmath

import "testing"

func TestCeilTo(t *testing.T) {
	expect := int64(21)
	got := CeilTo[int64](20.2)
	if got != expect {
		t.Errorf("expect %d, got %d", expect, got)
	}
}
