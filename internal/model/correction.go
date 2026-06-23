package model

// CorrectionErrorType 表示纠错问题分类。
type CorrectionErrorType string

// 当前模型使用的枚举常量。
const (
	// CorrectionErrorTypeGrammar 表示语法错误。
	CorrectionErrorTypeGrammar CorrectionErrorType = "grammar"
	// CorrectionErrorTypeVocabulary 表示用词不准确。
	CorrectionErrorTypeVocabulary CorrectionErrorType = "vocabulary"
	// CorrectionErrorTypeExpression 表示表达不自然。
	CorrectionErrorTypeExpression CorrectionErrorType = "expression"
	// CorrectionErrorTypeStructure 表示句子结构问题。
	CorrectionErrorTypeStructure CorrectionErrorType = "structure"
	// CorrectionErrorTypeScenario 表示不符合当前场景。
	CorrectionErrorTypeScenario CorrectionErrorType = "scenario"
)

// CorrectionError 表示用户表达中的一个具体问题。
type CorrectionError struct {
	Type        CorrectionErrorType `json:"type"`
	Span        string              `json:"span"`
	Suggestion  string              `json:"suggestion"`
	Explanation string              `json:"explanation"`
}

// CorrectionResult 表示单条用户消息的结构化纠错结果。
type CorrectionResult struct {
	MessageID         int               `json:"message_id"`
	SessionID         int               `json:"session_id"`
	OriginalText      string            `json:"original_text"`
	CorrectedText     string            `json:"corrected_text"`
	Errors            []CorrectionError `json:"errors"`
	BetterExpressions []string          `json:"better_expressions"`
}
