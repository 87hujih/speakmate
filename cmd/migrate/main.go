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

func main() {
	migrationsDir := flag.String("dir", "migrations", "directory containing ordered .sql migration files")
	timeoutSeconds := flag.Int("timeout", 60, "migration timeout in seconds")
	flag.Parse()

	cfg := config.Load()
	storage := config.StorageConfig{
		Mode:     config.StorageModeMySQL,
		MySQLDSN: cfg.Storage.MySQLDSN,
	}
	if err := storage.Validate(); err != nil {
		log.Fatalf("migration config invalid: %s", security.RedactString(err.Error()))
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*timeoutSeconds)*time.Second)
	defer cancel()

	db, err := database.OpenMySQL(ctx, storage)
	if err != nil {
		log.Fatalf("open mysql failed: %s", security.RedactString(err.Error()))
	}
	defer db.Close()

	migrations, err := database.LoadMigrations(*migrationsDir)
	if err != nil {
		log.Fatalf("load migrations failed: %s", security.RedactString(err.Error()))
	}
	if err := database.ApplyMigrations(ctx, db, migrations); err != nil {
		log.Fatalf("apply migrations failed: %s", security.RedactString(err.Error()))
	}

	log.Printf("applied %d migration files from %s", len(migrations), *migrationsDir)
}
