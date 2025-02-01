/*
 * @Author: Vincent Young
 * @Date: 2024-09-16 11:59:24
 * @LastEditors: Vincent Yang
 * @LastEditTime: 2025-02-01 03:21:41
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
	"sync"

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

	// Split text by newlines
	textParts := strings.Split(text, "\n")

	// Create channels for results
	type translationResult struct {
		index        int
		translation  string
		alternatives []string
		err          error
	}
	results := make(chan translationResult, len(textParts))

	// Create a wait group to track all goroutines
	var wg sync.WaitGroup

	// Launch goroutines for each text part
	for i := range textParts {
		wg.Add(1)
		go func(index int, text string) {
			defer wg.Done()

			if strings.TrimSpace(text) == "" {
				results <- translationResult{
					index:        index,
					translation:  "",
					alternatives: []string{""},
				}
				return
			}

			// Split text first
			splitResult, err := splitText(text, tagHandling == "html" || tagHandling == "xml", proxyURL, dlSession)
			if err != nil {
				results <- translationResult{index: index, err: err}
				return
			}

			// Get detected language if source language is auto
			currentSourceLang := sourceLang
			if currentSourceLang == "auto" || currentSourceLang == "" {
				currentSourceLang = strings.ToUpper(whatlanggo.DetectLang(text).Iso6391())
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
						Mode:            "translate",
						RegionalVariant: map[bool]string{true: targetLang, false: ""}[hasRegionalVariant],
					},
					Lang: Lang{
						SourceLangComputed: strings.ToUpper(currentSourceLang),
						TargetLang:         strings.ToUpper(targetLangCode),
					},
					Jobs:      jobs,
					Priority:  1,
					Timestamp: getTimeStamp(getICount(text)),
				},
			}

			// Make translation request
			result, err := makeRequest(postData, "LMT_handle_jobs", proxyURL, dlSession)
			if err != nil {
				results <- translationResult{index: index, err: err}
				return
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
				for i := 1; i < numBeams; i++ {
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
				results <- translationResult{index: index, err: fmt.Errorf("translation failed")}
				return
			}

			results <- translationResult{
				index:        index,
				translation:  partTranslation,
				alternatives: partAlternatives,
			}
		}(i, textParts[i])
	}

	// Close results channel when all goroutines are done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results maintaining original order
	translatedParts := make([]string, len(textParts))
	allAlternatives := make([][]string, len(textParts))

	for result := range results {
		if result.err != nil {
			return DeepLXTranslationResult{
				Code:    http.StatusServiceUnavailable,
				Message: result.err.Error(),
			}, nil
		}
		translatedParts[result.index] = result.translation
		allAlternatives[result.index] = result.alternatives
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
				altParts = append(altParts, "")
			} else {
				altParts = append(altParts, translatedParts[j])
			}
		}
		combinedAlternatives = append(combinedAlternatives, strings.Join(altParts, "\n"))
	}

	return DeepLXTranslationResult{
		Code:         http.StatusOK,
		ID:           getRandomNumber(),
		Data:         translatedText,
		Alternatives: combinedAlternatives,
		SourceLang:   sourceLang,
		TargetLang:   targetLang,
		Method:       map[bool]string{true: "Pro", false: "Free"}[dlSession != ""],
	}, nil
}
