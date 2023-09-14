/*
 * @Author: Vincent Young
 * @Date: 2023-07-01 21:45:34
 * @LastEditors: Vincent Young
 * @LastEditTime: 2023-09-14 13:34:42
 * @FilePath: /DeepLX/main.go
 * @Telegram: https://t.me/missuo
 *
 * Copyright © 2023 by Vincent, All Rights Reserved.
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

var port int

func init() {
	const (
		defaultPort = 1188
		usage       = "set up the port to listen on"
	)

	flag.IntVar(&port, "port", defaultPort, usage)
	flag.IntVar(&port, "p", defaultPort, usage)

	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

type Lang struct {
	SourceLangUserSelected string `json:"source_lang_user_selected"`
	TargetLang             string `json:"target_lang"`
}

type CommonJobParams struct {
	WasSpoken    bool   `json:"wasSpoken"`
	TranscribeAS string `json:"transcribe_as"`
	// RegionalVariant string `json:"regionalVariant"`
}

type Params struct {
	Texts           []Text          `json:"texts"`
	Splitting       string          `json:"splitting"`
	Lang            Lang            `json:"lang"`
	Timestamp       int64           `json:"timestamp"`
	CommonJobParams CommonJobParams `json:"commonJobParams"`
}

type Text struct {
	Text                string `json:"text"`
	RequestAlternatives int    `json:"requestAlternatives"`
}

type PostData struct {
	Jsonrpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	ID      int64  `json:"id"`
	Params  Params `json:"params"`
}

func initData(sourceLang string, targetLang string) *PostData {
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
	rand.Seed(time.Now().Unix())
	num := rand.Int63n(99999) + 8300000
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

type ResData struct {
	TransText  string `json:"text"`
	SourceLang string `json:"source_lang"`
	TargetLang string `json:"target_lang"`
}

func main() {
	// Parsing the command-line flags
	flag.Parse()

	// Displaying initialization information
	fmt.Printf("DeepL X has been successfully launched! Listening on 0.0.0.0:%v\n", port)
	fmt.Println("Developed by sjlleo <i@leo.moe> and missuo <me@missuo.me>.")

	// Generating a random ID
	id := getRandomNumber()

	// Setting the application to release mode
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(cors.Default())

	// Defining the root endpoint which returns the project details
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"code":    200,
			"message": "DeepL Free API, Developed by sjlleo <i@leo.moe> and missuo <me@missuo.me>. Go to /translate with POST. http://github.com/OwO-Network/DeepLX",
		})
	})

	// Defining the translation endpoint which receives translation requests and returns translations
	r.POST("/translate", func(c *gin.Context) {
		reqj := ResData{}
		c.BindJSON(&reqj)

		// Extracting details from the request JSON
		sourceLang := reqj.SourceLang
		targetLang := reqj.TargetLang
		translateText := reqj.TransText

		// If source language is not specified, auto-detect it
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
			c.JSON(http.StatusNotFound, gin.H{
				"code":    http.StatusNotFound,
				"message": "No Translate Text Found",
			})
			return
		}
		// Preparing the request data for the DeepL API
		url := "https://www2.deepl.com/jsonrpc"
		id = id + 1
		postData := initData(sourceLang, targetLang)
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
			return
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
			return
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
		body, err := io.ReadAll(bodyReader)
		// body, _ := io.ReadAll(resp.Body)
		res := gjson.ParseBytes(body)

		// Handling various response statuses and potential errors
		if res.Get("error.code").String() == "-32600" {
			log.Println(res.Get("error").String())
			c.JSON(http.StatusNotAcceptable, gin.H{
				"code":    http.StatusNotAcceptable,
				"message": "Invalid targetLang",
			})
			return
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":    http.StatusTooManyRequests,
				"message": "Too Many Requests",
			})
		} else {
			var alternatives []string
			res.Get("result.texts.0.alternatives").ForEach(func(key, value gjson.Result) bool {
				alternatives = append(alternatives, value.Get("text").String())
				return true
			})
			c.JSON(http.StatusOK, gin.H{
				"code":         http.StatusOK,
				"id":           id,
				"data":         res.Get("result.texts.0.text").String(),
				"alternatives": alternatives,
				"source_lang":  sourceLang,
				"target_lang":  targetLang,
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

	// Determining which port to run the server on, with a fallback to a default port
	envPort, ok := os.LookupEnv("PORT")
	if ok {
		r.Run(":" + envPort)
	} else {
		r.Run(fmt.Sprintf(":%v", port))
	}
}
