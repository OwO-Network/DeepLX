/*
 * @Author: Vincent Young
 * @Date: 2024-09-16 11:59:24
 * @LastEditors: Vincent Yang
 * @LastEditTime: 2026-05-22 00:00:00
 * @FilePath: /DeepLX/translate/translate.go
 * @Telegram: https://t.me/missuo
 * @GitHub: https://github.com/missuo
 *
 * Copyright © 2024 by Vincent, All Rights Reserved.
 */

package translate

import (
	"compress/flate"
	"compress/gzip"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/imroc/req/v3"
	"github.com/tidwall/gjson"
)

// DeepL's interactive web translator migrated to a SignalR/WebSocket
// channel and the legacy LMT_handle_texts backend on www2.deepl.com now
// 429s anonymous traffic within a handful of calls. The official Chrome
// extension instead POSTs to a stateless "oneshot" endpoint that lives
// on a separate rate-limit pool and accepts the literal header
// `Authorization: None` for anonymous requests — that is what we target.
//
// The request we send is reverse-engineered from the extension's
// background.js (Chrome Web Store ID cofdbpoegempjloogbagkncekinflcnj):
//   - URL builder   → mN() at ~offset 529948
//   - body builder  → IN() at ~offset 531200
//   - fetch wrapper → JO() at ~offset 508659
//   - app metadata  → Wo() at ~offset 16500
const (
	oneshotFreeEndpoint = "https://oneshot-free.www.deepl.com/v1/translate"
	oneshotProEndpoint  = "https://oneshot-pro.www.deepl.com/v1/translate"

	// Pinned to the Chrome version utls bundles into req v3 (HelloChrome_120).
	// Keep this in lockstep with the user-agent and app_information.os_version
	// so the TLS handshake, UA, and self-reported browser version all agree —
	// a mismatch on any one of those is a cheap signal for the WAF.
	impersonatedChromeMajor = "120"
	chromeExtensionVersion  = "1.86.0"
	chromeExtensionID       = "cofdbpoegempjloogbagkncekinflcnj"
)

// instanceID mirrors the UUID the extension persists in chrome.storage on
// install: stable for the life of the process, reused on every request.
// Rotating it per-request would be a far stronger signal than reusing one.
var instanceID = newInstanceID()

// A real extension fetch() inherits whatever cookies the browser has
// accumulated on .deepl.com. A cold visit to www.deepl.com sets
// userCountry=<iso2> and verifiedBot=false; users who have ever opened
// the site additionally have _ga / _ga_<id> from analytics JS. We share
// a process-wide cookie jar so every oneshot POST automatically carries
// whatever the warmup GET picked up.
var (
	cookieJar     http.CookieJar
	cookieJarOnce sync.Once
	cookieWarmer  sync.Once
)

func sharedCookieJar() http.CookieJar {
	cookieJarOnce.Do(func() {
		j, _ := cookiejar.New(nil)
		cookieJar = j
	})
	return cookieJar
}

// warmCookies primes the shared jar by GETting www.deepl.com once.
// The Set-Cookie response (userCountry / verifiedBot) lands on .deepl.com,
// which is the eTLD+1 of oneshot-free.www.deepl.com, so subsequent POSTs
// to the oneshot endpoint will carry those cookies automatically.
func warmCookies(client *req.Client) {
	cookieWarmer.Do(func() {
		_, _ = client.R().Get("https://www.deepl.com/translator")
	})
}

func newInstanceID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "00000000-0000-4000-8000-000000000000"
	}
	b[6] = (b[6] & 0x0f) | 0x40 // RFC 4122 v4
	b[8] = (b[8] & 0x3f) | 0x80
	s := hex.EncodeToString(b)
	return fmt.Sprintf("%s-%s-%s-%s-%s", s[0:8], s[8:12], s[12:16], s[16:20], s[20:32])
}

