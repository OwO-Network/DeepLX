/*
 * @Author: Vincent Yang
 * @Date: 2023-07-01 21:45:34
 * @LastEditors: Jason Lyu
 * @LastEditTime: 2025-04-08 13:45:00
 * @FilePath: /DeepLX/main.go
 * @Telegram: https://t.me/missuo
 * @GitHub: https://github.com/missuo
 *
 * Copyright © 2024 by Vincent, All Rights Reserved.
 */

package main

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/OwO-Network/DeepLX/service"
)

func main() {
	cfg := service.InitConfig()

	fmt.Printf("DeepL X has been successfully launched! Listening on %v:%v\n", cfg.IP, cfg.Port)
	fmt.Println("Developed by sjlleo <i@leo.moe> and missuo <me@missuo.me>.")

	// Setting the application to release mode
	gin.SetMode(gin.ReleaseMode)

	app := service.Router(cfg)
	app.Run(fmt.Sprintf("%v:%v", cfg.IP, cfg.Port))
}
