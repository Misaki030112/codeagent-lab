package calc

import "testing"

func TestAdd(t *testing.T) {
	if got := Add(2, 3); got != 5 {
		t.Fatalf("Add(2, 3) = %d, want 5", got)
	}
}

func TestDivide(t *testing.T) {
	got, err := Divide(8, 2)
	if err != nil {
		t.Fatalf("Divide(8, 2) unexpected error: %v", err)
	}
	if got != 4 {
		t.Fatalf("Divide(8, 2) = %d, want 4", got)
	}
}

func TestDivideByZero(t *testing.T) {
	_, err := Divide(10, 0)
	if err == nil {
		t.Fatal("Divide(10, 0) expected error, got nil")
	}
	if err.Error() != "division by zero" {
		t.Fatalf("Divide(10, 0) error = %q, want %q", err.Error(), "division by zero")
	}
}