// Language code tables mirror the bundled list in the extension's
// background.js (arrays `y` ~offset 6000 for the full target-capable
// set, `A` for source-only aliases). Keys are the uppercase forms
// callers pass; values are the lowercase BCP-47-ish forms the oneshot
// endpoint expects ("de", "en-US", "zh-Hans", ...).
//
// targetLangMap is what the API accepts as `target_lang`. EN and PT
// are intentionally absent — DeepL deprecated them as target codes in
// favour of EN-US/EN-GB and PT-BR/PT-PT, and the extension's y array
// reflects that. We accept EN/PT as a backward-compat convenience and
// resolve them to the regional default (en-US, pt-BR).
var targetLangMap = map[string]string{
	"AR": "ar", "BG": "bg", "CS": "cs", "DA": "da", "DE": "de", "EL": "el",
	"EN-GB": "en-GB", "EN-US": "en-US",
	"ES": "es", "ES-419": "es-419", "ET": "et", "FI": "fi", "FR": "fr",
	"HE": "he", "HU": "hu", "ID": "id", "IT": "it", "JA": "ja", "KO": "ko",
	"LT": "lt", "LV": "lv", "NB": "nb", "NL": "nl", "PL": "pl",
	"PT-BR": "pt-BR", "PT-PT": "pt-PT",
	"RO": "ro", "RU": "ru", "SK": "sk", "SL": "sl", "SV": "sv",
	"TR": "tr", "UK": "uk", "VI": "vi",
	"ZH": "zh-Hans", "ZH-HANS": "zh-Hans", "ZH-HANT": "zh-Hant",
	// Convenience aliases for legacy callers.
	"EN": "en-US",
	"PT": "pt-BR",
}

// sourceLangMap is what the API accepts as `source_lang`. It is a
// superset of targetLangMap: EN and PT are first-class source codes
// (extension array `A`) mapping to the generic "en"/"pt" — used when
// the caller knows the input is English/Portuguese but does not want
// to commit to a regional variant.
var sourceLangMap = func() map[string]string {
	m := make(map[string]string, len(targetLangMap)+2)
	for k, v := range targetLangMap {
		m[k] = v
	}
	m["EN"] = "en"
	m["PT"] = "pt"
	return m
}()

// resolveTargetLang validates and normalizes a user-supplied target
// language code. Returns "" and a non-nil error if the code is empty,
// "auto", or otherwise not in the supported set.
func resolveTargetLang(code string) (string, error) {
	if code == "" {
		return "", fmt.Errorf("target_lang is required")
	}
	if strings.EqualFold(code, "auto") {
		return "", fmt.Errorf("target_lang cannot be \"auto\"; pick one of: %s", supportedTargetLangsList())
	}
	if v, ok := targetLangMap[strings.ToUpper(code)]; ok {
		return v, nil
	}
	return "", fmt.Errorf("unsupported target_lang %q; valid codes: %s", code, supportedTargetLangsList())
}

// resolveSourceLang validates and normalizes a user-supplied source
// language code. An empty string or "auto" is allowed and returns
// ("", nil) so the caller omits source_lang and lets the server
// autodetect.
func resolveSourceLang(code string) (string, error) {
	if code == "" || strings.EqualFold(code, "auto") {
		return "", nil
	}
	if v, ok := sourceLangMap[strings.ToUpper(code)]; ok {
		return v, nil
	}
	return "", fmt.Errorf("unsupported source_lang %q; valid codes: %s (or \"auto\")", code, supportedSourceLangsList())
}

// supportedTargetLangsList / supportedSourceLangsList return a sorted,
// comma-separated rendering of the supported codes for use in error
// messages. Cached at first call.
var (
	targetLangsListOnce sync.Once
	targetLangsList     string
	sourceLangsListOnce sync.Once
	sourceLangsList     string
)

func supportedTargetLangsList() string {
	targetLangsListOnce.Do(func() {
		targetLangsList = sortedKeys(targetLangMap)
	})
	return targetLangsList
}

func supportedSourceLangsList() string {
	sourceLangsListOnce.Do(func() {
		sourceLangsList = sortedKeys(sourceLangMap)
	})
	return sourceLangsList
}

