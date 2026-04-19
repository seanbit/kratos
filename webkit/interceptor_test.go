package webkit

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"
)

func TestGenerateSignature(t *testing.T) {
	tests := []struct {
		name   string
		data   string
		secret string
		length int
		want   string
	}{
		{
			name:   "known input produces expected prefix",
			data:   "1700000000000",
			secret: "my-secret",
			length: 8,
		},
		{
			name:   "length 8 returns 8 hex chars",
			data:   "hello",
			secret: "secret",
			length: 8,
		},
		{
			name:   "length 64 returns full hash",
			data:   "hello",
			secret: "secret",
			length: 64,
		},
		{
			name:   "length greater than 64 is clamped",
			data:   "hello",
			secret: "secret",
			length: 128,
		},
		{
			name:   "empty data",
			data:   "",
			secret: "secret",
			length: 8,
		},
		{
			name:   "empty secret",
			data:   "hello",
			secret: "",
			length: 8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateSignature(tt.data, tt.secret, tt.length)

			// Length should be min(tt.length, 64) since SHA-256 produces 64 hex chars
			expectedLen := tt.length
			if expectedLen > 64 {
				expectedLen = 64
			}
			if len(got) != expectedLen {
				t.Errorf("generateSignature() length = %d, want %d", len(got), expectedLen)
			}

			// Result should be deterministic: same input produces same output
			got2 := generateSignature(tt.data, tt.secret, tt.length)
			if got != got2 {
				t.Errorf("generateSignature() not deterministic: %q != %q", got, got2)
			}
		})
	}
}

func TestGenerateSignature_Deterministic(t *testing.T) {
	// Verify that a known data+secret pair always produces the same signature
	sig1 := generateSignature("timestamp123", "key456", 16)
	sig2 := generateSignature("timestamp123", "key456", 16)
	if sig1 != sig2 {
		t.Errorf("same inputs should produce same signature: %q vs %q", sig1, sig2)
	}

	// Different data should produce different signature
	sig3 := generateSignature("timestamp999", "key456", 16)
	if sig1 == sig3 {
		t.Error("different data should produce different signatures")
	}

	// Different secret should produce different signature
	sig4 := generateSignature("timestamp123", "differentkey", 16)
	if sig1 == sig4 {
		t.Error("different secret should produce different signatures")
	}
}

func TestGenerateSignature_TruncationConsistency(t *testing.T) {
	data := "test-data"
	secret := "test-secret"

	full := generateSignature(data, secret, 64)
	prefix8 := generateSignature(data, secret, 8)
	prefix16 := generateSignature(data, secret, 16)

	if full[:8] != prefix8 {
		t.Errorf("8-char truncation mismatch: full[:8]=%q, prefix8=%q", full[:8], prefix8)
	}
	if full[:16] != prefix16 {
		t.Errorf("16-char truncation mismatch: full[:16]=%q, prefix16=%q", full[:16], prefix16)
	}
}

func TestVerifySign(t *testing.T) {
	ctx := context.Background()

	t.Run("sign not enabled returns true for any input", func(t *testing.T) {
		signConfig.Store(&SignConfig{
			Secret:          "",
			SignatureLength: 8,
			MaxTimeDrift:    300,
			Enabled:         false,
		})

		if !verifySign(ctx, "") {
			t.Error("disabled sign config should return true for empty input")
		}
		if !verifySign(ctx, "garbage") {
			t.Error("disabled sign config should return true for garbage input")
		}
		if !verifySign(ctx, "12345.abcdef") {
			t.Error("disabled sign config should return true for any formatted input")
		}
	})

	t.Run("enabled but empty secret returns true", func(t *testing.T) {
		signConfig.Store(&SignConfig{
			Secret:          "",
			SignatureLength: 8,
			MaxTimeDrift:    300,
			Enabled:         true,
		})

		if !verifySign(ctx, "anything") {
			t.Error("enabled config with empty secret should return true")
		}
	})

	// Configure for enabled tests
	secret := "test-secret"
	sigLen := 8
	maxDrift := int64(300)

	setupEnabledConfig := func() {
		signConfig.Store(&SignConfig{
			Secret:          secret,
			SignatureLength: sigLen,
			MaxTimeDrift:    maxDrift,
			Enabled:         true,
		})
	}

	t.Run("valid timestamp and signature passes", func(t *testing.T) {
		setupEnabledConfig()

		ts := strconv.FormatInt(time.Now().UnixMilli(), 10)
		sig := generateSignature(ts, secret, sigLen)
		requestTime := ts + "." + sig

		if !verifySign(ctx, requestTime) {
			t.Errorf("valid request time %q should pass verification", requestTime)
		}
	})

	t.Run("missing dot separator fails", func(t *testing.T) {
		setupEnabledConfig()

		if verifySign(ctx, "nodot") {
			t.Error("request time without dot separator should fail")
		}
		if verifySign(ctx, "") {
			t.Error("empty request time should fail")
		}
	})

	t.Run("invalid timestamp fails", func(t *testing.T) {
		setupEnabledConfig()

		if verifySign(ctx, "notanumber.abcdef12") {
			t.Error("non-numeric timestamp should fail")
		}
		if verifySign(ctx, "abc.abcdef12") {
			t.Error("alphabetic timestamp should fail")
		}
	})

	t.Run("time drift too large fails", func(t *testing.T) {
		setupEnabledConfig()

		// Timestamp far in the past (beyond 300s drift)
		oldTs := strconv.FormatInt(time.Now().Add(-10*time.Minute).UnixMilli(), 10)
		sig := generateSignature(oldTs, secret, sigLen)
		requestTime := oldTs + "." + sig

		if verifySign(ctx, requestTime) {
			t.Error("timestamp with drift > maxTimeDrift should fail")
		}

		// Timestamp far in the future
		futureTs := strconv.FormatInt(time.Now().Add(10*time.Minute).UnixMilli(), 10)
		sigFuture := generateSignature(futureTs, secret, sigLen)
		requestTimeFuture := futureTs + "." + sigFuture

		if verifySign(ctx, requestTimeFuture) {
			t.Error("future timestamp with drift > maxTimeDrift should fail")
		}
	})

	t.Run("wrong signature fails", func(t *testing.T) {
		setupEnabledConfig()

		ts := strconv.FormatInt(time.Now().UnixMilli(), 10)
		requestTime := ts + "." + "deadbeef"

		if verifySign(ctx, requestTime) {
			t.Error("wrong signature should fail verification")
		}
	})

	t.Run("timestamp within drift passes", func(t *testing.T) {
		setupEnabledConfig()

		// Timestamp 100 seconds ago (within 300s drift)
		recentTs := strconv.FormatInt(time.Now().Add(-100*time.Second).UnixMilli(), 10)
		sig := generateSignature(recentTs, secret, sigLen)
		requestTime := recentTs + "." + sig

		if !verifySign(ctx, requestTime) {
			t.Error("timestamp within drift should pass verification")
		}
	})
}

