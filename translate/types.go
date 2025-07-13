/*
 * @Author: Vincent Young
 * @Date: 2024-09-16 11:59:24
 * @LastEditors: Vincent Yang
 * @LastEditTime: 2025-03-01 04:16:07
 * @FilePath: /DeepLX/translate/types.go
 * @Telegram: https://t.me/missuo
 * @GitHub: https://github.com/missuo
 *
 * Copyright © 2024 by Vincent, All Rights Reserved.
 */

package translate

// Lang represents the language settings for translation
type Lang struct {
	SourceLangUserSelected string `json:"source_lang_user_selected"` // Can be "auto"
	TargetLang             string `json:"target_lang"`
	SourceLangComputed     string `json:"source_lang_computed,omitempty"`
}

// CommonJobParams represents common parameters for translation jobs
type CommonJobParams struct {
	Formality       string `json:"formality"` // Can be "undefined"
	TranscribeAs    string `json:"transcribe_as"`
	Mode            string `json:"mode"`
	WasSpoken       bool   `json:"wasSpoken"`
	AdvancedMode    bool   `json:"advancedMode"`
	TextType        string `json:"textType"`
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

// TextItem represents a text item for translation
type TextItem struct {
	Text                string `json:"text"`
	RequestAlternatives int    `json:"requestAlternatives"`
}

// Params represents parameters for translation requests
type Params struct {
	Splitting string     `json:"splitting"`
	Lang      Lang       `json:"lang"`
	Texts     []TextItem `json:"texts"`
	Timestamp int64      `json:"timestamp"`
}

// LegacyParams represents the old parameters structure for jobs (kept for compatibility)
type LegacyParams struct {
	CommonJobParams CommonJobParams `json:"commonJobParams"`
	Lang            Lang            `json:"lang"`
	Jobs            []Job           `json:"jobs"`
	Timestamp       int64           `json:"timestamp"`
}

// PostData represents the complete translation request
type PostData struct {
	Jsonrpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	ID      int64  `json:"id"`
	Params  Params `json:"params"`
}

// TextResponse represents a single text response
type TextResponse struct {
	Text         string   `json:"text"`
	Alternatives []struct {
		Text string `json:"text"`
	} `json:"alternatives"`
}

// TranslationResponse represents the response from LMT_handle_texts
type TranslationResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      int64  `json:"id"`
	Result  struct {
		Lang  string          `json:"lang"`
		Texts []TextResponse  `json:"texts"`
	} `json:"result"`
}

// LegacyTranslationResponse represents the old response format (kept for compatibility)
type LegacyTranslationResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      int64  `json:"id"`
	Result  struct {
		Translations []struct {
			Beams []struct {
				Sentences       []SentenceResponse `json:"sentences"`
				NumSymbols      int                `json:"num_symbols"`
				RephraseVariant struct {           // Added rephrase_variant
					Name string `json:"name"`
				} `json:"rephrase_variant"`
			} `json:"beams"`
			Quality string `json:"quality"` // Added quality
		} `json:"translations"`
		TargetLang            string                 `json:"target_lang"`
		SourceLang            string                 `json:"source_lang"`
		SourceLangIsConfident bool                   `json:"source_lang_is_confident"`
		DetectedLanguages     map[string]interface{} `json:"detectedLanguages"` // Use interface{} for now
	} `json:"result"`
}

// SentenceResponse is a helper struct for the response sentences
type SentenceResponse struct {
	Text string `json:"text"`
	IDS  []int  `json:"ids"` // Added IDS
}

// DeepLXTranslationResult represents the final translation result
type DeepLXTranslationResult struct {
	Code         int      `json:"code"`
	ID           int64    `json:"id"`
	Message      string   `json:"message,omitempty"`
	Data         string   `json:"data"`         // The primary translated text
	Alternatives []string `json:"alternatives"` // Other possible translations
	SourceLang   string   `json:"source_lang"`
	TargetLang   string   `json:"target_lang"`
	Method       string   `json:"method"`
}
