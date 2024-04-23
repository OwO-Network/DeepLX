/*
 * @Author: Vincent Yang
 * @Date: 2024-04-23 00:17:27
 * @LastEditors: Vincent Yang
 * @LastEditTime: 2024-04-23 00:17:29
 * @FilePath: /DeepLX/utils.go
 * @Telegram: https://t.me/missuo
 * @GitHub: https://github.com/missuo
 *
 * Copyright © 2024 by Vincent, All Rights Reserved.
 */

package main

import (
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

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
