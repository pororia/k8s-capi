package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	action := flag.String("action", "up", "Migration action: up | down | status")
	configPath := flag.String("config", "./configs/config.yaml", "Config file path")
	flag.Parse()

	// 설정 로드 (config 파일 없어도 환경변수로 동작)
	viper.SetConfigFile(*configPath)
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Warning: config file not found, using env vars: %v\n", err)
	}

	// DB 연결
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		viper.GetString("database.host"),
		viper.GetInt("database.port"),
		viper.GetString("database.user"),
		viper.GetString("database.password"),
		viper.GetString("database.name"),
		viper.GetString("database.ssl_mode"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		os.Exit(1)
	}

	// 마이그레이션 테이블 생성
	if err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version VARCHAR(255) PRIMARY KEY,
		applied_at TIMESTAMP NOT NULL DEFAULT NOW()
	)`).Error; err != nil {
		fmt.Printf("Failed to create migrations table: %v\n", err)
		os.Exit(1)
	}

	switch *action {
	case "up":
		runMigrations(db, "up")
	case "down":
		rollbackMigration(db)
	case "status":
		showStatus(db)
	default:
		fmt.Printf("Unknown action: %s\n", *action)
		os.Exit(1)
	}
}

func getMigrationFiles(direction string) ([]string, error) {
	pattern := fmt.Sprintf("migrations/*.%s.sql", direction)
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}

func runMigrations(db *gorm.DB, direction string) {
	files, err := getMigrationFiles(direction)
	if err != nil {
		fmt.Printf("Failed to list migration files: %v\n", err)
		os.Exit(1)
	}

	for _, file := range files {
		version := filepath.Base(file)
		version = strings.TrimSuffix(version, fmt.Sprintf(".%s.sql", direction))

		// 이미 적용된 마이그레이션인지 확인
		var count int64
		db.Raw("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", version).Scan(&count)
		if count > 0 {
			fmt.Printf("  SKIP: %s (already applied)\n", version)
			continue
		}

		// SQL 파일 읽기 및 실행
		content, err := os.ReadFile(file)
		if err != nil {
			fmt.Printf("  ERROR: failed to read %s: %v\n", file, err)
			os.Exit(1)
		}

		if err := db.Exec(string(content)).Error; err != nil {
			fmt.Printf("  ERROR: failed to run %s: %v\n", file, err)
			os.Exit(1)
		}

		// 적용 기록
		db.Exec("INSERT INTO schema_migrations (version) VALUES (?)", version)
		fmt.Printf("  OK: %s\n", version)
	}
	fmt.Println("Migrations completed.")
}

func rollbackMigration(db *gorm.DB) {
	// 가장 최근 적용된 마이그레이션 조회
	var version string
	db.Raw("SELECT version FROM schema_migrations ORDER BY applied_at DESC LIMIT 1").Scan(&version)
	if version == "" {
		fmt.Println("No migrations to rollback.")
		return
	}

	file := fmt.Sprintf("migrations/%s.down.sql", version)
	content, err := os.ReadFile(file)
	if err != nil {
		fmt.Printf("Failed to read rollback file %s: %v\n", file, err)
		os.Exit(1)
	}

	if err := db.Exec(string(content)).Error; err != nil {
		fmt.Printf("Failed to rollback %s: %v\n", version, err)
		os.Exit(1)
	}

	db.Exec("DELETE FROM schema_migrations WHERE version = ?", version)
	fmt.Printf("Rolled back: %s\n", version)
}

func showStatus(db *gorm.DB) {
	files, _ := getMigrationFiles("up")

	var applied []string
	db.Raw("SELECT version FROM schema_migrations ORDER BY version").Scan(&applied)

	appliedSet := make(map[string]bool)
	for _, v := range applied {
		appliedSet[v] = true
	}

	fmt.Println("Migration Status:")
	for _, file := range files {
		version := strings.TrimSuffix(filepath.Base(file), ".up.sql")
		status := "pending"
		if appliedSet[version] {
			status = "applied"
		}
		fmt.Printf("  [%s] %s\n", status, version)
	}
}