func sortedKeys(m map[string]string) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return strings.Join(keys, ", ")
}

// appInformation matches the snake_case shape produced by background.js
// Wo({isSnakeCase: true}). Values are pinned to the same Chrome version
// as the TLS handshake so the request tells one consistent story.
type appInformation struct {
	OS         string `json:"os"`
	OSVersion  string `json:"os_version"`
	AppVersion string `json:"app_version"`
	AppBuild   string `json:"app_build"`
	InstanceID string `json:"instance_id"`
}

// oneshotRequest mirrors the body assembled in background.js IN(...).
// Field order matches the extension's object literal so the serialized
// JSON is byte-identical (encoding/json honours struct field order).
type oneshotRequest struct {
	Text           []string       `json:"text"`
	TargetLang     string         `json:"target_lang"`
	SourceLang     string         `json:"source_lang,omitempty"`
	UsageType      string         `json:"usage_type"`
	AppInformation appInformation `json:"app_information"`
}

// newOneshotClient configures a req.Client whose outbound profile matches
// a chrome-extension service-worker fetch() byte-for-byte where it can.
// ImpersonateChrome gives us the Chrome 120 TLS ClientHello, HTTP/2
// SETTINGS, pseudo/header order, and a sec-ch-ua/user-agent set tied to
// the same version. It also installs a navigation-flavoured set of common
// headers (pragma, cache-control, upgrade-insecure-requests, sec-fetch-user)
// that a fetch() never emits — wipe those so the WAF cannot tell us apart
// on that axis.
func newOneshotClient(proxyURL string) (*req.Client, error) {
	client := req.C().ImpersonateChrome().SetCookieJar(sharedCookieJar())
	for _, h := range []string{
		"Pragma",
		"Cache-Control",
		"Upgrade-Insecure-Requests",
		"Sec-Fetch-User",
	} {
		client.Headers.Del(h)
	}
	// Chrome 120 fetch() advertises gzip/deflate/br (zstd only appeared
	// as a default in Chrome 123+). req's default of just "gzip" is a
	// distinguishable signal — match Chrome explicitly.
	client.SetCommonHeader("Accept-Encoding", "gzip, deflate, br")

	if proxyURL != "" {
		u, err := url.Parse(proxyURL)
		if err != nil {
			return nil, err
		}
		client.SetProxyURL(u.String())
	}
	return client, nil
}

// callOneshot POSTs to the oneshot endpoint and returns the parsed JSON.
// For anonymous traffic bearerToken is empty and we send the literal
// header `Authorization: None` — replicating the extension's JO() wrapper
// exactly. Omitting that header instead would put the request on a
// different server-side auth branch.
func callOneshot(endpoint string, body []byte, bearerToken, proxyURL string) (gjson.Result, int, error) {
	client, err := newOneshotClient(proxyURL)
	if err != nil {
		return gjson.Result{}, 0, err
	}
	warmCookies(client) // no-op after the first translation in the process

	authValue := "None"
	if bearerToken != "" {
		authValue = "Bearer " + bearerToken
	}

	resp, err := client.R().
		DisableAutoReadResponse().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "*/*").
		SetHeader("Authorization", authValue).
		SetHeader("Origin", "chrome-extension://"+chromeExtensionID).
		SetHeader("Sec-Fetch-Site", "cross-site").
		SetHeader("Sec-Fetch-Mode", "cors").
		SetHeader("Sec-Fetch-Dest", "empty").
		SetBodyBytes(body). // SetBodyBytes pins Content-Length; using an
		// io.Reader instead forces Transfer-Encoding: chunked, which a
		// real fetch() with JSON.stringify body never emits.
		Post(endpoint)
	if err != nil {
		return gjson.Result{}, 0, err
	}
	defer resp.Body.Close()

	// Once we set Accept-Encoding ourselves, Go's HTTP stack stops
	// transparently decompressing, so handle gzip/deflate/br by hand.
	var reader io.Reader = resp.Body
	switch strings.ToLower(resp.Header.Get("Content-Encoding")) {
	case "gzip":
		gr, err := gzip.NewReader(resp.Body)
		if err != nil {
			return gjson.Result{}, resp.StatusCode, fmt.Errorf("gzip reader: %w", err)
		}
		defer gr.Close()
		reader = gr
	case "deflate":
		reader = flate.NewReader(resp.Body)
	case "br":
		reader = brotli.NewReader(resp.Body)
	}
	raw, err := io.ReadAll(reader)
	if err != nil {
		return gjson.Result{}, resp.StatusCode, fmt.Errorf("read response body: %w", err)
	}
	return gjson.ParseBytes(raw), resp.StatusCode, nil
}

