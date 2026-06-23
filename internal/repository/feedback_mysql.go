package repository

import (
	"database/sql"

	"speakmate/internal/model"
)

// MySQLFeedbackRepository 使用 MySQL 保存纠错和评分。
type MySQLFeedbackRepository struct {
	db *sql.DB
}

// NewMySQLFeedbackRepository 创建 MySQL Feedback 仓库。
func NewMySQLFeedbackRepository(db *sql.DB) *MySQLFeedbackRepository {
	return &MySQLFeedbackRepository{db: db}
}

// SaveCorrection 保存或覆盖单条消息的纠错结果。
func (r *MySQLFeedbackRepository) SaveCorrection(correction model.CorrectionResult) error {
	errorsJSON, err := marshalJSON(correction.Errors)
	if err != nil {
		return err
	}
	betterExpressionsJSON, err := marshalJSON(correction.BetterExpressions)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(
		`INSERT INTO corrections (message_id, session_id, original_text, corrected_text, errors_json, better_expressions_json)
VALUES (?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
session_id = VALUES(session_id),
original_text = VALUES(original_text),
corrected_text = VALUES(corrected_text),
errors_json = VALUES(errors_json),
better_expressions_json = VALUES(better_expressions_json),
updated_at = CURRENT_TIMESTAMP`,
		correction.MessageID,
		correction.SessionID,
		correction.OriginalText,
		correction.CorrectedText,
		errorsJSON,
		betterExpressionsJSON,
	)

	return err
}

// SaveScore 保存单条消息评分。
func (r *MySQLFeedbackRepository) SaveScore(score model.ScoreResult) error {
	_, err := r.db.Exec(
		`INSERT INTO scores (message_id, session_id, fluency, grammar, expression, vocabulary, completion, total_score, comment)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
session_id = VALUES(session_id),
fluency = VALUES(fluency),
grammar = VALUES(grammar),
expression = VALUES(expression),
vocabulary = VALUES(vocabulary),
completion = VALUES(completion),
total_score = VALUES(total_score),
comment = VALUES(comment),
updated_at = CURRENT_TIMESTAMP`,
		score.MessageID,
		score.SessionID,
		score.Fluency,
		score.Grammar,
		score.Expression,
		score.Vocabulary,
		score.Completion,
		score.TotalScore,
		score.Comment,
	)

	return err
}

// FindCorrectionByMessageID 按 message_id 查询单条纠错结果。
func (r *MySQLFeedbackRepository) FindCorrectionByMessageID(messageID int) (model.CorrectionResult, error) {
	row := r.db.QueryRow(
		`SELECT message_id, session_id, original_text, corrected_text, errors_json, better_expressions_json FROM corrections WHERE message_id = ?`,
		messageID,
	)
	correction, err := scanCorrection(row)
	if err != nil {
		return model.CorrectionResult{}, notFoundFromNoRows(err, ErrCorrectionNotFound)
	}

	return correction, nil
}

// ListCorrectionsBySessionID 按 session_id 查询全部纠错结果。
func (r *MySQLFeedbackRepository) ListCorrectionsBySessionID(sessionID int) ([]model.CorrectionResult, error) {
	rows, err := r.db.Query(
		`SELECT message_id, session_id, original_text, corrected_text, errors_json, better_expressions_json FROM corrections WHERE session_id = ? ORDER BY message_id ASC`,
		sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	corrections := []model.CorrectionResult{}
	for rows.Next() {
		correction, err := scanCorrection(rows)
		if err != nil {
			return nil, err
		}
		corrections = append(corrections, correction)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(corrections) == 0 {
		return nil, ErrCorrectionNotFound
	}

	return corrections, nil
}

// FindCurrentScoreBySessionID 按 session_id 查询当前评分。
func (r *MySQLFeedbackRepository) FindCurrentScoreBySessionID(sessionID int) (model.ScoreResult, error) {
	row := r.db.QueryRow(
		`SELECT message_id, session_id, fluency, grammar, expression, vocabulary, completion, total_score, comment FROM scores WHERE session_id = ? ORDER BY message_id DESC LIMIT 1`,
		sessionID,
	)
	score, err := scanScore(row)
	if err != nil {
		return model.ScoreResult{}, notFoundFromNoRows(err, ErrScoreNotFound)
	}

	return score, nil
}

// scanCorrection 从数据库行读取纠错结果。
func scanCorrection(row scanner) (model.CorrectionResult, error) {
	var correction model.CorrectionResult
	var errorsJSON string
	var betterExpressionsJSON string
	if err := row.Scan(
		&correction.MessageID,
		&correction.SessionID,
		&correction.OriginalText,
		&correction.CorrectedText,
		&errorsJSON,
		&betterExpressionsJSON,
	); err != nil {
		return model.CorrectionResult{}, err
	}
	if err := unmarshalJSON(errorsJSON, &correction.Errors); err != nil {
		return model.CorrectionResult{}, err
	}
	if correction.Errors == nil {
		correction.Errors = []model.CorrectionError{}
	}
	if err := unmarshalJSON(betterExpressionsJSON, &correction.BetterExpressions); err != nil {
		return model.CorrectionResult{}, err
	}
	if correction.BetterExpressions == nil {
		correction.BetterExpressions = []string{}
	}

	return correction, nil
}

// scanScore 从数据库行读取评分结果。
func scanScore(row scanner) (model.ScoreResult, error) {
	var score model.ScoreResult
	if err := row.Scan(
		&score.MessageID,
		&score.SessionID,
		&score.Fluency,
		&score.Grammar,
		&score.Expression,
		&score.Vocabulary,
		&score.Completion,
		&score.TotalScore,
		&score.Comment,
	); err != nil {
		return model.ScoreResult{}, err
	}

	return score, nil
}
