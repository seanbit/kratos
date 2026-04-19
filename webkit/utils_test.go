package webkit

import (
	"strings"
	"testing"
)

func TestGenerateUniqueString(t *testing.T) {
	t.Run("valid size returns non-empty string without error", func(t *testing.T) {
		s, err := GenerateUniqueString(16)
		if err != nil {
			t.Fatalf("GenerateUniqueString(16) error = %v", err)
		}
		if s == "" {
			t.Error("GenerateUniqueString(16) returned empty string")
		}
	})

	t.Run("result contains timestamp prefix", func(t *testing.T) {
		s, err := GenerateUniqueString(8)
		if err != nil {
			t.Fatalf("GenerateUniqueString(8) error = %v", err)
		}
		// Timestamp is a unix epoch (10+ digits), followed by hex chars
		// The string should start with digits (the timestamp)
		hasDigitPrefix := false
		for _, c := range s {
			if c >= '0' && c <= '9' {
				hasDigitPrefix = true
				break
			}
		}
		if !hasDigitPrefix {
			t.Errorf("expected timestamp prefix in %q", s)
		}
	})

	t.Run("two calls produce different strings", func(t *testing.T) {
		s1, err1 := GenerateUniqueString(16)
		s2, err2 := GenerateUniqueString(16)
		if err1 != nil || err2 != nil {
			t.Fatalf("errors: %v, %v", err1, err2)
		}
		if s1 == s2 {
			t.Errorf("two calls should produce different strings, got %q both times", s1)
		}
	})

	t.Run("size 0 returns timestamp only", func(t *testing.T) {
		s, err := GenerateUniqueString(0)
		if err != nil {
			t.Fatalf("GenerateUniqueString(0) error = %v", err)
		}
		// With 0 random bytes, hex encoding is empty, so result is just the timestamp
		// Verify it's all digits (unix timestamp)
		for _, c := range s {
			if c < '0' || c > '9' {
				t.Errorf("size 0 should return digits only (timestamp), got %q", s)
				break
			}
		}
	})

	t.Run("larger size produces longer string", func(t *testing.T) {
		s4, _ := GenerateUniqueString(4)
		s32, _ := GenerateUniqueString(32)
		if len(s32) <= len(s4) {
			t.Errorf("size 32 (%d chars) should be longer than size 4 (%d chars)", len(s32), len(s4))
		}
	})
}

func TestFirstN(t *testing.T) {
	tests := []struct {
		name string
		s    string
		n    int
		want string
	}{
		{
			name: "string shorter than n returns full string",
			s:    "hi",
			n:    10,
			want: "hi",
		},
		{
			name: "string longer than n is truncated",
			s:    "hello world",
			n:    5,
			want: "hello",
		},
		{
			name: "string equal to n returns full string",
			s:    "exact",
			n:    5,
			want: "exact",
		},
		{
			name: "empty string stays empty",
			s:    "",
			n:    5,
			want: "",
		},
		{
			name: "n == 0 returns empty",
			s:    "hello",
			n:    0,
			want: "",
		},
		{
			name: "n == 1 returns first byte",
			s:    "hello",
			n:    1,
			want: "h",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FirstN(tt.s, tt.n)
			if got != tt.want {
				t.Errorf("FirstN(%q, %d) = %q, want %q", tt.s, tt.n, got, tt.want)
			}
		})
	}
}

func TestAddThousandsSeparatorFromInt(t *testing.T) {
	tests := []struct {
		name string
		num  int64
		want string
	}{
		{name: "zero", num: 0, want: "0"},
		{name: "single digit", num: 5, want: "5"},
		{name: "two digits", num: 42, want: "42"},
		{name: "three digits", num: 100, want: "100"},
		{name: "four digits", num: 1000, want: "1,000"},
		{name: "five digits", num: 12345, want: "12,345"},
		{name: "six digits", num: 100000, want: "100,000"},
		{name: "seven digits (million)", num: 1000000, want: "1,000,000"},
		{name: "large number", num: 1234567890, want: "1,234,567,890"},
		{name: "negative single", num: -5, want: "-5"},
		{name: "negative thousands", num: -1000, want: "-1,000"},
		{name: "negative million", num: -1000000, want: "-1,000,000"},
		{name: "negative large", num: -1234567890, want: "-1,234,567,890"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AddThousandsSeparatorFromInt(tt.num)
			if got != tt.want {
				t.Errorf("AddThousandsSeparatorFromInt(%d) = %q, want %q", tt.num, got, tt.want)
			}
		})
	}
}

func TestAddThousandsSeparatorFromFloatString(t *testing.T) {
	tests := []struct {
		name string
		num  string
		want string
	}{
		{name: "empty string", num: "", want: ""},
		{name: "integer string", num: "1000", want: "1,000"},
		{name: "float string", num: "1000.50", want: "1,000.50"},
		{name: "small float", num: "5.99", want: "5.99"},
		{name: "zero", num: "0", want: "0"},
		{name: "million with decimals", num: "1000000.123", want: "1,000,000.123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AddThousandsSeparatorFromFloatString(tt.num)
			if got != tt.want {
				t.Errorf("AddThousandsSeparatorFromFloatString(%q) = %q, want %q", tt.num, got, tt.want)
			}
		})
	}
}

func TestKeepDecimalPlaces(t *testing.T) {
	tests := []struct {
		name    string
		num     float64
		decimal int
		want    string
	}{
		{
			name:    "pi to 2 places",
			num:     3.14159,
			decimal: 2,
			want:    "3.14",
		},
		{
			name:    "half to 3 places pads zeros",
			num:     0.5,
			decimal: 3,
			want:    "0.500",
		},
		{
			name:    "whole number to 0 places",
			num:     100.0,
			decimal: 0,
			want:    "100.",
		},
		{
			name:    "zero to 2 places",
			num:     0.0,
			decimal: 2,
			want:    "0.00",
		},
		{
			name:    "large number to 1 place",
			num:     12345.6,
			decimal: 1,
			want:    "12345.6",
		},
		{
			name:    "small fraction to 4 places",
			num:     0.1234,
			decimal: 4,
			want:    "0.1234",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := KeepDecimalPlaces(tt.num, tt.decimal)
			if got != tt.want {
				t.Errorf("KeepDecimalPlaces(%v, %d) = %q, want %q", tt.num, tt.decimal, got, tt.want)
			}
		})
	}
}

func TestKeepDecimalPlaces_ContainsDot(t *testing.T) {
	// Every result should contain exactly one dot
	cases := []struct {
		num     float64
		decimal int
	}{
		{1.5, 1},
		{100.0, 0},
		{0.0, 3},
		{99.99, 2},
	}

	for _, c := range cases {
		result := KeepDecimalPlaces(c.num, c.decimal)
		if count := strings.Count(result, "."); count != 1 {
			t.Errorf("KeepDecimalPlaces(%v, %d) = %q has %d dots, want exactly 1",
				c.num, c.decimal, result, count)
		}
	}
}
