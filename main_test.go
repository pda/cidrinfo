package main

import (
	"testing"
)

func TestMaskLineLength(t *testing.T) {
	expectations := []string{
		"",
		"1",
		"2 ",
		"|3|",
		"|4 |",
		"| 5 |",
		"| 6 -|",
		"|- 7 -|",
		"|- 8 --|",
		"|-- 9 ---|",
		"|-- 10 ---|",
		"|--- 11 ---|",
		"|--- 12 ----|",
		"|---- 13 ----|",
		"|---- 14 -----|",
		"|----- 15 -----|",
		"|----- 16 ------|",
		"|------ 17 -------|",
		"|------- 18 -------|",
		// etc
	}
	for i := 0; i < len(expectations); i++ {
		l := maskLine(i)

		expectedLength := i + ((i - 1) / 8)
		if len(l) != expectedLength {
			t.Errorf("expected length of %d+%d=%d, got %d", i, (i-1)/8, expectedLength, len(l))
		}

		if l != expectations[i] {
			t.Errorf("\ngot      \"%s\"\nexpected \"%s\"", l, expectations[i])
		}
	}
}
