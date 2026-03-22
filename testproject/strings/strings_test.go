package strings

import "testing"

func TestReverse(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "olleh"},
		{"world", "dlrow"},
		{"", ""},
		{"a", "a"},
		{"ab", "ba"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := Reverse(tt.input)
			if result != tt.expected {
				t.Errorf("Reverse(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestToUpper(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "HELLO"},
		{"World", "WORLD"},
		{"", ""},
		{"ALREADY UPPER", "ALREADY UPPER"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ToUpper(tt.input)
			if result != tt.expected {
				t.Errorf("ToUpper(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestContains(t *testing.T) {
	if !Contains("hello world", "world") {
		t.Error("Contains(hello world, world) = false; want true")
	}

	if Contains("hello", "xyz") {
		t.Error("Contains(hello, xyz) = true; want false")
	}
}

func TestContainsFailing(t *testing.T) {
	if Contains("hello", "ll") != true {
		t.Error("Contains(hello, ll) should be true")
	}
}
