package main

import (
	"log"
	"os"

	"github.com/Yor-Jihons/portal-api/internal/db"
	"github.com/Yor-Jihons/portal-api/internal/routes"
	"github.com/joho/godotenv"
)

func main() {
	// .envファイルを読み込む (ファイルがなくてもエラーにしない)
	err1 := godotenv.Load()
	if err1 != nil {
		// もしカレントディレクトリで見つからない場合、1つ上の階層も探してみる (air対策)
		_ = godotenv.Load("../../.env")
	}

	// cmd/api/main.go の main関数内に追加
	if err2 := db.InitDB(); err2 != nil {
		log.Fatalf("DB接続失敗: %v", err2)
	}

	// デバッグ用：URLが空でないか確認
	if os.Getenv("TURSO_DATABASE_URL") == "" {
		log.Fatal("エラー: TURSO_DATABASE_URL が設定されていません。")
	}

	r := routes.SetupRouter()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
