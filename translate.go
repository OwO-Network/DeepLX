package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/abadojack/whatlanggo"
	"github.com/andybalholm/brotli"
	"github.com/tidwall/gjson"
)

func initDeepLXData(sourceLang string, targetLang string) *PostData {
	hasRegionalVariant := false
	targetLangParts := strings.Split(targetLang, "-")

	// targetLang can be "en", "pt", "pt-PT", "pt-BR"
	// targetLangCode is the first part of the targetLang, e.g. "pt" in "pt-PT"
	targetLangCode := targetLangParts[0]
	if len(targetLangParts) > 1 {
		hasRegionalVariant = true
	}

	commonJobParams := CommonJobParams{
		WasSpoken:    false,
		TranscribeAS: "",
	}
	if hasRegionalVariant {
		commonJobParams.RegionalVariant = targetLang
	}

	return &PostData{
		Jsonrpc: "2.0",
		Method:  "LMT_handle_texts",
		Params: Params{
			Splitting: "newlines",
			Lang: Lang{
				SourceLangUserSelected: sourceLang,
				TargetLang:             targetLangCode,
			},
			CommonJobParams: commonJobParams,
		},
	}
}

func translateByOfficialAPI(text string, sourceLang string, targetLang string, authKey string, proxyURL string) (string, error) {
	freeURL := "https://api-free.deepl.com/v2/translate"
	textArray := strings.Split(text, "\n")

	payload := PayloadAPI{
		Text:       textArray,
		TargetLang: targetLang,
		SourceLang: sourceLang,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", freeURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "DeepL-Auth-Key "+authKey)
	req.Header.Set("Content-Type", "application/json")

	var client *http.Client
	if proxyURL != "" {
		proxy, err := url.Parse(proxyURL)
		if err != nil {
			return "", err
		}
		transport := &http.Transport{
			Proxy: http.ProxyURL(proxy),
		}
		client = &http.Client{Transport: transport}
	} else {
		client = &http.Client{}
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Parsing the response
	var translationResponse TranslationResponse
	err = json.Unmarshal(body, &translationResponse)
	if err != nil {
		return "", err
	}

	// Concatenating the translations
	var sb strings.Builder
	for _, translation := range translationResponse.Translations {
		sb.WriteString(translation.Text)
	}

	return sb.String(), nil
}

func translateByDeepLX(sourceLang string, targetLang string, translateText string, tagHandling string, authKey string, proxyURL string) (DeepLXTranslationResult, error) {
	id := getRandomNumber()
	if sourceLang == "" {
		lang := whatlanggo.DetectLang(translateText)
		deepLLang := strings.ToUpper(lang.Iso6391())
		sourceLang = deepLLang
	}
	// If target language is not specified, set it to English
	if targetLang == "" {
		targetLang = "EN"
	}
	// Handling empty translation text
	if translateText == "" {
		return DeepLXTranslationResult{
			Code:    http.StatusNotFound,
			Message: "No text to translate",
		}, nil
	}

	// Preparing the request data for the DeepL API
	www2URL := "https://www2.deepl.com/jsonrpc"
	id = id + 1
	postData := initDeepLXData(sourceLang, targetLang)
	text := Text{
		Text:                translateText,
		RequestAlternatives: 3,
	}
	postData.ID = id
	postData.Params.Texts = append(postData.Params.Texts, text)
	postData.Params.Timestamp = getTimeStamp(getICount(translateText))

	if tagHandling == "html" || tagHandling == "xml" {
		postData.Params.TagHandling = tagHandling
	}

	// Marshalling the request data to JSON and making necessary string replacements
	post_byte, _ := json.Marshal(postData)
	postStr := string(post_byte)

	// Adding spaces to the JSON string based on the ID to adhere to DeepL's request formatting rules
	if (id+5)%29 == 0 || (id+3)%13 == 0 {
		postStr = strings.Replace(postStr, "\"method\":\"", "\"method\" : \"", -1)
	} else {
		postStr = strings.Replace(postStr, "\"method\":\"", "\"method\": \"", -1)
	}

	// Creating a new HTTP POST request with the JSON data as the body
	post_byte = []byte(postStr)
	reader := bytes.NewReader(post_byte)
	request, err := http.NewRequest("POST", www2URL, reader)

	if err != nil {
		log.Println(err)
		return DeepLXTranslationResult{
			Code:    http.StatusServiceUnavailable,
			Message: "Post request failed",
		}, nil
	}

	// Setting HTTP headers to mimic a request from the DeepL iOS App
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "*/*")
	request.Header.Set("x-app-os-name", "iOS")
	request.Header.Set("x-app-os-version", "16.3.0")
	request.Header.Set("Accept-Language", "en-US,en;q=0.9")
	request.Header.Set("Accept-Encoding", "gzip, deflate, br")
	request.Header.Set("x-app-device", "iPhone13,2")
	request.Header.Set("User-Agent", "DeepL-iOS/2.9.1 iOS 16.3.0 (iPhone13,2)")
	request.Header.Set("x-app-build", "510265")
	request.Header.Set("x-app-version", "2.9.1")
	request.Header.Set("Connection", "keep-alive")

	// Making the HTTP request to the DeepL API
	var client *http.Client
	if proxyURL != "" {
		proxy, err := url.Parse(proxyURL)
		if err != nil {
			return DeepLXTranslationResult{
				Code:    http.StatusServiceUnavailable,
				Message: "Uknown error",
			}, nil
		}
		transport := &http.Transport{
			Proxy: http.ProxyURL(proxy),
		}
		client = &http.Client{Transport: transport}
	} else {
		client = &http.Client{}
	}

	resp, err := client.Do(request)
	if err != nil {
		log.Println(err)
		return DeepLXTranslationResult{
			Code:    http.StatusServiceUnavailable,
			Message: "DeepL API request failed",
		}, nil
	}
	defer resp.Body.Close()

	// Handling potential Brotli compressed response body
	var bodyReader io.Reader
	switch resp.Header.Get("Content-Encoding") {
	case "br":
		bodyReader = brotli.NewReader(resp.Body)
	default:
		bodyReader = resp.Body
	}

	// Reading the response body and parsing it with gjson
	body, _ := io.ReadAll(bodyReader)
	// body, _ := io.ReadAll(resp.Body)
	res := gjson.ParseBytes(body)

	// Handling various response statuses and potential errors
	if res.Get("error.code").String() == "-32600" {
		log.Println(res.Get("error").String())
		return DeepLXTranslationResult{
			Code:    http.StatusNotAcceptable,
			Message: "Invalid target language",
		}, nil
	}

	if resp.StatusCode == http.StatusTooManyRequests && authKey != "" {
		authKeyArray := strings.Split(authKey, ",")
		for _, authKey := range authKeyArray {
			validity, err := checkUsageAuthKey(authKey)
			if err != nil {
				continue
			} else {
				if validity {
					translatedText, err := translateByOfficialAPI(translateText, sourceLang, targetLang, authKey, proxyURL)
					if err != nil {
						return DeepLXTranslationResult{
							Code:    http.StatusTooManyRequests,
							Message: "Too Many Requests",
						}, nil
					}
					return DeepLXTranslationResult{
						Code:       http.StatusOK,
						Message:    "Success",
						ID:         1000000,
						Data:       translatedText,
						SourceLang: sourceLang,
						TargetLang: targetLang,
						Method:     "Official API",
					}, nil
				}
			}

		}
	} else {
		var alternatives []string
		res.Get("result.texts.0.alternatives").ForEach(func(key, value gjson.Result) bool {
			alternatives = append(alternatives, value.Get("text").String())
			return true
		})
		if res.Get("result.texts.0.text").String() == "" {
			return DeepLXTranslationResult{
				Code:    http.StatusServiceUnavailable,
				Message: "Translation failed, API returns an empty result.",
			}, nil
		} else {
			return DeepLXTranslationResult{
				Code:         http.StatusOK,
				ID:           id,
				Message:      "Success",
				Data:         res.Get("result.texts.0.text").String(),
				Alternatives: alternatives,
				SourceLang:   sourceLang,
				TargetLang:   targetLang,
				Method:       "Free",
			}, nil
		}
	}
	return DeepLXTranslationResult{
		Code:    http.StatusServiceUnavailable,
		Message: "Uknown error",
	}, nil
}

func translateByDeepLXPro(sourceLang string, targetLang string, translateText string, tagHandling string, dlSession string, proxyURL string) (DeepLXTranslationResult, error) {
	id := getRandomNumber()
	if sourceLang == "" {
		lang := whatlanggo.DetectLang(translateText)
		deepLLang := strings.ToUpper(lang.Iso6391())
		sourceLang = deepLLang
	}
	// If target language is not specified, set it to English
	if targetLang == "" {
		targetLang = "EN"
	}
	// Handling empty translation text
	if translateText == "" {
		return DeepLXTranslationResult{
			Code:    http.StatusNotFound,
			Message: "No text to translate",
		}, nil
	}

	// Preparing the request data for the DeepL API
	proURL := "https://api.deepl.com/jsonrpc"
	id = id + 1
	postData := initDeepLXData(sourceLang, targetLang)
	text := Text{
		Text:                translateText,
		RequestAlternatives: 3,
	}
	postData.ID = id
	postData.Params.Texts = append(postData.Params.Texts, text)
	postData.Params.Timestamp = getTimeStamp(getICount(translateText))

	if tagHandling == "html" || tagHandling == "xml" {
		postData.Params.TagHandling = tagHandling
	}

	// Marshalling the request data to JSON and making necessary string replacements
	post_byte, _ := json.Marshal(postData)
	postStr := string(post_byte)

	// Adding spaces to the JSON string based on the ID to adhere to DeepL's request formatting rules
	if (id+5)%29 == 0 || (id+3)%13 == 0 {
		postStr = strings.Replace(postStr, "\"method\":\"", "\"method\" : \"", -1)
	} else {
		postStr = strings.Replace(postStr, "\"method\":\"", "\"method\": \"", -1)
	}

	// Creating a new HTTP POST request with the JSON data as the body
	post_byte = []byte(postStr)
	reader := bytes.NewReader(post_byte)
	request, err := http.NewRequest("POST", proURL, reader)

	if err != nil {
		log.Println(err)
		return DeepLXTranslationResult{
			Code:    http.StatusServiceUnavailable,
			Message: "Post request failed",
		}, nil
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "*/*")
	request.Header.Set("Accept-Language", "en-US,en;q=0.9")
	request.Header.Set("Accept-Encoding", "gzip, deflate, br")
	request.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36")
	request.Header.Set("Origin", "https://www.deepl.com")
	request.Header.Set("Referer", "https://www.deepl.com")
	request.Header.Set("Connection", "keep-alive")
	request.Header.Set("Cookie", "dl_session="+dlSession)

	// Making the HTTP request to the DeepL API
	var client *http.Client
	if proxyURL != "" {
		proxy, err := url.Parse(proxyURL)
		if err != nil {
			return DeepLXTranslationResult{
				Code:    http.StatusServiceUnavailable,
				Message: "DeepL API request failed",
			}, nil
		}
		transport := &http.Transport{
			Proxy: http.ProxyURL(proxy),
		}
		client = &http.Client{Transport: transport}
	} else {
		client = &http.Client{}
	}
	resp, err := client.Do(request)
	if err != nil {
		log.Println(err)
		return DeepLXTranslationResult{
			Code:    http.StatusServiceUnavailable,
			Message: "DeepL API request failed",
		}, nil
	}
	defer resp.Body.Close()

	// Handling potential Brotli compressed response body
	var bodyReader io.Reader
	switch resp.Header.Get("Content-Encoding") {
	case "br":
		bodyReader = brotli.NewReader(resp.Body)
	default:
		bodyReader = resp.Body
	}

	// Reading the response body and parsing it with gjson
	body, _ := io.ReadAll(bodyReader)
	// body, _ := io.ReadAll(resp.Body)
	res := gjson.ParseBytes(body)

	if res.Get("error.code").String() == "-32600" {
		log.Println(res.Get("error").String())
		return DeepLXTranslationResult{
			Code:    http.StatusNotAcceptable,
			Message: "Invalid target language",
		}, nil
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return DeepLXTranslationResult{
			Code:    http.StatusTooManyRequests,
			Message: "Too Many Requests",
		}, nil
	} else if resp.StatusCode == http.StatusUnauthorized {
		return DeepLXTranslationResult{
			Code:    http.StatusUnauthorized,
			Message: "dlsession is invalid",
		}, nil
	} else {
		var alternatives []string
		res.Get("result.texts.0.alternatives").ForEach(func(key, value gjson.Result) bool {
			alternatives = append(alternatives, value.Get("text").String())
			return true
		})
		if res.Get("result.texts.0.text").String() == "" {
			return DeepLXTranslationResult{
				Code:    http.StatusServiceUnavailable,
				Message: "Translation failed, API returns an empty result.",
			}, nil
		} else {
			return DeepLXTranslationResult{
				Code:         http.StatusOK,
				ID:           id,
				Message:      "Success",
				Data:         res.Get("result.texts.0.text").String(),
				Alternatives: alternatives,
				SourceLang:   sourceLang,
				TargetLang:   targetLang,
				Method:       "Pro",
			}, nil
		}
	}
}
