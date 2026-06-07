package database

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSplitSQLStatementsIgnoresCommentsAndKeepsSemicolonsInsideStrings(t *testing.T) {
	sql := `
-- create table first
CREATE TABLE IF NOT EXISTS examples (
  id BIGINT PRIMARY KEY,
  text_value TEXT NOT NULL
);

INSERT INTO examples (id, text_value) VALUES (1, 'keep;inside');
`

	got := SplitSQLStatements(sql)

	if len(got) != 2 {
		t.Fatalf("statement count = %d, want 2: %#v", len(got), got)
	}
	if got[0] != "CREATE TABLE IF NOT EXISTS examples (\n  id BIGINT PRIMARY KEY,\n  text_value TEXT NOT NULL\n)" {
		t.Fatalf("first statement = %q", got[0])
	}
	if got[1] != "INSERT INTO examples (id, text_value) VALUES (1, 'keep;inside')" {
		t.Fatalf("second statement = %q", got[1])
	}
}

func TestLoadMigrationsReturnsSQLFilesInNameOrder(t *testing.T) {
	dir := t.TempDir()
	writeMigrationFile(t, dir, "002_second.sql", "SELECT 2;")
	writeMigrationFile(t, dir, "001_first.sql", "SELECT 1;")
	writeMigrationFile(t, dir, "notes.txt", "ignored")

	got, err := LoadMigrations(dir)
	if err != nil {
		t.Fatalf("LoadMigrations returned error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("migration count = %d, want 2", len(got))
	}
	if got[0].Name != "001_first.sql" || got[0].SQL != "SELECT 1;" {
		t.Fatalf("first migration = %#v, want 001_first.sql", got[0])
	}
	if got[1].Name != "002_second.sql" || got[1].SQL != "SELECT 2;" {
		t.Fatalf("second migration = %#v, want 002_second.sql", got[1])
	}
}

func writeMigrationFile(t *testing.T, dir string, name string, content string) {
	t.Helper()

	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
}
