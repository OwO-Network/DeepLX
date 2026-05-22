/*
 * @Author: Vincent Young
 * @Date: 2024-09-16 11:59:24
 * @LastEditors: Vincent Yang
 * @LastEditTime: 2026-05-22 00:00:00
 * @FilePath: /DeepLX/translate/types.go
 * @Telegram: https://t.me/missuo
 * @GitHub: https://github.com/missuo
 *
 * Copyright © 2024 by Vincent, All Rights Reserved.
 */

package translate

// DeepLXTranslationResult is the public response shape consumed by the HTTP
// handlers in the service package. The structure predates the migration to
// the oneshot endpoint; Alternatives is now always empty because oneshot does
// not return alternative translations, and ID is synthesized from time.
type DeepLXTranslationResult struct {
	Code         int      `json:"code"`
	ID           int64    `json:"id"`
	Message      string   `json:"message,omitempty"`
	Data         string   `json:"data"`
	Alternatives []string `json:"alternatives"`
	SourceLang   string   `json:"source_lang"`
	TargetLang   string   `json:"target_lang"`
	Method       string   `json:"method"`
}
