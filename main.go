package main

import (
	"fmt"
	"log"

	"api-students/app/repository"
	"api-students/config"
	"api-students/database"
	"api-students/route"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

func main() {
	// Koneksi ke PostgreSQL
	pool, err := database.ConnectDB()
	if err != nil {
		log.Fatal("Gagal terhubung ke database:", err)
	}
	defer pool.Close()

	// Membuat repository
	studentRepo := repository.NewStudentRepository(pool)

	// Membuat aplikasi Fiber
	app := fiber.New(fiber.Config{
		AppName: "API Students - Praktikum Backend Lanjut",
	})

	// Middleware global
	app.Use(requestid.New())
	app.Use(config.NewLogger())
	app.Use(cors.New())

	// Registrasi seluruh route
	route.Register(app, pool, studentRepo)

	fmt.Println("Server berjalan di http://localhost:3000")

	log.Fatal(app.Listen(":3000"))
}