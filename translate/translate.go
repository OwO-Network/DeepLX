/*
 * @Author: Vincent Young
 * @Date: 2024-09-16 11:59:24
 * @LastEditors: Vincent Yang
 * @LastEditTime: 2025-07-13 23:09:49
 * @FilePath: /DeepLX/translate/translate.go
 * @Telegram: https://t.me/missuo
 * @GitHub: https://github.com/missuo
 *
 * Copyright © 2024 by Vincent, All Rights Reserved.
 */

package translate

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/abadojack/whatlanggo"
	"github.com/imroc/req/v3"

	"github.com/andybalholm/brotli"
	"github.com/tidwall/gjson"
)

// makeRequestWithBody makes an HTTP request with pre-formatted body using minimal headers
func makeRequestWithBody(postStr string, proxyURL string, dlSession string) (gjson.Result, error) {
	urlFull := "https://www2.deepl.com/jsonrpc"

	// Create a new req client
	client := req.C().SetTLSFingerprintRandomized()

	// Set headers to simulate browser request
	headers := http.Header{
		"Content-Type":    []string{"application/json"},
		"Accept":          []string{"*/*"},
		"Accept-Language": []string{"en-US,en;q=0.9"},
		"Accept-Encoding": []string{"gzip, deflate, br, zstd"},
		"Origin":          []string{"https://www.deepl.com"},
		"Referer":         []string{"https://www.deepl.com/"},
		"Sec-Fetch-Dest":  []string{"empty"},
		"Sec-Fetch-Mode":  []string{"cors"},
		"Sec-Fetch-Site":  []string{"same-site"},
		"User-Agent":      []string{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/141.0.0.0 Safari/537.36 Edg/141.0.0.0"},
	}

	if dlSession != "" {
		headers.Set("Cookie", "dl_session="+dlSession)
	}

	// Set proxy if provided
	if proxyURL != "" {
		proxy, err := url.Parse(proxyURL)
		if err != nil {
			return gjson.Result{}, err
		}
		client.SetProxyURL(proxy.String())
	}

	// Make the request
	r := client.R()
	r.Headers = headers
	resp, err := r.
		SetBody(bytes.NewReader([]byte(postStr))).
		Post(urlFull)

	if err != nil {
		return gjson.Result{}, err
	}

	// Check for blocked status like TypeScript version
	if resp.StatusCode == 429 {
		return gjson.Result{}, fmt.Errorf("too many requests, your IP has been blocked by DeepL temporarily, please don't request it frequently in a short time")
	}

	// Check for other error status codes
	if resp.StatusCode != 200 {
		return gjson.Result{}, fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}

	var bodyReader io.Reader
	contentEncoding := resp.Header.Get("Content-Encoding")
	switch contentEncoding {
	case "br":
		bodyReader = brotli.NewReader(resp.Body)
	case "gzip":
		bodyReader, err = gzip.NewReader(resp.Body)
		if err != nil {
			return gjson.Result{}, fmt.Errorf("failed to create gzip reader: %w", err)
		}
	case "deflate":
		bodyReader = flate.NewReader(resp.Body)
	default:
		bodyReader = resp.Body
	}

	body, err := io.ReadAll(bodyReader)
	if err != nil {
		return gjson.Result{}, fmt.Errorf("failed to read response body: %w", err)
	}
	return gjson.ParseBytes(body), nil
}

// TranslateByDeepLX performs translation using DeepL API
func TranslateByDeepLX(sourceLang, targetLang, text string, tagHandling string, proxyURL string, dlSession string) (DeepLXTranslationResult, error) {
	if text == "" {
		return DeepLXTranslationResult{
			Code:    http.StatusNotFound,
			Message: "No text to translate",
		}, nil
	}

	// Get detected language if source language is auto
	if sourceLang == "auto" || sourceLang == "" {
		sourceLang = strings.ToUpper(whatlanggo.DetectLang(text).Iso6391())
	}

	// Prepare translation request using new LMT_handle_texts method
	id := getRandomNumber()
	iCount := getICount(text)
	timestamp := getTimeStamp(iCount)

	postData := &PostData{
		Jsonrpc: "2.0",
		Method:  "LMT_handle_texts",
		ID:      id,
		Params: Params{
			Splitting: "newlines",
			Lang: Lang{
				SourceLangUserSelected: sourceLang,
				TargetLang:             targetLang,
			},
			Texts: []TextItem{{
				Text:                text,
				RequestAlternatives: 3,
			}},
			Timestamp: timestamp,
		},
	}

	// Format and apply body manipulation method like TypeScript
	postStr := formatPostString(postData)
	postStr = handlerBodyMethod(id, postStr)

	// Make translation request
	result, err := makeRequestWithBody(postStr, proxyURL, dlSession)
	if err != nil {
		return DeepLXTranslationResult{
			Code:    http.StatusServiceUnavailable,
			Message: err.Error(),
		}, nil
	}

	// Process translation results using new format
	textsArray := result.Get("result.texts").Array()
	if len(textsArray) == 0 {
		return DeepLXTranslationResult{
			Code:    http.StatusServiceUnavailable,
			Message: "Translation failed",
		}, nil
	}

	// Get main translation
	mainText := textsArray[0].Get("text").String()
	if mainText == "" {
		return DeepLXTranslationResult{
			Code:    http.StatusServiceUnavailable,
			Message: "Translation failed",
		}, nil
	}

	// Get alternatives
	var alternatives []string
	alternativesArray := textsArray[0].Get("alternatives").Array()
	for _, alt := range alternativesArray {
		altText := alt.Get("text").String()
		if altText != "" {
			alternatives = append(alternatives, altText)
		}
	}

	// Get detected source language from response
	detectedLang := result.Get("result.lang").String()
	if detectedLang != "" {
		sourceLang = detectedLang
	}

	return DeepLXTranslationResult{
		Code:         http.StatusOK,
		ID:           id,
		Data:         mainText,
		Alternatives: alternatives,
		SourceLang:   sourceLang,
		TargetLang:   targetLang,
		Method:       map[bool]string{true: "Pro", false: "Free"}[dlSession != ""],
	}, nil
}