// TranslateByDeepLX performs translation via the DeepL oneshot endpoint.
// Passing dlSession switches to the Pro endpoint; the value is sent
// verbatim as the Bearer token (i.e. it must be an OAuth access token,
// not the legacy dl_session cookie).
func TranslateByDeepLX(sourceLang, targetLang, text string, tagHandling string, proxyURL string, dlSession string) (DeepLXTranslationResult, error) {
	if text == "" {
		return DeepLXTranslationResult{
			Code:    http.StatusNotFound,
			Message: "No text to translate",
		}, nil
	}

	resolvedTarget, err := resolveTargetLang(targetLang)
	if err != nil {
		return DeepLXTranslationResult{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		}, nil
	}
	resolvedSource, err := resolveSourceLang(sourceLang)
	if err != nil {
		return DeepLXTranslationResult{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		}, nil
	}

	reqStruct := oneshotRequest{
		Text:       []string{text},
		TargetLang: resolvedTarget,
		SourceLang: resolvedSource, // empty = autodetect; omitempty drops the field
		UsageType:  "Translate",
		AppInformation: appInformation{
			OS:         "brex_macOS",
			OSVersion:  "brex_chrome_" + impersonatedChromeMajor + ".0.0.0",
			AppVersion: chromeExtensionVersion,
			AppBuild:   "chrome_web_store",
			InstanceID: instanceID,
		},
	}
	bodyBytes, _ := json.Marshal(reqStruct)

	endpoint := oneshotFreeEndpoint
	if dlSession != "" {
		endpoint = oneshotProEndpoint
	}

	id := time.Now().UnixMilli()
	result, status, err := callOneshot(endpoint, bodyBytes, dlSession, proxyURL)
	if err != nil {
		return DeepLXTranslationResult{
			ID:      id,
			Code:    http.StatusServiceUnavailable,
			Message: err.Error(),
		}, nil
	}

	switch status {
	case http.StatusOK:
		// fall through to body parsing
	case http.StatusTooManyRequests:
		return DeepLXTranslationResult{
			ID:      id,
			Code:    http.StatusTooManyRequests,
			Message: "too many requests, your IP has been blocked by DeepL temporarily, please don't request it frequently in a short time",
		}, nil
	default:
		return DeepLXTranslationResult{
			ID:      id,
			Code:    http.StatusServiceUnavailable,
			Message: fmt.Sprintf("request failed with status code: %d", status),
		}, nil
	}

	translations := result.Get("translations").Array()
	if len(translations) == 0 {
		return DeepLXTranslationResult{
			ID:      id,
			Code:    http.StatusServiceUnavailable,
			Message: "Translation failed",
		}, nil
	}

	mainText := translations[0].Get("text").String()
	if mainText == "" {
		return DeepLXTranslationResult{
			ID:      id,
			Code:    http.StatusServiceUnavailable,
			Message: "Translation failed",
		}, nil
	}

	if detected := translations[0].Get("detected_source_language").String(); detected != "" {
		sourceLang = strings.ToUpper(detected)
	}

	return DeepLXTranslationResult{
		Code:         http.StatusOK,
		ID:           id,
		Data:         mainText,
		Alternatives: nil, // oneshot does not return alternatives
		SourceLang:   sourceLang,
		TargetLang:   targetLang,
		Method:       map[bool]string{true: "Pro", false: "Free"}[dlSession != ""],
	}, nil
}
