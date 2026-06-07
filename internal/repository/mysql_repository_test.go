package repository_test

import (
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"speakmate/internal/model"
	"speakmate/internal/repository"
)

func TestMySQLSessionRepositoryCreateAndFindByID(t *testing.T) {
	db, mock, cleanup := newSQLMock(t)
	defer cleanup()
	repo := repository.NewMySQLSessionRepository(db)
	createdAt := time.Date(2026, 6, 7, 3, 0, 0, 0, time.UTC)

	mock.ExpectExec("INSERT INTO training_sessions").
		WithArgs(sqlmock.AnyArg(), 1, 42, string(model.SessionStatusRunning), 0, createdAt, nil).
		WillReturnResult(sqlmock.NewResult(7, 1))

	created, err := repo.Create(model.Session{
		ScenarioID: 1,
		UserID:     42,
		Status:     model.SessionStatusRunning,
		TurnCount:  0,
		CreatedAt:  createdAt,
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if created.ID != 7 {
		t.Fatalf("created id = %d, want 7", created.ID)
	}
	if !strings.HasPrefix(created.SessionNo, "S20260607") {
		t.Fatalf("session_no = %q, want S20260607 prefix", created.SessionNo)
	}

	sessionRows := sqlmock.NewRows([]string{
		"id", "session_no", "scenario_id", "user_id", "status", "turn_count", "created_at", "ended_at",
	}).AddRow(7, created.SessionNo, 1, 42, string(model.SessionStatusRunning), 1, createdAt, nil)
	messageRows := sqlmock.NewRows([]string{
		"id", "session_id", "role", "content", "stage", "created_at",
	}).AddRow(11, 7, string(model.MessageRoleUser), "hello", "自我介绍", createdAt)
	mock.ExpectQuery("SELECT (.+) FROM training_sessions WHERE id = \\?").
		WithArgs(7).
		WillReturnRows(sessionRows)
	mock.ExpectQuery("SELECT (.+) FROM messages WHERE session_id = \\?").
		WithArgs(7).
		WillReturnRows(messageRows)

	found, err := repo.FindByID(7)
	if err != nil {
		t.Fatalf("FindByID returned error: %v", err)
	}
	if found.ID != 7 {
		t.Fatalf("found id = %d, want 7", found.ID)
	}
	if len(found.Messages) != 1 {
		t.Fatalf("messages length = %d, want 1", len(found.Messages))
	}
	if found.Messages[0].Content != "hello" {
		t.Fatalf("message content = %q, want hello", found.Messages[0].Content)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestMySQLSessionRepositoryAppendTurnRollsBackWhenAIMessageInsertFails(t *testing.T) {
	db, mock, cleanup := newSQLMock(t)
	defer cleanup()
	repo := repository.NewMySQLSessionRepository(db)
	createdAt := time.Date(2026, 6, 7, 3, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO messages").
		WithArgs(9, string(model.MessageRoleUser), "hello", "自我介绍", createdAt).
		WillReturnResult(sqlmock.NewResult(101, 1))
	mock.ExpectExec("INSERT INTO messages").
		WithArgs(9, string(model.MessageRoleAI), "reply", "项目经历", createdAt).
		WillReturnError(errors.New("insert ai failed"))
	mock.ExpectRollback()

	_, err := repo.AppendTurn(
		9,
		model.Message{Role: model.MessageRoleUser, Content: "hello", Stage: "自我介绍", CreatedAt: createdAt},
		model.Message{Role: model.MessageRoleAI, Content: "reply", Stage: "项目经历", CreatedAt: createdAt},
	)
	if err == nil {
		t.Fatal("AppendTurn returned nil, want insert error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestMySQLSessionRepositoryAppendTurnReturnsFinishedWhenTurnUpdateAffectsNoRows(t *testing.T) {
	db, mock, cleanup := newSQLMock(t)
	defer cleanup()
	repo := repository.NewMySQLSessionRepository(db)
	createdAt := time.Date(2026, 6, 7, 3, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO messages").
		WithArgs(9, string(model.MessageRoleUser), "hello", "自我介绍", createdAt).
		WillReturnResult(sqlmock.NewResult(101, 1))
	mock.ExpectExec("INSERT INTO messages").
		WithArgs(9, string(model.MessageRoleAI), "reply", "项目经历", createdAt).
		WillReturnResult(sqlmock.NewResult(102, 1))
	mock.ExpectExec("UPDATE training_sessions SET turn_count = turn_count \\+ 1 WHERE id = \\? AND status <> \\?").
		WithArgs(9, string(model.SessionStatusFinished)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()
	mock.ExpectQuery("SELECT (.+) FROM training_sessions WHERE id = \\?").
		WithArgs(9).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "session_no", "scenario_id", "user_id", "status", "turn_count", "created_at", "ended_at",
		}).AddRow(9, "S202606070009", 1, 42, string(model.SessionStatusFinished), 1, createdAt, createdAt.Add(time.Minute)))

	_, err := repo.AppendTurn(
		9,
		model.Message{Role: model.MessageRoleUser, Content: "hello", Stage: "自我介绍", CreatedAt: createdAt},
		model.Message{Role: model.MessageRoleAI, Content: "reply", Stage: "项目经历", CreatedAt: createdAt},
	)
	if !errors.Is(err, repository.ErrSessionAlreadyFinished) {
		t.Fatalf("AppendTurn error = %v, want ErrSessionAlreadyFinished", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestMySQLScenarioRepositoryListParsesJSONFields(t *testing.T) {
	db, mock, cleanup := newSQLMock(t)
	defer cleanup()
	repo := repository.NewMySQLScenarioRepository(db)

	rows := sqlmock.NewRows([]string{
		"id", "code", "name", "description", "difficulty", "ai_role", "user_goal", "opening_message", "stages_json", "rubric_json",
	}).AddRow(
		1,
		"interview",
		"英语面试",
		"练习自我介绍",
		"medium",
		"技术面试官",
		"清晰介绍项目经历",
		"hello",
		`[{"name":"自我介绍","description":"介绍背景"}]`,
		`[{"name":"语法准确度","description":"语法是否准确"}]`,
	)
	mock.ExpectQuery("SELECT (.+) FROM scenarios ORDER BY id").WillReturnRows(rows)

	scenarios := repo.List()

	if len(scenarios) != 1 {
		t.Fatalf("scenarios length = %d, want 1", len(scenarios))
	}
	if scenarios[0].Code != "interview" {
		t.Fatalf("scenario code = %q, want interview", scenarios[0].Code)
	}
	if len(scenarios[0].Stages) != 1 || scenarios[0].Stages[0].Name != "自我介绍" {
		t.Fatalf("stages = %#v, want parsed stage", scenarios[0].Stages)
	}
	if len(scenarios[0].Rubric) != 1 || scenarios[0].Rubric[0].Name != "语法准确度" {
		t.Fatalf("rubric = %#v, want parsed rubric", scenarios[0].Rubric)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestMySQLFeedbackRepositoryFindsCorrectionAndCurrentScore(t *testing.T) {
	db, mock, cleanup := newSQLMock(t)
	defer cleanup()
	repo := repository.NewMySQLFeedbackRepository(db)

	mock.ExpectQuery("SELECT (.+) FROM corrections WHERE message_id = \\?").
		WithArgs(11).
		WillReturnRows(sqlmock.NewRows([]string{
			"message_id", "session_id", "original_text", "corrected_text", "errors_json", "better_expressions_json",
		}).AddRow(
			11,
			7,
			"I am study computer science.",
			"I am studying computer science.",
			`[{"type":"grammar","span":"am study","suggestion":"am studying","explanation":"be 动词后应接现在分词。"}]`,
			`["I major in computer science."]`,
		))

	correction, err := repo.FindCorrectionByMessageID(11)
	if err != nil {
		t.Fatalf("FindCorrectionByMessageID returned error: %v", err)
	}
	if correction.SessionID != 7 {
		t.Fatalf("correction session id = %d, want 7", correction.SessionID)
	}
	if len(correction.Errors) != 1 || correction.Errors[0].Suggestion != "am studying" {
		t.Fatalf("correction errors = %#v, want parsed error", correction.Errors)
	}
	if len(correction.BetterExpressions) != 1 {
		t.Fatalf("better expressions length = %d, want 1", len(correction.BetterExpressions))
	}

	mock.ExpectQuery("SELECT (.+) FROM scores WHERE session_id = \\? ORDER BY message_id DESC LIMIT 1").
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{
			"message_id", "session_id", "fluency", "grammar", "expression", "vocabulary", "completion", "total_score", "comment",
		}).AddRow(11, 7, 75, 72, 80, 76, 85, 77, "stable score"))

	score, err := repo.FindCurrentScoreBySessionID(7)
	if err != nil {
		t.Fatalf("FindCurrentScoreBySessionID returned error: %v", err)
	}
	if score.MessageID != 11 {
		t.Fatalf("score message id = %d, want 11", score.MessageID)
	}
	if score.TotalScore != 77 {
		t.Fatalf("score total = %d, want 77", score.TotalScore)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestMySQLReportRepositoryFindBySessionIDParsesJSONFields(t *testing.T) {
	db, mock, cleanup := newSQLMock(t)
	defer cleanup()
	repo := repository.NewMySQLReportRepository(db)
	createdAt := time.Date(2026, 6, 7, 3, 5, 0, 0, time.UTC)

	mock.ExpectQuery("SELECT (.+) FROM reports WHERE session_id = \\?").
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{
			"session_id", "scenario_id", "scenario_code", "scenario_name", "scenario_difficulty",
			"duration_seconds", "turn_count", "total_score", "scores_json", "summary",
			"major_problems_json", "frequent_errors_json", "better_expressions_json", "next_practice_plan_json", "created_at",
		}).AddRow(
			7,
			1,
			"interview",
			"英语面试",
			"medium",
			180,
			1,
			77,
			`{"message_id":11,"session_id":7,"fluency":75,"grammar":72,"expression":80,"vocabulary":76,"completion":85,"total_score":77,"comment":"stable score"}`,
			"summary",
			`["动词形式不稳定"]`,
			`["am study -> am studying"]`,
			`["I major in computer science."]`,
			`["用 STAR 结构重写项目经历回答。"]`,
			createdAt,
		))

	report, err := repo.FindBySessionID(7)
	if err != nil {
		t.Fatalf("FindBySessionID returned error: %v", err)
	}
	if report.SessionID != 7 {
		t.Fatalf("report session id = %d, want 7", report.SessionID)
	}
	if report.Scenario.Code != "interview" {
		t.Fatalf("scenario code = %q, want interview", report.Scenario.Code)
	}
	if report.Scores.TotalScore != 77 {
		t.Fatalf("score total = %d, want 77", report.Scores.TotalScore)
	}
	if len(report.NextPracticePlan) != 1 {
		t.Fatalf("next practice plan length = %d, want 1", len(report.NextPracticePlan))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func newSQLMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock, func()) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}

	return db, mock, func() {
		_ = db.Close()
	}
}
