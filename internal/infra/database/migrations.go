package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Migration struct {
	Name string
	SQL  string
}

// LoadMigrations reads .sql files from dir in lexical order.
func LoadMigrations(dir string) ([]Migration, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read migrations dir %q: %w", dir, err)
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		names = append(names, entry.Name())
	}
	sort.Strings(names)
	if len(names) == 0 {
		return nil, fmt.Errorf("no .sql migrations found in %q", dir)
	}

	migrations := make([]Migration, 0, len(names))
	for _, name := range names {
		content, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, fmt.Errorf("read migration %q: %w", name, err)
		}
		migrations = append(migrations, Migration{
			Name: name,
			SQL:  string(content),
		})
	}

	return migrations, nil
}

// ApplyMigrations executes all SQL statements in order. The current migrations
// are idempotent through IF NOT EXISTS and ON DUPLICATE KEY UPDATE.
func ApplyMigrations(ctx context.Context, db *sql.DB, migrations []Migration) error {
	if db == nil {
		return errors.New("database connection is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	for _, migration := range migrations {
		statements := SplitSQLStatements(migration.SQL)
		for index, statement := range statements {
			if _, err := db.ExecContext(ctx, statement); err != nil {
				return fmt.Errorf("apply migration %s statement %d failed: %w", migration.Name, index+1, err)
			}
		}
	}

	return nil
}

// SplitSQLStatements splits migration SQL into executable statements while
// ignoring line comments and semicolons inside quoted strings.
func SplitSQLStatements(sqlText string) []string {
	var statements []string
	var current strings.Builder
	inSingleQuote := false
	inDoubleQuote := false
	inBacktick := false

	for i := 0; i < len(sqlText); i++ {
		ch := sqlText[i]
		next := byte(0)
		if i+1 < len(sqlText) {
			next = sqlText[i+1]
		}

		if !inSingleQuote && !inDoubleQuote && !inBacktick && ch == '-' && next == '-' {
			for i < len(sqlText) && sqlText[i] != '\n' {
				i++
			}
			continue
		}

		switch ch {
		case '\'':
			current.WriteByte(ch)
			if inSingleQuote && next == '\'' {
				current.WriteByte(next)
				i++
				continue
			}
			if !inDoubleQuote && !inBacktick {
				inSingleQuote = !inSingleQuote
			}
			continue
		case '"':
			current.WriteByte(ch)
			if !inSingleQuote && !inBacktick {
				inDoubleQuote = !inDoubleQuote
			}
			continue
		case '`':
			current.WriteByte(ch)
			if !inSingleQuote && !inDoubleQuote {
				inBacktick = !inBacktick
			}
			continue
		case ';':
			if !inSingleQuote && !inDoubleQuote && !inBacktick {
				statement := strings.TrimSpace(current.String())
				if statement != "" {
					statements = append(statements, statement)
				}
				current.Reset()
				continue
			}
		}

		current.WriteByte(ch)
	}

	statement := strings.TrimSpace(current.String())
	if statement != "" {
		statements = append(statements, statement)
	}

	return statements
}
