/*
 * @Author: Vincent Young
 * @Date: 2024-09-16 11:59:24
 * @LastEditors: Vincent Yang
 * @LastEditTime: 2025-04-08 14:27:21
 * @FilePath: /DeepLX/translate/utils.go
 * @Telegram: https://t.me/missuo
 * @GitHub: https://github.com/missuo
 *
 * Copyright © 2024 by Vincent, All Rights Reserved.
 */

package translate

import (
	"encoding/json"
	"math/rand"
	"strings"
	"time"
)

// getICount returns the number of 'i' characters in the text
func getICount(translateText string) int64 {
	return int64(strings.Count(translateText, "i"))
}

// getRandomNumber generates a random number for request ID
func getRandomNumber() int64 {
	src := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(src)
	num := rng.Int63n(99999) + 100000
	return num * 1000
}

// getTimeStamp generates timestamp for request based on i count
func getTimeStamp(iCount int64) int64 {
	ts := time.Now().UnixMilli()
	if iCount != 0 {
		iCount = iCount + 1
		return ts - (ts % iCount) + iCount
	}
	return ts
}

// formatPostString formats the request JSON string with specific spacing rules
func formatPostString(postData *PostData) string {
	postBytes, _ := json.Marshal(postData)
	postStr := string(postBytes)
	return postStr
}

// handlerBodyMethod manipulates the request body based on random number calculation
func handlerBodyMethod(random int64, body string) string {
	calc := (random+5)%29 == 0 || (random+3)%13 == 0
	if calc {
		return strings.Replace(body, `"method":"`, `"method" : "`, 1)
	}
	return strings.Replace(body, `"method":"`, `"method": "`, 1)
}
