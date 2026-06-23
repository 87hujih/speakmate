package repository

import (
	"database/sql"

	"speakmate/internal/model"
)

// MySQLScenarioRepository 使用 MySQL 提供场景查询能力。
type MySQLScenarioRepository struct {
	db *sql.DB
}

// NewMySQLScenarioRepository 创建 MySQL Scenario 仓库。
func NewMySQLScenarioRepository(db *sql.DB) *MySQLScenarioRepository {
	return &MySQLScenarioRepository{db: db}
}

// List 返回全部训练场景。
func (r *MySQLScenarioRepository) List() []model.Scenario {
	rows, err := r.db.Query(`SELECT id, code, name, description, difficulty, ai_role, user_goal, opening_message, stages_json, rubric_json FROM scenarios ORDER BY id`)
	if err != nil {
		return []model.Scenario{}
	}
	defer rows.Close()

	scenarios := []model.Scenario{}
	for rows.Next() {
		scenario, err := scanScenario(rows)
		if err != nil {
			return []model.Scenario{}
		}
		scenarios = append(scenarios, scenario)
	}
	if err := rows.Err(); err != nil {
		return []model.Scenario{}
	}

	return scenarios
}

// FindByID 按数字 ID 查询单个训练场景。
func (r *MySQLScenarioRepository) FindByID(id int) (model.Scenario, error) {
	row := r.db.QueryRow(
		`SELECT id, code, name, description, difficulty, ai_role, user_goal, opening_message, stages_json, rubric_json FROM scenarios WHERE id = ?`,
		id,
	)
	scenario, err := scanScenario(row)
	if err != nil {
		return model.Scenario{}, notFoundFromNoRows(err, ErrScenarioNotFound)
	}

	return scenario, nil
}

// scanScenario 从数据库行读取训练场景。
func scanScenario(row scanner) (model.Scenario, error) {
	var scenario model.Scenario
	var stagesJSON string
	var rubricJSON string
	if err := row.Scan(
		&scenario.ID,
		&scenario.Code,
		&scenario.Name,
		&scenario.Description,
		&scenario.Difficulty,
		&scenario.AIRole,
		&scenario.UserGoal,
		&scenario.OpeningMessage,
		&stagesJSON,
		&rubricJSON,
	); err != nil {
		return model.Scenario{}, err
	}
	if err := unmarshalJSON(stagesJSON, &scenario.Stages); err != nil {
		return model.Scenario{}, err
	}
	if scenario.Stages == nil {
		scenario.Stages = []model.ScenarioStage{}
	}
	if err := unmarshalJSON(rubricJSON, &scenario.Rubric); err != nil {
		return model.Scenario{}, err
	}
	if scenario.Rubric == nil {
		scenario.Rubric = []model.ScenarioRubric{}
	}

	return scenario, nil
}
