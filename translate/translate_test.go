package translate

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"testing"
)

func TestResolveTargetLang(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		want    string
		wantErr bool
		errHas  string
	}{
		{"lowercase normalizes to canonical", "zh-hans", "zh-Hans", false, ""},
		{"uppercase ZH maps to Simplified", "ZH", "zh-Hans", false, ""},
		{"EN convenience alias resolves to en-US", "EN", "en-US", false, ""},
		{"PT convenience alias resolves to pt-BR", "PT", "pt-BR", false, ""},
		{"regional EN-GB preserved", "en-gb", "en-GB", false, ""},
		{"plain de", "de", "de", false, ""},
		{"empty is required", "", "", true, "required"},
		{"auto is rejected as a target", "auto", "", true, "cannot be"},
		{"AUTO rejected case-insensitively", "AUTO", "", true, "cannot be"},
		{"unsupported code is reported", "xx", "", true, "unsupported target_lang"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := resolveTargetLang(tc.in)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("resolveTargetLang(%q) = %q, want error", tc.in, got)
				}
				if tc.errHas != "" && !strings.Contains(err.Error(), tc.errHas) {
					t.Errorf("error %q does not contain %q", err.Error(), tc.errHas)
				}
				if got != "" {
					t.Errorf("on error want empty result, got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("resolveTargetLang(%q) unexpected error: %v", tc.in, err)
			}
			if got != tc.want {
				t.Errorf("resolveTargetLang(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestResolveSourceLang(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{"empty means autodetect", "", "", false},
		{"auto means autodetect", "auto", "", false},
		{"AUTO case-insensitive", "AUTO", "", false},
		// EN/PT are first-class *source* codes and resolve to the generic
		// form, unlike the target map which forces a regional default
		// (EN -> en-US there). This asymmetry is easy to regress.
		{"EN as source resolves to generic en", "EN", "en", false},
		{"PT as source resolves to generic pt", "PT", "pt", false},
		{"target-only code ZH is also a valid source", "ZH", "zh-Hans", false},
		{"plain de", "de", "de", false},
		{"unsupported code errors", "xx", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := resolveSourceLang(tc.in)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("resolveSourceLang(%q) = %q, want error", tc.in, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("resolveSourceLang(%q) unexpected error: %v", tc.in, err)
			}
			if got != tc.want {
				t.Errorf("resolveSourceLang(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

// TestTranslateByDeepLXPreflight exercises only the early-return validation
// paths that complete BEFORE any network I/O, so the suite stays hermetic.
func TestTranslateByDeepLXPreflight(t *testing.T) {
	cases := []struct {
		name     string
		source   string
		target   string
		text     string
		wantCode int
		msgHas   string
	}{
		{"empty text returns 404 (shipped contract)", "", "EN", "", http.StatusNotFound, "No text to translate"},
		{"invalid target returns 400", "", "XX", "hello", http.StatusBadRequest, "unsupported target_lang"},
		{"invalid source returns 400", "XX", "EN", "hello", http.StatusBadRequest, "unsupported source_lang"},
		{"over-length ASCII returns 413", "", "EN", strings.Repeat("a", maxFreeTextLength+1), http.StatusRequestEntityTooLarge, "exceeds maximum length"},
		// The cap counts runes, not bytes: 1501 multibyte runes is well over
		// 1500 even though it is ~4500 bytes.
		{"over-length multibyte counts runes returns 413", "", "EN", strings.Repeat("你", maxFreeTextLength+1), http.StatusRequestEntityTooLarge, "exceeds maximum length"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := TranslateByDeepLX(tc.source, tc.target, tc.text, "", "", "")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if res.Code != tc.wantCode {
				t.Fatalf("Code = %d, want %d (message: %q)", res.Code, tc.wantCode, res.Message)
			}
			if tc.msgHas != "" && !strings.Contains(res.Message, tc.msgHas) {
				t.Errorf("Message %q does not contain %q", res.Message, tc.msgHas)
			}
		})
	}
}

var uuidV4Re = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func TestNewInstanceIDIsUUIDv4(t *testing.T) {
	for i := 0; i < 100; i++ {
		id := newInstanceID()
		if !uuidV4Re.MatchString(id) {
			t.Fatalf("newInstanceID() = %q, not a valid RFC 4122 v4 UUID", id)
		}
	}
	// The process-wide instanceID must itself be a valid v4 UUID.
	if !uuidV4Re.MatchString(instanceID) {
		t.Errorf("package instanceID = %q, not a valid v4 UUID", instanceID)
	}
}

// TestOneshotRequestJSONShape locks the serialized request body: the field
// order and omitempty behavior are reverse-engineered to be byte-identical
// to the Chrome extension, so a struct-field reorder or tag change is a
// silent WAF signal that this test turns into a failure.
func TestOneshotRequestJSONShape(t *testing.T) {
	anon := oneshotRequest{
		Text:       []string{"hello"},
		TargetLang: "de",
		UsageType:  "Translate",
		AppInformation: appInformation{
			OS:         "brex_macOS",
			OSVersion:  "brex_chrome_" + impersonatedChromeMajor + ".0.0.0",
			AppVersion: chromeExtensionVersion,
			AppBuild:   "chrome_web_store",
			InstanceID: instanceID,
		},
	}
	b, err := json.Marshal(anon)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(b)

	if strings.Contains(s, "source_lang") {
		t.Errorf("source_lang must be omitted when empty, got %s", s)
	}
	for _, key := range []string{`"text"`, `"target_lang"`, `"usage_type"`, `"app_information"`, `"os_version"`, `"app_version"`, `"app_build"`, `"instance_id"`} {
		if !strings.Contains(s, key) {
			t.Errorf("missing expected key %s in %s", key, s)
		}
	}
	// encoding/json emits struct fields in declaration order; the body
	// contract depends on it.
	assertOrder(t, s, `"text"`, `"target_lang"`, `"usage_type"`, `"app_information"`)

	// os_version must track the pinned Chrome major so the TLS handshake,
	// user-agent, and self-reported version all tell one consistent story.
	if !strings.Contains(s, "brex_chrome_"+impersonatedChromeMajor+".0.0.0") {
		t.Errorf("os_version should embed impersonatedChromeMajor %q; got %s", impersonatedChromeMajor, s)
	}

	// With a source language set, the field appears with its value.
	withSrc := anon
	withSrc.SourceLang = "en"
	b2, _ := json.Marshal(withSrc)
	if !strings.Contains(string(b2), `"source_lang":"en"`) {
		t.Errorf("source_lang should be present when set, got %s", string(b2))
	}
}

func assertOrder(t *testing.T, haystack string, needles ...string) {
	t.Helper()
	last := -1
	for _, n := range needles {
		idx := strings.Index(haystack, n)
		if idx == -1 {
			t.Fatalf("missing %s in %s", n, haystack)
		}
		if idx < last {
			t.Fatalf("field %s appears out of order in %s", n, haystack)
		}
		last = idx
	}
}
