package math

import "testing"

func TestAdd(t *testing.T) {
	tests := []struct {
		name     string
		a, b     int
		expected int
	}{
		{"positive numbers", 2, 3, 5},
		{"negative numbers", -1, -1, -2},
		{"mixed numbers", -5, 3, -2},
		{"zero", 0, 5, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Add(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Add(%d, %d) = %d; want %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestSubtract(t *testing.T) {
	result := Subtract(10, 3)
	if result != 7 {
		t.Errorf("Subtract(10, 3) = %d; want 7", result)
	}
}

func TestMultiply(t *testing.T) {
	result := Multiply(4, 5)
	if result != 20 {
		t.Errorf("Multiply(4, 5) = %d; want 20", result)
	}
}

func TestDivide(t *testing.T) {
	tests := []struct {
		name     string
		a, b     int
		expected int
	}{
		{"normal division", 10, 2, 5},
		{"division by 1", 7, 1, 7},
		{"division by zero", 5, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Divide(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Divide(%d, %d) = %d; want %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestAddFailing(t *testing.T) {
	result := Add(2, 2)
	if result != 5 {
		t.Errorf("Add(2, 2) = %d; want 5", result)
	}
}

func TestSkipped(t *testing.T) {
	t.Skip("Skipping this test for now")
	result := Add(1, 1)
	if result != 2 {
		t.Errorf("Add(1, 1) = %d; want 2", result)
	}
}
