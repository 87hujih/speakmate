package main

import (
	"context"
	"flag"
	"log"
	"time"

	"speakmate/internal/config"
	"speakmate/internal/infra/database"
	"speakmate/internal/security"
)

// main 是当前命令的入口，负责串联配置加载和执行流程。
func main() {
	migrationsDir := flag.String("dir", "migrations", "按顺序存放 .sql 迁移文件的目录")
	timeoutSeconds := flag.Int("timeout", 60, "迁移超时时间，单位秒")
	flag.Parse()

	cfg := config.Load()
	storage := config.StorageConfig{
		Mode:     config.StorageModeMySQL,
		MySQLDSN: cfg.Storage.MySQLDSN,
	}
	if err := storage.Validate(); err != nil {
		log.Fatalf("迁移配置无效：%s", security.RedactString(err.Error()))
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*timeoutSeconds)*time.Second)
	defer cancel()

	db, err := database.OpenMySQL(ctx, storage)
	if err != nil {
		log.Fatalf("打开 MySQL 失败：%s", security.RedactString(err.Error()))
	}
	defer db.Close()

	migrations, err := database.LoadMigrations(*migrationsDir)
	if err != nil {
		log.Fatalf("加载迁移文件失败：%s", security.RedactString(err.Error()))
	}
	if err := database.ApplyMigrations(ctx, db, migrations); err != nil {
		log.Fatalf("执行迁移失败：%s", security.RedactString(err.Error()))
	}

	log.Printf("已执行 %d 个迁移文件，目录 %s", len(migrations), *migrationsDir)
}
