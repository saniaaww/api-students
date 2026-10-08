package main

import (
	"fmt"
	"log"

	"api-students/app/repository"
	"api-students/config"
	"api-students/database"
	"api-students/helper"
	"api-students/route"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/joho/godotenv"
)

func main() {

	// =====================================================
	// LOAD ENV
	// =====================================================

	if err := godotenv.Load(); err != nil {
		log.Println("Warning: file .env tidak ditemukan")
	}

	env := config.LoadEnv()

	// =====================================================
	// DATABASE
	// =====================================================

	pool, err := database.ConnectDB()
	if err != nil {
		log.Fatal("Gagal terhubung ke database:", err)
	}
	defer pool.Close()

	// =====================================================
	// REPOSITORY
	// =====================================================

	studentRepo := repository.NewStudentRepository(pool)
	userRepo := repository.NewUserRepository(pool)
	courseRepo := repository.NewCourseRepository(pool)
	enrollmentRepo := repository.NewEnrollmentRepository(pool)
	refreshRepo := repository.NewRefreshTokenRepository(pool)

	// =====================================================
	// FIBER APPLICATION
	// =====================================================

	app := fiber.New(fiber.Config{
		AppName:      "SIAKAD Mini API",
		ErrorHandler: helper.ErrorHandler,
	})

	// =====================================================
	// GLOBAL MIDDLEWARE
	// =====================================================

	app.Use(requestid.New())
	app.Use(config.NewLogger())
	app.Use(cors.New())

	// =====================================================
	// ROUTES
	// =====================================================

	route.Register(
		app,
		pool,
		studentRepo,
		userRepo,
		courseRepo,
		enrollmentRepo,
		env.JWTSecret,
		refreshRepo,
	)

	// =====================================================
	// 404 HANDLER
	// =====================================================

	app.Use(func(c *fiber.Ctx) error {
		return helper.NotFound("endpoint tidak ditemukan")
	})

	// =====================================================
	// START SERVER
	// =====================================================

	fmt.Println("SIAKAD Mini API berjalan di http://localhost:3000")

	log.Fatal(app.Listen(":3000"))
}