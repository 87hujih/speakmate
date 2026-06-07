CREATE TABLE IF NOT EXISTS scenarios (
  id BIGINT NOT NULL PRIMARY KEY,
  code VARCHAR(64) NOT NULL UNIQUE,
  name VARCHAR(128) NOT NULL,
  description VARCHAR(512) NOT NULL,
  difficulty VARCHAR(32) NOT NULL,
  ai_role VARCHAR(128) NOT NULL,
  user_goal TEXT NOT NULL,
  opening_message TEXT NOT NULL,
  stages_json JSON NOT NULL,
  rubric_json JSON NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS training_sessions (
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  session_no VARCHAR(64) NOT NULL UNIQUE,
  scenario_id BIGINT NOT NULL,
  user_id BIGINT NOT NULL DEFAULT 1,
  status VARCHAR(32) NOT NULL,
  turn_count INT NOT NULL DEFAULT 0,
  created_at DATETIME(6) NOT NULL,
  ended_at DATETIME(6) NULL,
  INDEX idx_training_sessions_user_created (user_id, created_at),
  INDEX idx_training_sessions_scenario (scenario_id),
  CONSTRAINT fk_training_sessions_scenario
    FOREIGN KEY (scenario_id) REFERENCES scenarios(id)
    ON UPDATE CASCADE
    ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS messages (
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  session_id BIGINT NOT NULL,
  role VARCHAR(16) NOT NULL,
  content TEXT NOT NULL,
  stage VARCHAR(128) NOT NULL,
  created_at DATETIME(6) NOT NULL,
  INDEX idx_messages_session_id (session_id, id),
  CONSTRAINT fk_messages_session
    FOREIGN KEY (session_id) REFERENCES training_sessions(id)
    ON UPDATE CASCADE
    ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS corrections (
  message_id BIGINT NOT NULL PRIMARY KEY,
  session_id BIGINT NOT NULL,
  original_text TEXT NOT NULL,
  corrected_text TEXT NOT NULL,
  errors_json JSON NOT NULL,
  better_expressions_json JSON NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_corrections_session (session_id, message_id),
  CONSTRAINT fk_corrections_message
    FOREIGN KEY (message_id) REFERENCES messages(id)
    ON UPDATE CASCADE
    ON DELETE CASCADE,
  CONSTRAINT fk_corrections_session
    FOREIGN KEY (session_id) REFERENCES training_sessions(id)
    ON UPDATE CASCADE
    ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS scores (
  message_id BIGINT NOT NULL PRIMARY KEY,
  session_id BIGINT NOT NULL,
  fluency INT NOT NULL,
  grammar INT NOT NULL,
  expression INT NOT NULL,
  vocabulary INT NOT NULL,
  completion INT NOT NULL,
  total_score INT NOT NULL,
  comment TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_scores_session (session_id, message_id),
  CONSTRAINT fk_scores_message
    FOREIGN KEY (message_id) REFERENCES messages(id)
    ON UPDATE CASCADE
    ON DELETE CASCADE,
  CONSTRAINT fk_scores_session
    FOREIGN KEY (session_id) REFERENCES training_sessions(id)
    ON UPDATE CASCADE
    ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS reports (
  session_id BIGINT NOT NULL PRIMARY KEY,
  scenario_id BIGINT NOT NULL,
  scenario_code VARCHAR(64) NOT NULL,
  scenario_name VARCHAR(128) NOT NULL,
  scenario_difficulty VARCHAR(32) NOT NULL,
  duration_seconds INT NOT NULL,
  turn_count INT NOT NULL,
  total_score INT NOT NULL,
  scores_json JSON NOT NULL,
  summary TEXT NOT NULL,
  major_problems_json JSON NOT NULL,
  frequent_errors_json JSON NOT NULL,
  better_expressions_json JSON NOT NULL,
  next_practice_plan_json JSON NOT NULL,
  created_at DATETIME(6) NOT NULL,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_reports_scenario (scenario_id),
  CONSTRAINT fk_reports_session
    FOREIGN KEY (session_id) REFERENCES training_sessions(id)
    ON UPDATE CASCADE
    ON DELETE CASCADE,
  CONSTRAINT fk_reports_scenario
    FOREIGN KEY (scenario_id) REFERENCES scenarios(id)
    ON UPDATE CASCADE
    ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
