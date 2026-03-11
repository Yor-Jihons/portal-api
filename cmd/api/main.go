package main

import (
	"os"

	"github.com/Yor-Jihons/portal-api/internal/routes"
)

func main() {
	r := routes.SetupRouter()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r.Run(":" + port)
}
