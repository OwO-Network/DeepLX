//go:build e2e

// End-to-end tests that exercise the real DeepL upstream through
// TranslateByDeepLX. Gated behind the `e2e` build tag so the default
// `go test ./...` (and existing CI) never hit the network or risk
// tripping DeepL's rate limiting. Run explicitly:
//
//	go test -tags=e2e -v ./translate/...
//
// A 429 means THIS IP got rate-limited, not that the code is broken —
// those cases back off and ultimately t.Skip rather than t.Fail, so the
// suite stays a signal for "DeepLX can still reach DeepL", not a flaky
// red build.
package translate_test

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/OwO-Network/DeepLX/translate"
)

const (
	e2ePace       = 2 * time.Second
	e2eBackoff    = 3 * time.Second
	e2eMaxRetries = 3
)

// translateFree paces the call and backs off on 429. A persistent 429
// skips the test (upstream/IP problem); anything else is returned to
// assert on.
func translateFree(t *testing.T, source, target, text string) translate.DeepLXTranslationResult {
	t.Helper()
	var res translate.DeepLXTranslationResult
	for attempt := 1; attempt <= e2eMaxRetries; attempt++ {
		time.Sleep(e2ePace)
		var err error
		res, err = translate.TranslateByDeepLX(source, target, text, "", "", "")
		if err != nil {
			t.Fatalf("TranslateByDeepLX returned a transport error: %v", err)
		}
		if res.Code == http.StatusTooManyRequests {
			wait := time.Duration(attempt) * e2eBackoff
			t.Logf("attempt %d/%d: DeepL returned 429 (this IP is rate-limited); backing off %s", attempt, e2eMaxRetries, wait)
			time.Sleep(wait)
			continue
		}
		return res
	}
	t.Skipf("DeepL kept returning 429 after %d attempts — upstream rate limit / IP block, not a code defect", e2eMaxRetries)
	return res
}

func TestE2EFreeBasic(t *testing.T) {
	res := translateFree(t, "EN", "ZH", "Hello, world")
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Message)
	}
	if res.Data == "" {
		t.Fatal("expected a non-empty translation, got empty Data")
	}
	if res.Method != "Free" {
		t.Errorf("expected Method=Free, got %q", res.Method)
	}
	t.Logf("EN->ZH %q => %q", "Hello, world", res.Data)
}

func TestE2EFreeAutoDetect(t *testing.T) {
	// Empty source => server-side language autodetection.
	res := translateFree(t, "", "EN", "Bonjour le monde")
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Message)
	}
	if res.Data == "" {
		t.Fatal("expected a non-empty translation, got empty Data")
	}
	if res.SourceLang == "" {
		t.Error("expected a detected SourceLang, got empty")
	}
	t.Logf("auto->EN %q => %q (detected %s)", "Bonjour le monde", res.Data, res.SourceLang)
}

func TestE2EFreeJapanese(t *testing.T) {
	res := translateFree(t, "EN", "JA", "Good morning")
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Message)
	}
	if res.Data == "" {
		t.Fatal("expected a non-empty translation, got empty Data")
	}
}

// TestE2ENoRateLimitOnNormalUse is the core of "make sure it doesn't 429":
// a handful of properly-paced requests should essentially all succeed. If
// every one is blocked we skip (IP problem); otherwise any 429 on paced
// traffic means the browser impersonation is degrading and needs a look.
func TestE2ENoRateLimitOnNormalUse(t *testing.T) {
	phrases := []string{"one", "two", "three", "four", "five"}
	var ok, rateLimited int
	for i, p := range phrases {
		time.Sleep(e2ePace)
		res, err := translate.TranslateByDeepLX("EN", "ZH", p, "", "", "")
		if err != nil {
			t.Fatalf("round %d: unexpected transport error: %v", i, err)
		}
		switch res.Code {
		case http.StatusOK:
			ok++
		case http.StatusTooManyRequests:
			rateLimited++
		default:
			t.Errorf("round %d: unexpected code %d: %s", i, res.Code, res.Message)
		}
	}
	t.Logf("normal-use rounds=%d ok=%d rateLimited=%d", len(phrases), ok, rateLimited)
	if ok == 0 {
		t.Skipf("all %d requests were rate-limited — upstream/IP block, not a code defect", len(phrases))
	}
	if rateLimited > 0 {
		t.Errorf("got %d/%d 429s on properly-paced requests; DeepL impersonation may be degrading", rateLimited, len(phrases))
	}
}

// TestE2EInputValidation never reaches the network — these inputs are
// rejected before any DeepL call, so it is safe and free even under -tags=e2e.
func TestE2EInputValidation(t *testing.T) {
	cases := []struct {
		name           string
		source, target string
		text           string
		wantCode       int
	}{
		{"empty text", "EN", "ZH", "", http.StatusNotFound},
		{"target auto", "EN", "auto", "hello", http.StatusBadRequest},
		{"unsupported target", "EN", "XX", "hello", http.StatusBadRequest},
		{"too long", "EN", "ZH", strings.Repeat("a", 1501), http.StatusRequestEntityTooLarge},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := translate.TranslateByDeepLX(tc.source, tc.target, tc.text, "", "", "")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if res.Code != tc.wantCode {
				t.Errorf("want code %d, got %d: %s", tc.wantCode, res.Code, res.Message)
			}
		})
	}
}
