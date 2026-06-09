package repository

import (
	"database/sql"
	"fmt"
	"time"

	"speakmate/internal/model"
)

// MySQLSessionRepository 使用 MySQL 保存训练 Session 和消息。
type MySQLSessionRepository struct {
	db *sql.DB
}

// NewMySQLSessionRepository 创建 MySQL Session 仓库。
func NewMySQLSessionRepository(db *sql.DB) *MySQLSessionRepository {
	return &MySQLSessionRepository{db: db}
}

// Create 保存新的 Session，并生成 session_id 和 session_no。
func (r *MySQLSessionRepository) Create(session model.Session) (model.Session, error) {
	if session.CreatedAt.IsZero() {
		session.CreatedAt = time.Now()
	}
	if session.Status == "" {
		session.Status = model.SessionStatusRunning
	}
	if session.Messages == nil {
		session.Messages = []model.Message{}
	}
	if session.SessionNo == "" {
		session.SessionNo = provisionalSessionNo(session.CreatedAt)
	}

	result, err := r.db.Exec(
		`INSERT INTO training_sessions (session_no, scenario_id, user_id, status, turn_count, created_at, ended_at)
VALUES (?, ?, ?, ?, ?, ?, ?)`,
		session.SessionNo,
		session.ScenarioID,
		session.UserID,
		string(session.Status),
		session.TurnCount,
		session.CreatedAt,
		nullableTimePtr(session.EndedAt),
	)
	if err != nil {
		return model.Session{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return model.Session{}, err
	}
	session.ID = int(id)

	return cloneSession(session), nil
}

// FindByID 按内部数字 ID 查询 Session，并加载关联消息。
func (r *MySQLSessionRepository) FindByID(id int) (model.Session, error) {
	session, err := r.findSessionOnly(id)
	if err != nil {
		return model.Session{}, err
	}
	messages, err := r.listMessagesBySessionID(id)
	if err != nil {
		return model.Session{}, err
	}
	session.Messages = messages

	return session, nil
}

// ListSessions 按分页条件查询训练 Session 历史。
func (r *MySQLSessionRepository) ListSessions(query model.SessionListQuery) (model.SessionListResult, error) {
	where := ""
	args := []any{}
	if query.UserID > 0 {
		where = " WHERE user_id = ?"
		args = append(args, query.UserID)
	}

	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM training_sessions"+where, args...).Scan(&total); err != nil {
		return model.SessionListResult{}, err
	}
	offset := (query.Page - 1) * query.PageSize
	listArgs := append(append([]any{}, args...), query.PageSize, offset)
	rows, err := r.db.Query(
		`SELECT id, session_no, scenario_id, user_id, status, turn_count, created_at, ended_at
FROM training_sessions`+where+`
ORDER BY created_at DESC, id DESC
LIMIT ? OFFSET ?`,
		listArgs...,
	)
	if err != nil {
		return model.SessionListResult{}, err
	}
	defer rows.Close()

	sessions := []model.Session{}
	for rows.Next() {
		session, err := scanSession(rows)
		if err != nil {
			return model.SessionListResult{}, err
		}
		session.Messages = []model.Message{}
		sessions = append(sessions, session)
	}
	if err := rows.Err(); err != nil {
		return model.SessionListResult{}, err
	}

	return model.SessionListResult{
		Sessions: sessions,
		Total:    total,
	}, nil
}

