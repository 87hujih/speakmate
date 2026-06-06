package model

// ScoreResult 表示单条用户消息在当前训练场景下的分项评分。
type ScoreResult struct {
	MessageID  int    `json:"message_id"`
	SessionID  int    `json:"session_id"`
	Fluency    int    `json:"fluency"`
	Grammar    int    `json:"grammar"`
	Expression int    `json:"expression"`
	Vocabulary int    `json:"vocabulary"`
	Completion int    `json:"completion"`
	TotalScore int    `json:"total_score"`
	Comment    string `json:"comment"`
}
