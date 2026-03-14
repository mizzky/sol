//go:build integration

package tests

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// テスト全体で共有する DB 接続
var testDB *sql.DB

func TestMain(m *testing.M) {
	ctx := context.Background()

	// 1. PostgreSQL コンテナを起動
	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:17-trixie",
		tcpostgres.WithDatabase("test_db"),
		tcpostgres.WithUsername("user"),
		tcpostgres.WithPassword("password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2),
		),
	)
	if err != nil {
		log.Fatalf("コンテナ起動失敗: %v", err)
	}

	// 2. 接続文字列を取得
	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("接続文字列取得失敗: %v", err)
	}

	// 3. DB 接続
	testDB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("DB 接続失敗: %v", err)
	}

	// 4. マイグレーション適用
	// go test はパッケージディレクトリ（backend/tests）を作業ディレクトリにする
	wd, _ := os.Getwd()
	migrationsPath := filepath.Join(wd, "..", "db", "migrations")

	migrator, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		connStr,
	)
	if err != nil {
		log.Fatalf("マイグレーション初期化失敗: %v", err)
	}
	if err := migrator.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("マイグレーション適用失敗: %v", err)
	}

	// 5. テスト実行
	code := m.Run()

	// 6. クリーンアップ（defer は os.Exit 前に実行されないため明示的に呼ぶ）
	testDB.Close()
	pgContainer.Terminate(ctx)

	os.Exit(code)
}

func TestIntegration_DBReady(t *testing.T) {
	if testDB == nil {
		t.Fatalf("testDB is nil")
	}
	if err := testDB.Ping(); err != nil {
		t.Fatalf("test DB ping failed: %v", err)
	}

	var currentDB string
	if err := testDB.QueryRow(`SELECT current_database()`).Scan(&currentDB); err != nil {
		t.Fatalf("failed to query current_database(): %v", err)
	}
	if currentDB != "test_db" {
		t.Fatalf("unexpected database name: got=%s wnat=test_db", currentDB)
	}
}