// ListSessionsByWindow 按创建时间窗口查询训练 Session。
func (r *MySQLSessionRepository) ListSessionsByWindow(query model.SessionWindowQuery) ([]model.Session, error) {
	where := "WHERE created_at >= ? AND created_at < ?"
	args := []any{query.StartedAt, query.EndedAt}
	if query.UserID > 0 {
		where = "WHERE user_id = ? AND created_at >= ? AND created_at < ?"
		args = []any{query.UserID, query.StartedAt, query.EndedAt}
	}
	limit := ""
	if query.Limit > 0 {
		limit = `
LIMIT ?`
		args = append(args, query.Limit)
	}

	rows, err := r.db.Query(
		`SELECT id, session_no, scenario_id, user_id, status, turn_count, created_at, ended_at
FROM training_sessions
`+where+`
ORDER BY created_at DESC, id DESC`+limit,
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessions := []model.Session{}
	for rows.Next() {
		session, err := scanSession(rows)
		if err != nil {
			return nil, err
		}
		session.Messages = []model.Message{}
		sessions = append(sessions, session)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return sessions, nil
}

// Finish 把 running Session 原子地结束为 finished。
func (r *MySQLSessionRepository) Finish(id int, endedAt time.Time) (model.Session, error) {
	result, err := r.db.Exec(
		`UPDATE training_sessions SET status = ?, ended_at = ? WHERE id = ? AND status <> ?`,
		string(model.SessionStatusFinished),
		endedAt,
		id,
		string(model.SessionStatusFinished),
	)
	if err != nil {
		return model.Session{}, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return model.Session{}, err
	}
	if affected == 0 {
		existing, findErr := r.findSessionOnly(id)
		if findErr != nil {
			return model.Session{}, findErr
		}
		if existing.Status == model.SessionStatusFinished {
			return model.Session{}, ErrSessionAlreadyFinished
		}
	}

	return r.FindByID(id)
}

// AppendTurn 原子地追加一轮用户消息和 AI 消息，并递增 Session 轮次。
func (r *MySQLSessionRepository) AppendTurn(id int, userMessage model.Message, aiMessage model.Message) (model.Session, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return model.Session{}, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	userMessage.SessionID = id
	userID, err := insertMessage(tx, userMessage)
	if err != nil {
		return model.Session{}, err
	}
	userMessage.ID = userID
	aiMessage.SessionID = id
	aiID, err := insertMessage(tx, aiMessage)
	if err != nil {
		return model.Session{}, err
	}
	aiMessage.ID = aiID

	result, err := tx.Exec(
		`UPDATE training_sessions SET turn_count = turn_count + 1 WHERE id = ? AND status <> ?`,
		id,
		string(model.SessionStatusFinished),
	)
	if err != nil {
		return model.Session{}, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return model.Session{}, err
	}
	if affected == 0 {
		_ = tx.Rollback()
		committed = true
		existing, findErr := r.findSessionOnly(id)
		if findErr != nil {
			return model.Session{}, findErr
		}
		if existing.Status == model.SessionStatusFinished {
			return model.Session{}, ErrSessionAlreadyFinished
		}

		return model.Session{}, ErrSessionNotFound
	}
	if err := tx.Commit(); err != nil {
		return model.Session{}, err
	}
	committed = true

	return r.FindByID(id)
}

func (r *MySQLSessionRepository) findSessionOnly(id int) (model.Session, error) {
	row := r.db.QueryRow(
		`SELECT id, session_no, scenario_id, user_id, status, turn_count, created_at, ended_at FROM training_sessions WHERE id = ?`,
		id,
	)
	session, err := scanSession(row)
	if err != nil {
		return model.Session{}, notFoundFromNoRows(err, ErrSessionNotFound)
	}
	session.Messages = []model.Message{}

	return session, nil
}

func (r *MySQLSessionRepository) listMessagesBySessionID(sessionID int) ([]model.Message, error) {
	rows, err := r.db.Query(
		`SELECT id, session_id, role, content, stage, created_at FROM messages WHERE session_id = ? ORDER BY id ASC`,
		sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := []model.Message{}
	for rows.Next() {
		var message model.Message
		var role string
		if err := rows.Scan(&message.ID, &message.SessionID, &role, &message.Content, &message.Stage, &message.CreatedAt); err != nil {
			return nil, err
		}
		message.Role = model.MessageRole(role)
		messages = append(messages, message)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

func scanSession(row scanner) (model.Session, error) {
	var session model.Session
	var status string
	var endedAt sql.NullTime
	if err := row.Scan(
		&session.ID,
		&session.SessionNo,
		&session.ScenarioID,
		&session.UserID,
		&status,
		&session.TurnCount,
		&session.CreatedAt,
		&endedAt,
	); err != nil {
		return model.Session{}, err
	}
	session.Status = model.SessionStatus(status)
	session.EndedAt = timePtrFromNull(endedAt)

	return session, nil
}

func insertMessage(tx *sql.Tx, message model.Message) (int, error) {
	result, err := tx.Exec(
		`INSERT INTO messages (session_id, role, content, stage, created_at) VALUES (?, ?, ?, ?, ?)`,
		message.SessionID,
		string(message.Role),
		message.Content,
		message.Stage,
		message.CreatedAt,
	)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func provisionalSessionNo(createdAt time.Time) string {
	return fmt.Sprintf("S%s%06d", createdAt.Format("20060102"), time.Now().UnixNano()%1000000)
}
