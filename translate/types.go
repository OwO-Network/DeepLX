/*
 * @Author: Vincent Young
 * @Date: 2024-09-16 11:59:24
 * @LastEditors: Vincent Young
 * @LastEditTime: 2024-09-16 12:06:36
 * @FilePath: /DeepLX/translate/types.go
 * @Telegram: https://t.me/missuo
 * @GitHub: https://github.com/missuo
 *
 * Copyright © 2024 by Vincent, All Rights Reserved.
 */

package translate

type Lang struct {
	SourceLangUserSelected string `json:"source_lang_user_selected"`
	TargetLang             string `json:"target_lang"`
}

type CommonJobParams struct {
	WasSpoken       bool   `json:"wasSpoken"`
	TranscribeAS    string `json:"transcribe_as"`
	RegionalVariant string `json:"regionalVariant,omitempty"`
}

type Params struct {
	Texts           []Text          `json:"texts"`
	Splitting       string          `json:"splitting"`
	Lang            Lang            `json:"lang"`
	Timestamp       int64           `json:"timestamp"`
	CommonJobParams CommonJobParams `json:"commonJobParams"`
	TagHandling     string          `json:"tag_handling"`
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

type Translation struct {
	Text string `json:"text"`
}

type TranslationResponse struct {
	Translations []Translation `json:"translations"`
}

type DeepLUsageResponse struct {
	CharacterCount int `json:"character_count"`
	CharacterLimit int `json:"character_limit"`
}

type DeepLXTranslationResult struct {
	Code         int
	ID           int64
	Message      string
	Data         string
	Alternatives []string
	SourceLang   string
	TargetLang   string
	Method       string
}
