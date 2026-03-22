package strings_test

import (
	"strings"
	"testing"
)

func TestExternal(t *testing.T) {
	s := "hello world"
	if !strings.Contains(s, "world") {
		t.Error("expected hello world to contain world")
	}
}
