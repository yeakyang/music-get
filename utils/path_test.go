package utils

import "testing"

func TestTrimInvalidFilePathChars(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{
			input: `*h|:e/\l"l<>o?`,
			want:  "hello",
		},
		{
			input: `**h||::e//\\l""l<<>>o??`,
			want:  "hello",
		},
	}

	for _, test := range tests {
		if got := TrimInvalidFilePathChars(test.input); got != test.want {
			t.Errorf("TrimInvalidFilePathChars(%q) got: %s, want: %s", test.input, got, test.want)
		}
	}
}
