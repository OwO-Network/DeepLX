/*
 * @Author: Vincent Yang
 * @Date: 2023-07-01 21:45:34
 * @LastEditors: Vincent Yang
 * @LastEditTime: 2024-04-13 18:50:25
 * @FilePath: /DeepLX/main.go
 * @Telegram: https://t.me/missuo
 * @GitHub: https://github.com/missuo
 *
 * Copyright © 2024 by Vincent, All Rights Reserved.
 */
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/abadojack/whatlanggo"
	"github.com/andybalholm/brotli"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

func initConfig() *Config {
	cfg := &Config{
		Port: 1188,
	}

	flag.IntVar(&cfg.Port, "port", cfg.Port, "set up the port to listen on")
	flag.IntVar(&cfg.Port, "p", cfg.Port, "set up the port to listen on")

	flag.StringVar(&cfg.Token, "token", "", "set the access token for /translate endpoint")
	if cfg.Token == "" {
		if token, ok := os.LookupEnv("TOKEN"); ok {
			cfg.Token = token
		}
	}

	flag.StringVar(&cfg.AuthKey, "authkey", "", "The authentication key for DeepL API")
	if cfg.AuthKey == "" {
		if authKey, ok := os.LookupEnv("AUTHKEY"); ok {
			cfg.AuthKey = authKey
		}
	}

	flag.Parse()
	return cfg
}

func initDeepLXData(sourceLang string, targetLang string) *PostData {
	return &PostData{
		Jsonrpc: "2.0",
		Method:  "LMT_handle_texts",
		Params: Params{
			Splitting: "newlines",
			Lang: Lang{
				SourceLangUserSelected: sourceLang,
				TargetLang:             targetLang,
			},
			CommonJobParams: CommonJobParams{
				WasSpoken:    false,
				TranscribeAS: "",
				// RegionalVariant: "en-US",
			},
		},
	}
}

func getICount(translateText string) int64 {
	return int64(strings.Count(translateText, "i"))
}

func getRandomNumber() int64 {
	src := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(src)
	num := rng.Int63n(99999) + 8300000
	return num * 1000
}

func getTimeStamp(iCount int64) int64 {
	ts := time.Now().UnixMilli()
	if iCount != 0 {
		iCount = iCount + 1
		return ts - ts%iCount + iCount
	} else {
		return ts
	}
}

func checkUsageAuthKey(authKey string) (bool, error) {
	url := "https://api-free.deepl.com/v2/usage"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}

	req.Header.Add("Authorization", "DeepL-Auth-Key "+authKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	var response DeepLUsageResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return false, err
	}
	return response.CharacterCount < 499900, nil
}

func translateByOfficialAPI(text string, sourceLang string, targetLang string, authKey string) (string, error) {
	url := "https://api-free.deepl.com/v2/translate"
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

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "DeepL-Auth-Key "+authKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
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

func translateByDeepLX(sourceLang string, targetLang string, translateText string, authKey string) (DeepLXTranslationResult, error) {
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
	url := "https://www2.deepl.com/jsonrpc"
	id = id + 1
	postData := initDeepLXData(sourceLang, targetLang)
	text := Text{
		Text:                translateText,
		RequestAlternatives: 3,
	}
	postData.ID = id
	postData.Params.Texts = append(postData.Params.Texts, text)
	postData.Params.Timestamp = getTimeStamp(getICount(translateText))

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
	request, err := http.NewRequest("POST", url, reader)

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
	client := &http.Client{}
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
					translatedText, err := translateByOfficialAPI(translateText, sourceLang, targetLang, authKey)
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

func authMiddleware(cfg *Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cfg.Token != "" {
			providedTokenInQuery := c.Query("token")
			providedTokenInHeader := c.GetHeader("Authorization")
			if providedTokenInHeader != "Bearer "+cfg.Token && providedTokenInQuery != cfg.Token {
				c.JSON(http.StatusUnauthorized, gin.H{
					"code":    http.StatusUnauthorized,
					"message": "Invalid access token",
				})
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

func main() {
	cfg := initConfig()

	fmt.Printf("DeepL X has been successfully launched! Listening on 0.0.0.0:%v\n", cfg.Port)
	fmt.Println("Developed by sjlleo <i@leo.moe> and missuo <me@missuo.me>.")

	if cfg.Token != "" {
		fmt.Println("Access token is set.")
	}
	if cfg.AuthKey != "" {
		fmt.Println("DeepL Official Authentication key is set.")
	}

	// Setting the application to release mode
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(cors.Default())

	// Defining the root endpoint which returns the project details
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"code":    http.StatusOK,
			"message": "DeepL Free API, Developed by sjlleo and missuo. Go to /translate with POST. http://github.com/OwO-Network/DeepLX",
		})
	})

	// Defining the translation endpoint which receives translation requests and returns translations
	r.POST("/translate", authMiddleware(cfg), func(c *gin.Context) {
		req := PayloadFree{}
		c.BindJSON(&req)

		sourceLang := req.SourceLang
		targetLang := req.TargetLang
		translateText := req.TransText
		authKey := cfg.AuthKey

		result, err := translateByDeepLX(sourceLang, targetLang, translateText, authKey)
		if err != nil {
			log.Fatalf("Translation failed: %s", err)
		}

		if result.Code == http.StatusOK {
			c.JSON(http.StatusOK, gin.H{
				"code":         http.StatusOK,
				"id":           result.ID,
				"data":         result.Data,
				"alternatives": result.Alternatives,
				"source_lang":  result.SourceLang,
				"target_lang":  result.TargetLang,
				"method":       result.Method,
			})
		} else {
			c.JSON(result.Code, gin.H{
				"code":    result.Code,
				"message": result.Message,
			})

		}
	})

	r.POST("/v2/translate", authMiddleware(cfg), func(c *gin.Context) {
		authorizationHeader := c.GetHeader("Authorization")
		parts := strings.Split(authorizationHeader, " ")
		var authKey string
		if len(parts) == 2 {
			authKey = parts[1]
		}
		translateText := c.PostForm("text")
		targetLang := c.PostForm("target_lang")
		result, err := translateByDeepLX("", targetLang, translateText, authKey)
		if err != nil {
			log.Fatalf("Translation failed: %s", err)
		}
		if result.Code == http.StatusOK {
			c.JSON(http.StatusOK, gin.H{
				"translations": []interface{}{
					map[string]interface{}{
						"detected_source_language": result.SourceLang,
						"text":                     result.Data,
					},
				},
			})
		} else {
			c.JSON(result.Code, gin.H{
				"code":    result.Code,
				"message": result.Message,
			})
		}
	})

	// Catch-all route to handle undefined paths
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    http.StatusNotFound,
			"message": "Path not found",
		})
	})

	envPort, ok := os.LookupEnv("PORT")
	if ok {
		r.Run(":" + envPort)
	} else {
		r.Run(fmt.Sprintf(":%v", cfg.Port))
	}
}
