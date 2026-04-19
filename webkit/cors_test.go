package webkit

import (
	"strings"
	"testing"

	"github.com/go-kratos/kratos/v2/transport"
)

// mockHeader implements transport.Header for testing.
type mockHeader struct {
	vals map[string]string
}

func newMockHeader() *mockHeader {
	return &mockHeader{vals: make(map[string]string)}
}

func (h *mockHeader) Get(key string) string {
	return h.vals[key]
}

func (h *mockHeader) Set(key string, value string) {
	h.vals[key] = value
}

func (h *mockHeader) Add(key string, value string) {
	h.vals[key] = value
}

func (h *mockHeader) Keys() []string {
	keys := make([]string, 0, len(h.vals))
	for k := range h.vals {
		keys = append(keys, k)
	}
	return keys
}

func (h *mockHeader) Values(key string) []string {
	if v, ok := h.vals[key]; ok {
		return []string{v}
	}
	return nil
}

// mockTransporter implements transport.Transporter for testing.
type mockTransporter struct {
	reqHeader *mockHeader
	repHeader *mockHeader
}

func newMockTransporter() *mockTransporter {
	return &mockTransporter{
		reqHeader: newMockHeader(),
		repHeader: newMockHeader(),
	}
}

func (t *mockTransporter) Kind() transport.Kind          { return transport.KindHTTP }
func (t *mockTransporter) Endpoint() string              { return "" }
func (t *mockTransporter) Operation() string             { return "" }
func (t *mockTransporter) RequestHeader() transport.Header { return t.reqHeader }
func (t *mockTransporter) ReplyHeader() transport.Header   { return t.repHeader }

func TestDealWithHeader_AllowedDomain(t *testing.T) {
	tr := newMockTransporter()
	tr.reqHeader.Set(Origin, "https://example.com")

	err := DealWithHeader(tr, []string{"example.com"})
	if err != nil {
		t.Fatalf("expected no error for allowed domain, got: %v", err)
	}
}

func TestDealWithHeader_SubdomainMatch(t *testing.T) {
	tr := newMockTransporter()
	tr.reqHeader.Set(Origin, "https://sub.example.com")

	err := DealWithHeader(tr, []string{"example.com"})
	if err != nil {
		t.Fatalf("expected no error for subdomain match, got: %v", err)
	}
}

func TestDealWithHeader_DisallowedDomain(t *testing.T) {
	tr := newMockTransporter()
	tr.reqHeader.Set(Origin, "https://evil.com")

	err := DealWithHeader(tr, []string{"example.com"})
	if err == nil {
		t.Fatal("expected error for disallowed domain, got nil")
	}
}

func TestDealWithHeader_EmptyOrigin(t *testing.T) {
	tr := newMockTransporter()
	// No Origin or Referer set.

	err := DealWithHeader(tr, []string{"example.com"})
	if err != nil {
		t.Fatalf("expected no error for empty origin, got: %v", err)
	}
}

func TestDealWithHeader_FallsBackToReferer(t *testing.T) {
	tr := newMockTransporter()
	tr.reqHeader.Set(Referer, "https://example.com/page")

	err := DealWithHeader(tr, []string{"example.com"})
	if err != nil {
		t.Fatalf("expected no error when Referer matches allowed domain, got: %v", err)
	}
}

func TestDealWithHeader_CORSHeadersSet(t *testing.T) {
	tr := newMockTransporter()
	tr.reqHeader.Set(Origin, "https://example.com")

	err := DealWithHeader(tr, []string{"example.com"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		header string
		want   string
	}{
		{AllowOrigin, "https://example.com"},
		{AllowMethods, "GET,POST,OPTIONS,PUT,PATCH,DELETE"},
		{AllowCredentials, "true"},
		{ExposeHeaders, "Content-Length,X-Token"},
		{SecurityHeaders, "max-age=31536000"},
	}

	for _, tc := range tests {
		got := tr.repHeader.Get(tc.header)
		if got != tc.want {
			t.Errorf("header %s = %q, want %q", tc.header, got, tc.want)
		}
	}
}

func TestDealWithHeader_AllowHeadersContainsPlatformAndXRequestedWith(t *testing.T) {
	tr := newMockTransporter()
	tr.reqHeader.Set(Origin, "https://example.com")

	err := DealWithHeader(tr, []string{"example.com"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	allowHeaders := tr.repHeader.Get(AllowHeaders)

	// Verify that Platform and X-Requested-With are separated by a comma (bug fix check).
	expected := "Platform,X-Requested-With"
	if !strings.Contains(allowHeaders, expected) {
		t.Errorf("AllowHeaders should contain %q with comma separation, got: %q", expected, allowHeaders)
	}
}

func TestDealWithHeader_DoesNotOverwriteExistingHeaders(t *testing.T) {
	tr := newMockTransporter()
	tr.reqHeader.Set(Origin, "https://example.com")
	tr.repHeader.Set(AllowOrigin, "https://other.com")

	err := DealWithHeader(tr, []string{"example.com"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := tr.repHeader.Get(AllowOrigin); got != "https://other.com" {
		t.Errorf("AllowOrigin should not be overwritten, got %q, want %q", got, "https://other.com")
	}
}

func TestDealWithHeader_EmptyOriginNoAllowOriginSet(t *testing.T) {
	tr := newMockTransporter()
	// No Origin header set.

	err := DealWithHeader(tr, []string{"example.com"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := tr.repHeader.Get(AllowOrigin); got != "" {
		t.Errorf("AllowOrigin should not be set when origin is empty, got %q", got)
	}
}