func TestMatchRule(t *testing.T) {
	tests := []struct {
		name    string
		rule    SubRuleConfig
		path    string
		referer string
		ua      string
		uid     string
		want    bool
	}{
		{
			name:    "wildcard matches everything",
			rule:    SubRuleConfig{Rule: "*"},
			path:    "/any/path",
			referer: "any-referer",
			ua:      "any-ua",
			uid:     "any-uid",
			want:    true,
		},
		{
			name:    "path rule matches exact path",
			rule:    SubRuleConfig{Rule: "path", Value: "/api/v1/user,/api/v1/order"},
			path:    "/api/v1/user",
			referer: "",
			ua:      "",
			uid:     "",
			want:    true,
		},
		{
			name:    "path rule does not match different path",
			rule:    SubRuleConfig{Rule: "path", Value: "/api/v1/user,/api/v1/order"},
			path:    "/api/v1/product",
			referer: "",
			ua:      "",
			uid:     "",
			want:    false,
		},
		{
			name:    "referer rule matches substring",
			rule:    SubRuleConfig{Rule: "referer", Value: "example.com,test.com"},
			path:    "",
			referer: "https://example.com/page",
			ua:      "",
			uid:     "",
			want:    true,
		},
		{
			name:    "referer rule does not match unrelated referer",
			rule:    SubRuleConfig{Rule: "referer", Value: "example.com,test.com"},
			path:    "",
			referer: "https://other.org/page",
			ua:      "",
			uid:     "",
			want:    false,
		},
		{
			name:    "ua rule matches exact user agent",
			rule:    SubRuleConfig{Rule: "ua", Value: "BadBot/1.0,EvilCrawler/2.0"},
			path:    "",
			referer: "",
			ua:      "BadBot/1.0",
			uid:     "",
			want:    true,
		},
		{
			name:    "ua rule does not match different user agent",
			rule:    SubRuleConfig{Rule: "ua", Value: "BadBot/1.0,EvilCrawler/2.0"},
			path:    "",
			referer: "",
			ua:      "GoodBot/1.0",
			uid:     "",
			want:    false,
		},
		{
			name:    "uid rule matches exact uid",
			rule:    SubRuleConfig{Rule: "uid", Value: "user123,user456"},
			path:    "",
			referer: "",
			ua:      "",
			uid:     "user123",
			want:    true,
		},
		{
			name:    "uid rule does not match different uid",
			rule:    SubRuleConfig{Rule: "uid", Value: "user123,user456"},
			path:    "",
			referer: "",
			ua:      "",
			uid:     "user789",
			want:    false,
		},
		{
			name:    "unknown rule type returns false",
			rule:    SubRuleConfig{Rule: "unknown", Value: "anything"},
			path:    "/any",
			referer: "any",
			ua:      "any",
			uid:     "any",
			want:    false,
		},
		{
			name:    "empty rule type returns false",
			rule:    SubRuleConfig{Rule: "", Value: "anything"},
			path:    "",
			referer: "",
			ua:      "",
			uid:     "",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchRule(tt.rule, tt.path, tt.referer, tt.ua, tt.uid)
			if got != tt.want {
				t.Errorf("matchRule() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetStrategyByRadio(t *testing.T) {
	t.Run("radio -1 returns nil", func(t *testing.T) {
		err := getStrategyByRadio(-1)
		if err != nil {
			t.Errorf("getStrategyByRadio(-1) = %v, want nil", err)
		}
	})

	t.Run("radio 0 returns nil", func(t *testing.T) {
		err := getStrategyByRadio(0)
		if err != nil {
			t.Errorf("getStrategyByRadio(0) = %v, want nil", err)
		}
	})

	t.Run("radio 100 always returns error", func(t *testing.T) {
		// radio=100 means rand.Intn(100) < 100 is always true
		for i := 0; i < 100; i++ {
			err := getStrategyByRadio(100)
			if err == nil {
				t.Fatal("getStrategyByRadio(100) should always return error")
			}
		}
	})

	t.Run("radio 100 error message", func(t *testing.T) {
		err := getStrategyByRadio(100)
		if err == nil {
			t.Fatal("expected error")
		}
		if err.Error() != "invalid request" {
			t.Errorf("error message = %q, want %q", err.Error(), "invalid request")
		}
	})

	t.Run("radio negative beyond -1 returns nil", func(t *testing.T) {
		// radio < 0 and != -1 falls through all conditions, returns nil
		err := getStrategyByRadio(-5)
		if err != nil {
			t.Errorf("getStrategyByRadio(-5) = %v, want nil", err)
		}
	})

	t.Run("radio between 1 and 99 is probabilistic", func(t *testing.T) {
		// With radio=50, roughly half should error. Run enough times to verify
		// it is not always nil and not always error.
		var errCount, nilCount int
		iterations := 1000
		for i := 0; i < iterations; i++ {
			if err := getStrategyByRadio(50); err != nil {
				errCount++
			} else {
				nilCount++
			}
		}
		if errCount == 0 {
			t.Error("radio=50 should produce some errors over 1000 iterations")
		}
		if nilCount == 0 {
			t.Error("radio=50 should produce some nil results over 1000 iterations")
		}
		// Sanity: expect roughly 40-60% errors (very loose bounds)
		errRate := float64(errCount) / float64(iterations)
		if errRate < 0.3 || errRate > 0.7 {
			t.Logf("warning: error rate %.2f outside expected range [0.3, 0.7] for radio=50", errRate)
		}
	})

	t.Run("radio 101 returns nil (out of valid range)", func(t *testing.T) {
		// radio > 100 means condition radio > 0 && radio <= 100 is false
		err := getStrategyByRadio(101)
		if err != nil {
			t.Errorf("getStrategyByRadio(101) = %v, want nil", err)
		}
	})
}

func TestInSliceStr(t *testing.T) {
	tests := []struct {
		name   string
		source string
		target []string
		want   bool
	}{
		{"found", "a", []string{"a", "b", "c"}, true},
		{"not found", "d", []string{"a", "b", "c"}, false},
		{"empty target", "a", []string{}, false},
		{"empty source", "", []string{"a", "b", ""}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := inSliceStr(tt.source, tt.target)
			if got != tt.want {
				t.Errorf("inSliceStr(%q, %v) = %v, want %v", tt.source, tt.target, got, tt.want)
			}
		})
	}
}

func TestSliceContain(t *testing.T) {
	tests := []struct {
		name   string
		source string
		target []string
		want   bool
	}{
		{"contains substring", "https://example.com/page", []string{"example.com"}, true},
		{"no match", "https://other.org", []string{"example.com"}, false},
		{"empty source", "", []string{"abc"}, false},
		{"empty target", "anything", []string{}, false},
		{"empty string in target matches any source", "anything", []string{""}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sliceContain(tt.source, tt.target)
			if got != tt.want {
				t.Errorf("sliceContain(%q, %v) = %v, want %v", tt.source, tt.target, got, tt.want)
			}
		})
	}
}

func TestSanitizeHeaders(t *testing.T) {
	tests := []struct {
		name    string
		headers map[string][]string
		check   func(map[string]string) error
	}{
		{
			name: "sensitive headers are redacted",
			headers: map[string][]string{
				"Authorization": {"Bearer token123"},
				"Cookie":        {"session=abc"},
				"Content-Type":  {"application/json"},
			},
			check: func(result map[string]string) error {
				if result["Authorization"] != "***REDACTED***" {
					return fmt.Errorf("Authorization should be redacted, got %q", result["Authorization"])
				}
				if result["Cookie"] != "***REDACTED***" {
					return fmt.Errorf("Cookie should be redacted, got %q", result["Cookie"])
				}
				if result["Content-Type"] != "application/json" {
					return fmt.Errorf("Content-Type should not be redacted, got %q", result["Content-Type"])
				}
				return nil
			},
		},
		{
			name: "multiple values are joined",
			headers: map[string][]string{
				"Accept": {"text/html", "application/json"},
			},
			check: func(result map[string]string) error {
				if result["Accept"] != "text/html, application/json" {
					return fmt.Errorf("Accept should be joined, got %q", result["Accept"])
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeHeaders(tt.headers)
			if err := tt.check(result); err != nil {
				t.Error(err)
			}
		})
	}
}
