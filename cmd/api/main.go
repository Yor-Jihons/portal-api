package main

import (
	"os"

	"github.com/Yor-Jihons/portal-api/internal/routes"
	"github.com/joho/godotenv"
)

func main() {
	// .envファイルを読み込む (ファイルがなくてもエラーにしない)
	_ = godotenv.Load()

	r := routes.SetupRouter()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
