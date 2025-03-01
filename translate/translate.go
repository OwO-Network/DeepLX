/*
 * @Author: Vincent Young
 * @Date: 2024-09-16 11:59:24
 * @LastEditors: Vincent Yang
 * @LastEditTime: 2025-03-01 04:23:49
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

// makeRequest makes an HTTP request to DeepL API
func makeRequest(postData *PostData, proxyURL string, dlSession string) (gjson.Result, error) {
	urlFull := "https://www2.deepl.com/jsonrpc"
	postStr := formatPostString(postData)

	// Create a new req client
	client := req.C().SetTLSFingerprintRandomized()

	// Set headers
	headers := http.Header{
		"Content-Type":     []string{"application/json"},
		"User-Agent":       []string{"DeepL/1627620 CFNetwork/3826.500.62.2.1 Darwin/24.4.0"},
		"Accept":           []string{"*/*"},
		"X-App-Os-Name":    []string{"iOS"},
		"X-App-Os-Version": []string{"18.4.0"},
		"Accept-Language":  []string{"en-US,en;q=0.9"},
		"Accept-Encoding":  []string{"gzip, deflate, br"}, // Keep this!
		"X-App-Device":     []string{"iPhone16,2"},
		"Referer":          []string{"https://www.deepl.com/"},
		"X-Product":        []string{"translator"},
		"X-App-Build":      []string{"1627620"},
		"X-App-Version":    []string{"25.1"},
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

	var bodyReader io.Reader
	contentEncoding := resp.Header.Get("Content-Encoding")
	switch contentEncoding {
	case "br":
		bodyReader = brotli.NewReader(resp.Body)
	case "gzip":
		bodyReader, err = gzip.NewReader(resp.Body) // Use gzip.NewReader
		if err != nil {
			return gjson.Result{}, fmt.Errorf("failed to create gzip reader: %w", err)
		}
	case "deflate": // Less common, but good to handle
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

	if tagHandling == "" {
		tagHandling = "plaintext"
	}

	// Split text by newlines and store them for later reconstruction
	textParts := strings.Split(text, "\n")
	var translatedParts []string
	var allAlternatives [][]string // Store alternatives for each part

	for _, part := range textParts {
		if strings.TrimSpace(part) == "" {
			translatedParts = append(translatedParts, "")
			allAlternatives = append(allAlternatives, []string{""})
			continue
		}

		// Get detected language if source language is auto
		if sourceLang == "auto" || sourceLang == "" {
			sourceLang = strings.ToUpper(whatlanggo.DetectLang(part).Iso6391())
		}

		// Prepare jobs from split result
		var jobs []Job

		jobs = append(jobs, Job{
			Kind:               "default",
			PreferredNumBeams:  4,
			RawEnContextBefore: []string{},
			RawEnContextAfter:  []string{},
			Sentences: []Sentence{{
				Prefix: "",
				Text:   text,
				ID:     0,
			}},
		})

		hasRegionalVariant := false
		targetLangCode := targetLang
		targetLangParts := strings.Split(targetLang, "-")
		if len(targetLangParts) > 1 {
			targetLangCode = targetLangParts[0]
			hasRegionalVariant = true
		}

		// Prepare translation request
		id := getRandomNumber()

		postData := &PostData{
			Jsonrpc: "2.0",
			Method:  "LMT_handle_jobs",
			ID:      id,
			Params: Params{
				CommonJobParams: CommonJobParams{
					Mode:         "translate",
					Formality:    "undefined",
					TranscribeAs: "romanize",
					AdvancedMode: false,
					TextType:     tagHandling,
					WasSpoken:    false,
				},
				Lang: Lang{
					SourceLangUserSelected: "auto",
					TargetLang:             strings.ToUpper(targetLangCode),
					SourceLangComputed:     strings.ToUpper(sourceLang),
				},
				Jobs:      jobs,
				Timestamp: getTimeStamp(getICount(part)),
			},
		}

		if hasRegionalVariant {
			postData = &PostData{
				Jsonrpc: "2.0",
				Method:  "LMT_handle_jobs",
				ID:      id,
				Params: Params{
					CommonJobParams: CommonJobParams{
						Mode:            "translate",
						Formality:       "undefined",
						TranscribeAs:    "romanize",
						AdvancedMode:    false,
						TextType:        tagHandling,
						WasSpoken:       false,
						RegionalVariant: targetLang,
					},
					Lang: Lang{
						SourceLangUserSelected: "auto",
						TargetLang:             strings.ToUpper(targetLangCode),
						SourceLangComputed:     strings.ToUpper(sourceLang),
					},
					Jobs:      jobs,
					Timestamp: getTimeStamp(getICount(part)),
				},
			}
		}

		// Make translation request
		result, err := makeRequest(postData, proxyURL, dlSession)
		if err != nil {
			return DeepLXTranslationResult{
				Code:    http.StatusServiceUnavailable,
				Message: err.Error(),
			}, nil
		}

		// Process translation results
		var partTranslation string
		var partAlternatives []string

		translations := result.Get("result.translations").Array()
		if len(translations) > 0 {
			// Process main translation
			for _, translation := range translations {
				partTranslation += translation.Get("beams.0.sentences.0.text").String() + " "
			}
			partTranslation = strings.TrimSpace(partTranslation)

			// Process alternatives
			numBeams := len(translations[0].Get("beams").Array())
			for i := 1; i < numBeams; i++ { // Start from 1 since 0 is the main translation
				var altText string
				for _, translation := range translations {
					beams := translation.Get("beams").Array()
					if i < len(beams) {
						altText += beams[i].Get("sentences.0.text").String() + " "
					}
				}
				if altText != "" {
					partAlternatives = append(partAlternatives, strings.TrimSpace(altText))
				}
			}
		}

		if partTranslation == "" {
			return DeepLXTranslationResult{
				Code:    http.StatusServiceUnavailable,
				Message: "Translation failed",
			}, nil
		}

		translatedParts = append(translatedParts, partTranslation)
		allAlternatives = append(allAlternatives, partAlternatives)
	}

	// Join all translated parts with newlines
	translatedText := strings.Join(translatedParts, "\n")

	// Combine alternatives with proper newline handling
	var combinedAlternatives []string
	maxAlts := 0
	for _, alts := range allAlternatives {
		if len(alts) > maxAlts {
			maxAlts = len(alts)
		}
	}

	// Create combined alternatives preserving line structure
	for i := 0; i < maxAlts; i++ {
		var altParts []string
		for j, alts := range allAlternatives {
			if i < len(alts) {
				altParts = append(altParts, alts[i])
			} else if len(translatedParts[j]) == 0 {
				altParts = append(altParts, "") // Keep empty lines
			} else {
				altParts = append(altParts, translatedParts[j]) // Use main translation if no alternative
			}
		}
		combinedAlternatives = append(combinedAlternatives, strings.Join(altParts, "\n"))
	}

	return DeepLXTranslationResult{
		Code:         http.StatusOK,
		ID:           getRandomNumber(), // Using new ID for the complete translation
		Data:         translatedText,
		Alternatives: combinedAlternatives,
		SourceLang:   sourceLang,
		TargetLang:   targetLang,
		Method:       map[bool]string{true: "Pro", false: "Free"}[dlSession != ""],
	}, nil
}
