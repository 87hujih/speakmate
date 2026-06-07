package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"speakmate/internal/config"
)

// OpenMySQL 按 MySQL 存储配置初始化数据库连接。
func OpenMySQL(ctx context.Context, storage config.StorageConfig) (*sql.DB, error) {
	if err := storage.Validate(); err != nil {
		return nil, err
	}
	if !storage.IsMySQL() {
		return nil, fmt.Errorf("storage mode %q is not mysql", storage.Mode)
	}

	db, err := sql.Open("mysql", storage.MySQLDSN)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	pingCtx := ctx
	if pingCtx == nil {
		pingCtx = context.Background()
	}
	pingCtx, cancel := context.WithTimeout(pingCtx, 5*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}
