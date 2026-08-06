package translate

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/imroc/req/v3"
)

// useHermeticClient pre-seeds the process-wide client cache for the empty
// proxy key so callOneshot reuses it instead of building the real
// ImpersonateChrome client, whose creation also fires a background warmup
// GET to www.deepl.com. The client mirrors production where it matters for
// this layer: it pins Accept-Encoding so Go's transport does not auto-
// decompress, which is exactly why callOneshot decompresses by hand.
func useHermeticClient(t *testing.T) {
	t.Helper()
	c := req.C().
		SetTimeout(5*time.Second).
		SetCommonHeader("Accept-Encoding", "gzip, deflate, br")
	oneshotClients.Store("", c)
}

func TestCallOneshotRequestShaping(t *testing.T) {
	useHermeticClient(t)

	var gotHeader http.Header
	var gotBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Clone()
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"translations":[{"text":"hallo","detected_source_language":"EN"}]}`))
	}))
	defer srv.Close()

	body := []byte(`{"text":["hello"],"target_lang":"de"}`)

	t.Run("anonymous sends Authorization None and extension headers", func(t *testing.T) {
		res, status, err := callOneshot(srv.URL, body, "", "")
		if err != nil {
			t.Fatalf("callOneshot error: %v", err)
		}
		if status != http.StatusOK {
			t.Fatalf("status = %d, want 200", status)
		}
		if got := res.Get("translations.0.text").String(); got != "hallo" {
			t.Errorf("parsed body wrong, translations.0.text = %q", got)
		}
		if h := gotHeader.Get("Authorization"); h != "None" {
			t.Errorf("Authorization = %q, want %q", h, "None")
		}
		if h := gotHeader.Get("Content-Type"); h != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", h)
		}
		if h := gotHeader.Get("Origin"); h != "chrome-extension://"+chromeExtensionID {
			t.Errorf("Origin = %q, want chrome-extension://%s", h, chromeExtensionID)
		}
		for k, want := range map[string]string{
			"Sec-Fetch-Site": "cross-site",
			"Sec-Fetch-Mode": "cors",
			"Sec-Fetch-Dest": "empty",
		} {
			if h := gotHeader.Get(k); h != want {
				t.Errorf("%s = %q, want %q", k, h, want)
			}
		}
		if !bytes.Equal(gotBody, body) {
			t.Errorf("forwarded body = %q, want %q", gotBody, body)
		}
	})

	t.Run("bearer token used when session provided", func(t *testing.T) {
		if _, _, err := callOneshot(srv.URL, body, "tok123", ""); err != nil {
			t.Fatalf("callOneshot error: %v", err)
		}
		if h := gotHeader.Get("Authorization"); h != "Bearer tok123" {
			t.Errorf("Authorization = %q, want %q", h, "Bearer tok123")
		}
	})
}

// TestCallOneshotDecompression covers the three hand-rolled decode paths
// (gzip/deflate/br) plus identity. This code exists only because the client
// pins Accept-Encoding, and it was previously untested — a decode regression
// would surface to users as a blind "Translation failed".
func TestCallOneshotDecompression(t *testing.T) {
	useHermeticClient(t)

	const payload = `{"translations":[{"text":"decoded"}]}`

	cases := []struct {
		name     string
		encoding string
		compress func([]byte) []byte
	}{
		{"identity", "", func(b []byte) []byte { return b }},
		{"gzip", "gzip", func(b []byte) []byte {
			var buf bytes.Buffer
			w := gzip.NewWriter(&buf)
			_, _ = w.Write(b)
			_ = w.Close()
			return buf.Bytes()
		}},
		{"deflate", "deflate", func(b []byte) []byte {
			var buf bytes.Buffer
			w, _ := flate.NewWriter(&buf, flate.DefaultCompression)
			_, _ = w.Write(b)
			_ = w.Close()
			return buf.Bytes()
		}},
		{"brotli", "br", func(b []byte) []byte {
			var buf bytes.Buffer
			w := brotli.NewWriter(&buf)
			_, _ = w.Write(b)
			_ = w.Close()
			return buf.Bytes()
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tc.encoding != "" {
					w.Header().Set("Content-Encoding", tc.encoding)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write(tc.compress([]byte(payload)))
			}))
			defer srv.Close()

			res, status, err := callOneshot(srv.URL, []byte(`{}`), "", "")
			if err != nil {
				t.Fatalf("callOneshot(%s) error: %v", tc.name, err)
			}
			if status != http.StatusOK {
				t.Fatalf("status = %d, want 200", status)
			}
			if got := res.Get("translations.0.text").String(); got != "decoded" {
				t.Errorf("decoded text = %q, want %q (encoding %q)", got, "decoded", tc.encoding)
			}
		})
	}
}
