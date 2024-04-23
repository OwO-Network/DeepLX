/*
 * @Author: Vincent Yang
 * @Date: 2023-07-01 21:45:34
 * @LastEditors: Vincent Yang
 * @LastEditTime: 2024-04-23 00:42:28
 * @FilePath: /DeepLX/main.go
 * @Telegram: https://t.me/missuo
 * @GitHub: https://github.com/missuo
 *
 * Copyright © 2024 by Vincent, All Rights Reserved.
 */
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func authMiddleware(cfg *Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cfg.Token != "" {
			providedTokenInQuery := c.Query("token")
			providedTokenInHeader := c.GetHeader("Authorization")

			// Compatability with the Bearer token format
			if providedTokenInHeader != "" {
				parts := strings.Split(providedTokenInHeader, " ")
				if len(parts) == 2 {
					if parts[0] == "Bearer" || parts[0] == "DeepL-Auth-Key" {
						providedTokenInHeader = parts[1]
					} else {
						providedTokenInHeader = ""
					}
				} else {
					providedTokenInHeader = ""
				}
			}

			if providedTokenInHeader != cfg.Token && providedTokenInQuery != cfg.Token {
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

	// Free API endpoint, No Pro Account required
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

	// Pro API endpoint, Pro Account required
	r.POST("/v1/translate", authMiddleware(cfg), func(c *gin.Context) {
		req := PayloadFree{}
		c.BindJSON(&req)

		sourceLang := req.SourceLang
		targetLang := req.TargetLang
		translateText := req.TransText

		dlSession := cfg.DlSession
		if dlSession == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"message": "No dl_session Found",
			})
			return
		} else if strings.Contains(dlSession, ".") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"message": "Your account is not a Pro account. Please upgrade your account or switch to a different account.",
			})
			return
		}

		result, err := translateByDeepLXPro(sourceLang, targetLang, translateText, dlSession)
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

	// Free API endpoint, Consistent with the official API format
	r.POST("/v2/translate", authMiddleware(cfg), func(c *gin.Context) {
		authorizationHeader := c.GetHeader("Authorization")
		var authKey string

		if strings.HasPrefix(authorizationHeader, "DeepL-Auth-Key") {
			parts := strings.Split(authorizationHeader, " ")
			if len(parts) >= 2 && strings.HasSuffix(parts[len(parts)-1], ":fx") {
				authKey = parts[len(parts)-1]
			}
		}

		var translateText string
		var targetLang string

		translateText = c.PostForm("text")
		targetLang = c.PostForm("target_lang")

		if translateText == "" || targetLang == "" {
			var jsonData struct {
				Text       []string `json:"text"`
				TargetLang string   `json:"target_lang"`
			}

			if err := c.BindJSON(&jsonData); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"code":    http.StatusBadRequest,
					"message": "Invalid request payload",
				})
				return
			}

			translateText = strings.Join(jsonData.Text, "\n")
			targetLang = jsonData.TargetLang
		}

		result, err := translateByDeepLX("", targetLang, translateText, authKey)
		if err != nil {
			log.Fatalf("Translation failed: %s", err)
		}

		if result.Code == http.StatusOK {
			c.JSON(http.StatusOK, gin.H{
				"translations": []map[string]interface{}{
					{
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
