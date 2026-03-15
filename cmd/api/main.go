package main

import (
	"log"
	"os"

	"github.com/Yor-Jihons/portal-api/internal/db"
	"github.com/Yor-Jihons/portal-api/internal/routes"
	"github.com/joho/godotenv"
)

func main() {
	// .envファイルを読み込む
	err1 := godotenv.Load()
	if err1 != nil {
		_ = godotenv.Load("../../.env")
	}

	// DB接続
	if err2 := db.InitDB(); err2 != nil {
		log.Fatalf("DB接続失敗: %v", err2)
	}

	// ルーターのセットアップ (DB接続を渡す)
	r := routes.SetupRouter(db.DB)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
