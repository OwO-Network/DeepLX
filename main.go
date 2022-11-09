package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

type Lang struct {
	SourceLangUserSelected string `json:"source_lang_user_selected"`
	TargetLang             string `json:"target_lang"`
}

type CommonJobParams struct {
	WasSpoken       bool   `json:"wasSpoken"`
	RegionalVariant string `json:"regionalVariant"`
}

type Params struct {
	Texts     []Text `json:"texts"`
	Splitting string `json:"splitting"`
	Lang      Lang   `json:"lang"`
	Timestamp int64  `json:"timestamp"`
	// CommonJobParams CommonJobParams `json:"commonJobParams"`
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

func init_data(source_lang string, target_lang string) *PostData {
	return &PostData{
		Jsonrpc: "2.0",
		Method:  "LMT_handle_texts",
		Params: Params{
			Splitting: "newlines",
			Lang: Lang{
				SourceLangUserSelected: source_lang,
				TargetLang:             target_lang,
			},
		},
	}
}

func get_i_count(translate_text string) int64 {
	return int64(strings.Count(translate_text, "i"))
}

func getRandomNumber() int64 {
	rand.Seed(time.Now().Unix())
	num := rand.Int63n(99999) + 100000
	return num * 1000
}

func getTimeStamp(i_count int64) int64 {
	ts := time.Now().UnixMilli()
	if i_count != 0 {
		return ts - ts%i_count + i_count
	} else {
		return ts
	}

}

type ResData struct {
	Trans_Text  string `json:"text"`
	Source_Lang string `json:"source_lang"`
	Target_Lang string `json:"target_lang"`
}

func main() {
	// create a random id
	id := getRandomNumber()
	r := gin.Default()
	// r.SetTrustedProxies([]string{"192.168.36.153"})
	r.GET("/", func(c *gin.Context) {

		c.JSON(200, gin.H{
			"code": 200,
			"msg":  "DeepL API, Made by sjlleo and missuo. Go to /translate with POST.",
		})

	})

	r.POST("/translate", func(c *gin.Context) {
		reqj := ResData{}
		c.BindJSON(&reqj)
		// fmt.Printf("%v", &reqj)
		source_lang := reqj.Source_Lang
		target_lang := reqj.Target_Lang
		// fmt.Println(reqj)
		if source_lang == "" {
			source_lang = "ZH"
		}
		if target_lang == "" {
			target_lang = "EN"
		}
		translate_text := reqj.Trans_Text
		// fmt.Printf("%v", translate_text)
		if translate_text != "" {

			url := "https://www2.deepl.com/jsonrpc"

			id = id + 1

			post_data := init_data(source_lang, target_lang)

			text := Text{
				Text:                translate_text,
				RequestAlternatives: 3,
			}

			// set id
			post_data.ID = id
			// set text
			post_data.Params.Texts = append(post_data.Params.Texts, text)
			// set timestamp
			post_data.Params.Timestamp = getTimeStamp(get_i_count(translate_text))

			post_byte, _ := json.Marshal(post_data)

			post_str := string(post_byte)

			// add space if necessary
			if (id+5)%29 == 0 || (id+3)%13 == 0 {
				post_str = strings.Replace(post_str, "\"method\":\"", "\"method\" : \"", -1)
			} else {
				post_str = strings.Replace(post_str, "\"method\":\"", "\"method\": \"", -1)
			}

			post_byte = []byte(post_str)

			reader := bytes.NewReader(post_byte)
			request, err := http.NewRequest("POST", url, reader)
			if err != nil {
				log.Println(err)
				return
			}

			request.Header.Set("Content-Type", "application/json")
			client := &http.Client{}
			resp, err := client.Do(request)
			if err != nil {
				log.Println(err)
				return
			}

			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			res := gjson.ParseBytes(body)
			// fmt.Println(res)
			// fmt.Println(res.Get("result").Bool())
			if res.Get("error.code").String() == "-32600" {
				log.Println(res.Get("error").String())
				c.JSON(406, gin.H{
					"code": 406,
					"msg":  "target_lang is not supported",
				})
				return
			} else {
				c.JSON(200, gin.H{
					"code": 200,
					"id":   id,
					"data": res.Get("result.texts.0.text").String(),
				})
			}

			// data = res.Get("result.texts.0.text").String()
			// if res.Get("result.lang_is_confident").String() == "false" {

			// fmt.Printf(res.Get("result.texts.0.text").String())

		} else {
			c.JSON(404, gin.H{
				"code": 404,
				"msg":  "no text found",
			})
		}
	})
	r.Run(":1199") // listen and serve on 0.0.0.0:1199

}
