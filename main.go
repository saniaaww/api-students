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
	"github.com/joho/godotenv"
)

func main() {
	// Membaca file .env
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: file .env tidak ditemukan")
	}

	// Membaca konfigurasi environment
	env := config.LoadEnv()

	// Koneksi ke PostgreSQL
	pool, err := database.ConnectDB()
	if err != nil {
		log.Fatal("Gagal terhubung ke database:", err)
	}
	defer pool.Close()

	// Membuat repository
	studentRepo := repository.NewStudentRepository(pool)
	userRepo := repository.NewUserRepository(pool)
	refreshRepo := repository.NewRefreshTokenRepository(pool)

	// Membuat aplikasi Fiber
	app := fiber.New(fiber.Config{
		AppName: "API Students - Praktikum Backend Lanjut",
	})

	// Middleware global
	app.Use(requestid.New())
	app.Use(config.NewLogger())
	app.Use(cors.New())

	// Registrasi seluruh route
	route.Register(app, pool, studentRepo, userRepo, env.JWTSecret, refreshRepo)

	fmt.Println("Server berjalan di http://localhost:3000")

	log.Fatal(app.Listen(":3000"))
}
