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
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/imroc/req/v3"
	"github.com/tidwall/gjson"
)

// DeepL's web frontend retired LMT_handle_jobs/LMT_handle_texts on www2.deepl.com
// for the interactive translator (now a SignalR/WebSocket channel). The browser
// extension and iOS app still use a stateless REST endpoint called "oneshot",
// which is what we target here. It accepts anonymous traffic with a literal
// `Authorization: None` header and lives on a separate rate-limit pool from
// the JSON-RPC backends, so it is far less prone to "Too many requests" 429s.
const (
	oneshotFreeEndpoint = "https://oneshot-free.www.deepl.com/v1/translate"
	oneshotProEndpoint  = "https://oneshot-pro.www.deepl.com/v1/translate"
)

// oneshot uses lowercase, BCP-47-ish language codes (de, en-US, zh-Hans).
// Callers historically pass DeepL's uppercase codes (DE, EN, ZH) — translate
// them here. Unknown codes fall through lowercased so future additions still
// work without a code change.
var langCodeToOneshot = map[string]string{
	"AR": "ar", "BG": "bg", "CS": "cs", "DA": "da", "DE": "de", "EL": "el",
	"EN": "en-US", "EN-GB": "en-GB", "EN-US": "en-US",
	"ES": "es", "ET": "et", "FI": "fi", "FR": "fr", "HU": "hu",
	"ID": "id", "IT": "it", "JA": "ja", "KO": "ko", "LT": "lt", "LV": "lv",
	"NB": "nb", "NL": "nl", "PL": "pl",
	"PT": "pt-BR", "PT-BR": "pt-BR", "PT-PT": "pt-PT",
	"RO": "ro", "RU": "ru", "SK": "sk", "SL": "sl", "SV": "sv",
	"TR": "tr", "UK": "uk",
	"ZH": "zh-Hans", "ZH-HANS": "zh-Hans", "ZH-HANT": "zh-Hant",
}

func toOneshotLang(code string) string {
	if v, ok := langCodeToOneshot[strings.ToUpper(code)]; ok {
		return v
	}
	return strings.ToLower(code)
}

// oneshotResponse is the JSON shape returned by /v1/translate.
type oneshotResponse struct {
	Translations []struct {
		DetectedSourceLanguage string `json:"detected_source_language"`
		Text                   string `json:"text"`
	} `json:"translations"`
}

// callOneshot POSTs the prepared body and returns the parsed JSON.
// `bearerToken` is empty for anonymous (free) requests, in which case the
// extension sends the literal string "None" — replicate that exactly, because
// omitting the header changes the server's auth-handling branch.
func callOneshot(endpoint string, body []byte, bearerToken, proxyURL string) (gjson.Result, int, error) {
	client := req.C().ImpersonateChrome()
	if proxyURL != "" {
		proxy, err := url.Parse(proxyURL)
		if err != nil {
			return gjson.Result{}, 0, err
		}
		client.SetProxyURL(proxy.String())
	}

	authValue := "None"
	if bearerToken != "" {
		authValue = "Bearer " + bearerToken
	}

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "*/*").
		SetHeader("Authorization", authValue).
		SetHeader("Origin", "https://www.deepl.com").
		SetHeader("Referer", "https://www.deepl.com/").
		SetHeader("Sec-Fetch-Site", "same-site").
		SetHeader("Sec-Fetch-Mode", "cors").
		SetHeader("Sec-Fetch-Dest", "empty").
		SetBody(bytes.NewReader(body)).
		Post(endpoint)
	if err != nil {
		return gjson.Result{}, 0, err
	}

	raw, err := resp.ToBytes()
	if err != nil {
		return gjson.Result{}, resp.StatusCode, fmt.Errorf("failed to read response body: %w", err)
	}
	return gjson.ParseBytes(raw), resp.StatusCode, nil
}

// TranslateByDeepLX performs translation via DeepL's oneshot endpoint.
// Passing dlSession switches to the Pro endpoint; the value is sent verbatim
// as the Bearer token, so callers must supply an OAuth access token (not the
// legacy `dl_session` cookie) when using Pro.
func TranslateByDeepLX(sourceLang, targetLang, text string, tagHandling string, proxyURL string, dlSession string) (DeepLXTranslationResult, error) {
	if text == "" {
		return DeepLXTranslationResult{
			Code:    http.StatusNotFound,
			Message: "No text to translate",
		}, nil
	}

	reqBody := map[string]any{
		"text":        []string{text},
		"target_lang": toOneshotLang(targetLang),
	}
	if sourceLang != "" && !strings.EqualFold(sourceLang, "auto") {
		reqBody["source_lang"] = toOneshotLang(sourceLang)
	}
	bodyBytes, _ := json.Marshal(reqBody)

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

	detected := translations[0].Get("detected_source_language").String()
	if detected != "" {
		// Normalize back to DeepL-style uppercase for response continuity.
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
