package percent

import (
	"testing"
)

func TestAddition(t *testing.T) {
	expect := int64(10)
	got := Addition(100, 1000)
	if got != expect {
		t.Errorf("expect %d, got %d", expect, got)
	}
}

func TestAdditionFloat(t *testing.T) {
	expect := 0.2
	got := AdditionFloat(20, 100)
	if got != expect {
		t.Errorf("expect %f, got %f", expect, got)
	}
}
