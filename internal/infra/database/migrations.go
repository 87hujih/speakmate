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

// Migration 表示一个待执行的 SQL 迁移文件。
type Migration struct {
	Name string
	SQL  string
}

// LoadMigrations 按文件名字典序读取目录中的 .sql 迁移文件。
func LoadMigrations(dir string) ([]Migration, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("读取迁移目录 %q 失败：%w", dir, err)
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
		return nil, fmt.Errorf("迁移目录 %q 中未找到 .sql 文件", dir)
	}

	migrations := make([]Migration, 0, len(names))
	for _, name := range names {
		content, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, fmt.Errorf("读取迁移文件 %q 失败：%w", name, err)
		}
		migrations = append(migrations, Migration{
			Name: name,
			SQL:  string(content),
		})
	}

	return migrations, nil
}

// ApplyMigrations 按顺序执行所有 SQL 语句。
// 当前迁移通过 IF NOT EXISTS 和 ON DUPLICATE KEY UPDATE 保持幂等。
func ApplyMigrations(ctx context.Context, db *sql.DB, migrations []Migration) error {
	if db == nil {
		return errors.New("数据库连接不能为空")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	for _, migration := range migrations {
		statements := SplitSQLStatements(migration.SQL)
		for index, statement := range statements {
			if _, err := db.ExecContext(ctx, statement); err != nil {
				return fmt.Errorf("执行迁移 %s 的第 %d 条语句失败：%w", migration.Name, index+1, err)
			}
		}
	}

	return nil
}

// SplitSQLStatements 将迁移 SQL 拆成可执行语句，并忽略行注释和字符串中的分号。
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
