package hw02unpackstring

import (
	"errors"
	"testing"

	//nolint:depguard
	"github.com/stretchr/testify/require"
)

func TestUnpack(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{input: "a4bc2d5e", expected: "aaaabccddddde"},
		{input: "abccd", expected: "abccd"},
		{input: "", expected: ""},
		{input: "aaa0b", expected: "aab"},
		// uncomment if task with asterisk completed
		{input: `qwe\4\5`, expected: `qwe45`},
		{input: `qwe\45`, expected: `qwe44444`},
		{input: `qwe\\5`, expected: `qwe\\\\\`},
		{input: `qwe\\\3`, expected: `qwe\3`},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			result, err := Unpack(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestUnpackInvalidString(t *testing.T) {
	invalidStrings := []string{"3abc", "45", "aaa10b"}
	for _, tc := range invalidStrings {
		tc := tc
		t.Run(tc, func(t *testing.T) {
			_, err := Unpack(tc)
			require.Truef(t, errors.Is(err, ErrInvalidString), "actual error %q", err)
		})
	}
}

func TestUnpackCustom(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		err      bool
	}{
		{input: "", expected: "", err: false},
		{input: "a", expected: "a", err: false},
		{input: "a1b", expected: "ab", err: false},
		{input: `a\\b`, expected: `a\b`, err: false},
		{input: "a11b", expected: "", err: true},
		{input: "a2b3", expected: "aabbb", err: false},
		{input: `a\2b\3`, expected: `a2b3`, err: false},
		{input: `a\12b`, expected: `a11b`, err: false},
		{input: `a1\2b\3c4\5`, expected: `a2b3cccc5`, err: false},
		{input: `a1\2b\3c4\`, expected: "", err: true},
		{input: `aaa\`, expected: "", err: true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			result, err := Unpack(tc.input)
			switch {
			case err != nil && !tc.err:
				t.Errorf("Unpack(%q) returned error %v, want nil", tc.input, err)
			case err == nil && tc.err:
				t.Errorf("Unpack(%q) returned nil, want error", tc.input)
			case result != tc.expected:
				t.Errorf("Unpack(%q) returned %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}
