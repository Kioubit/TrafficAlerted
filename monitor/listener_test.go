package monitor

import (
	"testing"
)

func TestParseIPVersion(t *testing.T) {
	type TestCase struct {
		got  byte
		want int
	}

	cases := []TestCase{
		{0x60, 6},
		{0x6F, 6},
		{0x6D, 6},
		{0x40, 4},
		{0x4F, 4},
		{0x4D, 4},
		{0xEF, 0},
		{0x00, 0},
	}

	for _, testCase := range cases {
		res := parseIPVersion(testCase.got)
		if res != testCase.want {
			t.Errorf("Got: %d, want %d for %x", res, testCase.want, testCase.got)
		}
	}
}
