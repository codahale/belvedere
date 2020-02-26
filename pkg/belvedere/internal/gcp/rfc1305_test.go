package gcp

import (
	"strings"
	"testing"
)

func TestValidateRFC1035(t *testing.T) {
	goodValues := []string{
		"a", "ab", "abc", "a1", "a-1", "a--1--2--b",
		strings.Repeat("a", 63),
	}
	for _, val := range goodValues {
		if err := ValidateRFC1035(val); err != nil {
			t.Errorf("expected valid for '%s': %v", val, err)
		}
	}

	badValues := []string{
		"0", "01", "012", "1a", "1-a", "1--a--b--2",
		"", "A", "ABC", "aBc", "A1", "A-1", "1-A",
		"-", "a-", "-a", "1-", "-1",
		"_", "a_", "_a", "a_b", "1_", "_1", "1_2",
		".", "a.", ".a", "a.b", "1.", ".1", "1.2",
		" ", "a ", " a", "a b", "1 ", " 1", "1 2",
		strings.Repeat("a", 64),
	}
	for _, val := range badValues {
		if err := ValidateRFC1035(val); err == nil {
			t.Errorf("expected invalid for '%s'", val)
		}
	}
}
