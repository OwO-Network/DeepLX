/*
 * @Author: Vincent Young
 * @Date: 2024-09-16 11:59:24
 * @LastEditors: Vincent Yang
 * @LastEditTime: 2024-11-01 23:18:56
 * @FilePath: /DeepLX/translate/types.go
 * @Telegram: https://t.me/missuo
 * @GitHub: https://github.com/missuo
 *
 * Copyright © 2024 by Vincent, All Rights Reserved.
 */

package translate

// Lang represents the language settings for translation
type Lang struct {
	SourceLangComputed string `json:"source_lang_computed,omitempty"`
	TargetLang         string `json:"target_lang"`
	LangUserSelected   string `json:"lang_user_selected,omitempty"`
}

// CommonJobParams represents common parameters for translation jobs
type CommonJobParams struct {
	Mode            string `json:"mode"`
	RegionalVariant string `json:"regionalVariant,omitempty"`
}

// Sentence represents a sentence in the translation request
type Sentence struct {
	Prefix string `json:"prefix"`
	Text   string `json:"text"`
	ID     int    `json:"id"`
}

// Job represents a translation job
type Job struct {
	Kind               string     `json:"kind"`
	PreferredNumBeams  int        `json:"preferred_num_beams"`
	RawEnContextBefore []string   `json:"raw_en_context_before"`
	RawEnContextAfter  []string   `json:"raw_en_context_after"`
	Sentences          []Sentence `json:"sentences"`
}

// Params represents parameters for translation requests
type Params struct {
	CommonJobParams CommonJobParams `json:"commonJobParams"`
	Lang            Lang            `json:"lang"`
	Texts           []string        `json:"texts,omitempty"`
	TextType        string          `json:"textType,omitempty"`
	Jobs            []Job           `json:"jobs,omitempty"`
	Priority        int             `json:"priority,omitempty"`
	Timestamp       int64           `json:"timestamp"`
}

// PostData represents the complete translation request
type PostData struct {
	Jsonrpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	ID      int64  `json:"id"`
	Params  Params `json:"params"`
}

// SplitTextResponse represents the response from text splitting
type SplitTextResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      int64  `json:"id"`
	Result  struct {
		Lang struct {
			Detected string `json:"detected"`
		} `json:"lang"`
		Texts []struct {
			Chunks []struct {
				Sentences []struct {
					Prefix string `json:"prefix"`
					Text   string `json:"text"`
				} `json:"sentences"`
			} `json:"chunks"`
		} `json:"texts"`
	} `json:"result"`
}

// TranslationResponse represents the response from translation
type TranslationResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      int64  `json:"id"`
	Result  struct {
		Translations []struct {
			Beams []struct {
				Sentences []struct {
					Text string `json:"text"`
				} `json:"sentences"`
			} `json:"beams"`
		} `json:"translations"`
		SourceLang string `json:"source_lang"`
		TargetLang string `json:"target_lang"`
	} `json:"result"`
}

// DeepLXTranslationResult represents the final translation result
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
