package calc

import "testing"

func TestAdd(t *testing.T) {
	if got := Add(2, 3); got != 5 {
		t.Fatalf("Add(2, 3) = %d, want 5", got)
	}
}

func TestDivide(t *testing.T) {
	if got := Divide(8, 2); got != 4 {
		t.Fatalf("Divide(8, 2) = %d, want 4", got)
	}
}
