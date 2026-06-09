package model

import "testing"

func TestParseMotivationNumericAndLegacyValues(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Motivation
	}{
		{name: "legacy low", input: "baja", expected: MotivationLow},
		{name: "legacy medium", input: "media", expected: MotivationMedium},
		{name: "legacy high", input: "alta", expected: MotivationHigh},
		{name: "numeric", input: "7", expected: Motivation(7)},
		{name: "trimmed numeric", input: " 10 ", expected: MotivationHigh},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMotivation(tt.input)
			if err != nil {
				t.Fatalf("ParseMotivation(%q) returned error: %v", tt.input, err)
			}
			if got != tt.expected {
				t.Fatalf("ParseMotivation(%q) = %v, want %v", tt.input, got, tt.expected)
			}
			if got.String() != tt.expected.String() {
				t.Fatalf("String() = %q, want %q", got.String(), tt.expected.String())
			}
		})
	}
}

func TestParseMotivationRejectsOutOfRangeValues(t *testing.T) {
	for _, input := range []string{"-1", "11", "abc"} {
		if _, err := ParseMotivation(input); err == nil {
			t.Fatalf("ParseMotivation(%q) should fail", input)
		}
	}
}
