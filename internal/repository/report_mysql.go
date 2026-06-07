package repository

import (
	"database/sql"

	"speakmate/internal/model"
)

// MySQLReportRepository 使用 MySQL 保存课后报告。
type MySQLReportRepository struct {
	db *sql.DB
}

// NewMySQLReportRepository 创建 MySQL Report 仓库。
func NewMySQLReportRepository(db *sql.DB) *MySQLReportRepository {
	return &MySQLReportRepository{db: db}
}

// Save 按 session_id 保存或覆盖报告。
func (r *MySQLReportRepository) Save(report model.Report) error {
	scoresJSON, err := marshalJSON(report.Scores)
	if err != nil {
		return err
	}
	majorProblemsJSON, err := marshalJSON(report.MajorProblems)
	if err != nil {
		return err
	}
	frequentErrorsJSON, err := marshalJSON(report.FrequentErrors)
	if err != nil {
		return err
	}
	betterExpressionsJSON, err := marshalJSON(report.BetterExpressions)
	if err != nil {
		return err
	}
	nextPracticePlanJSON, err := marshalJSON(report.NextPracticePlan)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(
		`INSERT INTO reports (
session_id, scenario_id, scenario_code, scenario_name, scenario_difficulty,
duration_seconds, turn_count, total_score, scores_json, summary,
major_problems_json, frequent_errors_json, better_expressions_json, next_practice_plan_json, created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
scenario_id = VALUES(scenario_id),
scenario_code = VALUES(scenario_code),
scenario_name = VALUES(scenario_name),
scenario_difficulty = VALUES(scenario_difficulty),
duration_seconds = VALUES(duration_seconds),
turn_count = VALUES(turn_count),
total_score = VALUES(total_score),
scores_json = VALUES(scores_json),
summary = VALUES(summary),
major_problems_json = VALUES(major_problems_json),
frequent_errors_json = VALUES(frequent_errors_json),
better_expressions_json = VALUES(better_expressions_json),
next_practice_plan_json = VALUES(next_practice_plan_json),
created_at = VALUES(created_at),
updated_at = CURRENT_TIMESTAMP`,
		report.SessionID,
		report.Scenario.ID,
		report.Scenario.Code,
		report.Scenario.Name,
		report.Scenario.Difficulty,
		report.DurationSeconds,
		report.TurnCount,
		report.TotalScore,
		scoresJSON,
		report.Summary,
		majorProblemsJSON,
		frequentErrorsJSON,
		betterExpressionsJSON,
		nextPracticePlanJSON,
		report.CreatedAt,
	)

	return err
}

// FindBySessionID 按 session_id 查询报告。
func (r *MySQLReportRepository) FindBySessionID(sessionID int) (model.Report, error) {
	row := r.db.QueryRow(
		`SELECT session_id, scenario_id, scenario_code, scenario_name, scenario_difficulty, duration_seconds, turn_count, total_score, scores_json, summary, major_problems_json, frequent_errors_json, better_expressions_json, next_practice_plan_json, created_at FROM reports WHERE session_id = ?`,
		sessionID,
	)
	report, err := scanReport(row)
	if err != nil {
		return model.Report{}, notFoundFromNoRows(err, ErrReportNotFound)
	}

	return report, nil
}

func scanReport(row scanner) (model.Report, error) {
	var report model.Report
	var scoresJSON string
	var majorProblemsJSON string
	var frequentErrorsJSON string
	var betterExpressionsJSON string
	var nextPracticePlanJSON string
	if err := row.Scan(
		&report.SessionID,
		&report.Scenario.ID,
		&report.Scenario.Code,
		&report.Scenario.Name,
		&report.Scenario.Difficulty,
		&report.DurationSeconds,
		&report.TurnCount,
		&report.TotalScore,
		&scoresJSON,
		&report.Summary,
		&majorProblemsJSON,
		&frequentErrorsJSON,
		&betterExpressionsJSON,
		&nextPracticePlanJSON,
		&report.CreatedAt,
	); err != nil {
		return model.Report{}, err
	}
	if err := unmarshalJSON(scoresJSON, &report.Scores); err != nil {
		return model.Report{}, err
	}
	if err := unmarshalStringList(majorProblemsJSON, &report.MajorProblems); err != nil {
		return model.Report{}, err
	}
	if err := unmarshalStringList(frequentErrorsJSON, &report.FrequentErrors); err != nil {
		return model.Report{}, err
	}
	if err := unmarshalStringList(betterExpressionsJSON, &report.BetterExpressions); err != nil {
		return model.Report{}, err
	}
	if err := unmarshalStringList(nextPracticePlanJSON, &report.NextPracticePlan); err != nil {
		return model.Report{}, err
	}

	return report, nil
}

func unmarshalStringList(raw string, values *[]string) error {
	if err := unmarshalJSON(raw, values); err != nil {
		return err
	}
	if *values == nil {
		*values = []string{}
	}

	return nil
}
