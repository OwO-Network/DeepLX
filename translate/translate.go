/*
 * @Author: Vincent Young
 * @Date: 2024-09-16 11:59:24
 * @LastEditors: Vincent Yang
 * @LastEditTime: 2024-12-03 11:23:23
 * @FilePath: /DeepLX/translate/translate.go
 * @Telegram: https://t.me/missuo
 * @GitHub: https://github.com/missuo
 *
 * Copyright © 2024 by Vincent, All Rights Reserved.
 */

package translate

import (
	"bytes"
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

const baseURL = "https://www2.deepl.com/jsonrpc"

// makeRequest makes an HTTP request to DeepL API

func makeRequest(postData *PostData, urlMethod string, proxyURL string, dlSession string) (gjson.Result, error) {
	urlFull := fmt.Sprintf("%s?client=chrome-extension,1.28.0&method=%s", baseURL, urlMethod)

	postStr := formatPostString(postData)

	// Create a new req client
	client := req.C().SetTLSFingerprintRandomized()

	// Set headers
	headers := http.Header{
		"Accept":          []string{"*/*"},
		"Accept-Language": []string{"en-US,en;q=0.9,zh-CN;q=0.8,zh-TW;q=0.7,zh-HK;q=0.6,zh;q=0.5"},
		"Authorization":   []string{"None"},
		"Cache-Control":   []string{"no-cache"},
		"Content-Type":    []string{"application/json"},
		"DNT":             []string{"1"},
		"Origin":          []string{"chrome-extension://cofdbpoegempjloogbagkncekinflcnj"},
		"Pragma":          []string{"no-cache"},
		"Priority":        []string{"u=1, i"},
		"Referer":         []string{"https://www.deepl.com/"},
		"Sec-Fetch-Dest":  []string{"empty"},
		"Sec-Fetch-Mode":  []string{"cors"},
		"Sec-Fetch-Site":  []string{"none"},
		"Sec-GPC":         []string{"1"},
		"User-Agent":      []string{"DeepLBrowserExtension/1.28.0 Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0.0.0 Safari/537.36"},
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
	if resp.Header.Get("Content-Encoding") == "br" {
		bodyReader = brotli.NewReader(resp.Body)
	} else {
		bodyReader = resp.Body
	}

	body, err := io.ReadAll(bodyReader)
	if err != nil {
		return gjson.Result{}, err
	}
	return gjson.ParseBytes(body), nil
}

// splitText splits the input text for translation
func splitText(text string, tagHandling bool, proxyURL string, dlSession string) (gjson.Result, error) {
	postData := &PostData{
		Jsonrpc: "2.0",
		Method:  "LMT_split_text",
		ID:      getRandomNumber(),
		Params: Params{
			CommonJobParams: CommonJobParams{
				Mode: "translate",
			},
			Lang: Lang{
				LangUserSelected: "auto",
			},
			Texts:    []string{text},
			TextType: map[bool]string{true: "richtext", false: "plaintext"}[tagHandling || isRichText(text)],
		},
	}

	return makeRequest(postData, "LMT_split_text", proxyURL, dlSession)
}

// TranslateByDeepLX performs translation using DeepL API
func TranslateByDeepLX(sourceLang, targetLang, text string, tagHandling string, proxyURL string, dlSession string) (DeepLXTranslationResult, error) {
	if text == "" {
		return DeepLXTranslationResult{
			Code:    http.StatusNotFound,
			Message: "No text to translate",
		}, nil
	}

	// Split text first
	splitResult, err := splitText(text, tagHandling == "html" || tagHandling == "xml", proxyURL, dlSession)
	if err != nil {
		return DeepLXTranslationResult{
			Code:    http.StatusServiceUnavailable,
			Message: err.Error(),
		}, nil
	}

	// Get detected language if source language is auto
	if sourceLang == "auto" || sourceLang == "" {
		sourceLang = strings.ToUpper(whatlanggo.DetectLang(text).Iso6391())
	}

	// Prepare jobs from split result
	var jobs []Job
	chunks := splitResult.Get("result.texts.0.chunks").Array()
	for idx, chunk := range chunks {
		sentence := chunk.Get("sentences.0")

		// Handle context
		contextBefore := []string{}
		contextAfter := []string{}
		if idx > 0 {
			contextBefore = []string{chunks[idx-1].Get("sentences.0.text").String()}
		}
		if idx < len(chunks)-1 {
			contextAfter = []string{chunks[idx+1].Get("sentences.0.text").String()}
		}

		jobs = append(jobs, Job{
			Kind:               "default",
			PreferredNumBeams:  4,
			RawEnContextBefore: contextBefore,
			RawEnContextAfter:  contextAfter,
			Sentences: []Sentence{{
				Prefix: sentence.Get("prefix").String(),
				Text:   sentence.Get("text").String(),
				ID:     idx + 1,
			}},
		})
	}

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
				Mode: "translate",
			},
			Lang: Lang{
				SourceLangComputed: strings.ToUpper(sourceLang),
				TargetLang:         strings.ToUpper(targetLangCode),
			},
			Jobs:      jobs,
			Priority:  1,
			Timestamp: getTimeStamp(getICount(text)),
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
					RegionalVariant: map[bool]string{true: targetLang, false: ""}[hasRegionalVariant],
				},
				Lang: Lang{
					SourceLangComputed: strings.ToUpper(sourceLang),
					TargetLang:         strings.ToUpper(targetLangCode),
				},
				Jobs:      jobs,
				Priority:  1,
				Timestamp: getTimeStamp(getICount(text)),
			},
		}
	}

	// Make translation request
	result, err := makeRequest(postData, "LMT_handle_jobs", proxyURL, dlSession)
	if err != nil {
		return DeepLXTranslationResult{
			Code:    http.StatusServiceUnavailable,
			Message: err.Error(),
		}, nil
	}

	// Process translation results
	var alternatives []string
	var translatedText string

	translations := result.Get("result.translations").Array()
	if len(translations) > 0 {
		// Get alternatives
		numBeams := len(translations[0].Get("beams").Array())
		for i := 0; i < numBeams; i++ {
			var altText string
			for _, translation := range translations {
				beams := translation.Get("beams").Array()
				if i < len(beams) {
					altText += beams[i].Get("sentences.0.text").String()
				}
			}
			if altText != "" {
				alternatives = append(alternatives, altText)
			}
		}

		// Get main translation
		for _, translation := range translations {
			translatedText += translation.Get("beams.0.sentences.0.text").String() + " "
		}
		translatedText = strings.TrimSpace(translatedText)
	}

	if translatedText == "" {
		return DeepLXTranslationResult{
			Code:    http.StatusServiceUnavailable,
			Message: "Translation failed",
		}, nil
	}

	return DeepLXTranslationResult{
		Code:         http.StatusOK,
		ID:           id,
		Data:         translatedText,
		Alternatives: alternatives,
		SourceLang:   sourceLang,
		TargetLang:   targetLang,
		Method:       map[bool]string{true: "Pro", false: "Free"}[dlSession != ""],
	}, nil
}
