package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

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
			// CommonJobParams: CommonJobParams{
			// 	WasSpoken:       false,
			// 	RegionalVariant: "en-US",
			// },
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

func main() {

	url := "https://www2.deepl.com/jsonrpc"
	id := getRandomNumber()
	fmt.Printf("ID = %d\n", id)

	// ZH - 中文
	// EN - 英文
	post_data := init_data("", "EN")

	translate_text := `
	一是课题来源渠道，比如指导性课题、指令性课题、机构科研项目等,
	二是课题选题来源渠道，比如课题指南、自拟选题等,
	三是有时候课题来源也理解为课题选题依据。
	一般来说，课题来源主要是指课题的从什么地方获得，即课题的方向。
	`
	text := Text{
		Text: translate_text,
		// 不要尝试修改 RequestAlternatives，会被 Ban IP
		RequestAlternatives: 3,
	}

	// 设置 id
	post_data.ID = id
	// 设置翻译文本
	post_data.Params.Texts = append(post_data.Params.Texts, text)
	// 设置时间戳
	post_data.Params.Timestamp = getTimeStamp(get_i_count(translate_text))

	post_byte, _ := json.Marshal(post_data)

	post_str := string(post_byte)

	// 判断是否需要加空格
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
	if res.Get("result.lang_is_confident").String() == "false" {
		fmt.Println("引擎可能无法正确判断源文语言")
	}
	// 源语言
	fmt.Println(res.Get("result.lang").String())
	// 译文
	fmt.Println(res.Get("result.texts.0.text").String())
	// 译文候选一
	// fmt.Println(res.Get("result.texts.0.alternatives.0.text").String())
	// 译文候选二
	// fmt.Println(res.Get("result.texts.0.alternatives.1.text").String())
	// 译文候选三
	// fmt.Println(res.Get("result.texts.0.alternatives.2.text").String())
}
