package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"api-students/app/repository"
	"api-students/database"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

var metodeBerbody = map[string]bool{
	fiber.MethodPost:  true,
	fiber.MethodPut:   true,
	fiber.MethodPatch: true,
}

func requireJSON(c *fiber.Ctx) error {
	if metodeBerbody[c.Method()] {
		ct := c.Get("Content-Type")

		if !strings.HasPrefix(ct, fiber.MIMEApplicationJSON) {
			return errorResponse(
				c,
				fiber.StatusUnsupportedMediaType,
				"Content-Type harus application/json",
				nil,
			)
		}
	}

	return c.Next()
}

func main() {

	// Koneksi ke PostgreSQL
	pool, err := database.ConnectDB()
	if err != nil {
		log.Fatal("Gagal terhubung ke database:", err)
	}
	defer pool.Close()

	// Membuat repository untuk mengakses data student
	studentRepo := repository.NewStudentRepository(pool)

	app := fiber.New(fiber.Config{
		AppName: "API Students - Praktikum Backend Lanjut",

		ErrorHandler: func(c *fiber.Ctx, err error) error {
			status := fiber.StatusInternalServerError
			message := "terjadi kesalahan pada server"

			if e, ok := err.(*fiber.Error); ok {
				status = e.Code
				message = e.Message
			}

			return errorResponse(c, status, message, nil)
		},
	})

	// Middleware global
	app.Use(requestid.New())

	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${locals:requestid} ${method} ${path} ${status} ${latency}\n",
	}))

	app.Use(cors.New())

	// Endpoint utama
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("API Students berjalan!")
	})

	// API version
	api := app.Group("/api/v1")

	// Health check
	api.Get("/health", func(c *fiber.Ctx) error {

		// Mengecek koneksi PostgreSQL
		if err := pool.Ping(c.Context()); err != nil {
			return errorResponse(
				c,
				fiber.StatusServiceUnavailable,
				"database tidak dapat diakses",
				nil,
			)
		}

		return successResponse(
			c,
			fiber.StatusOK,
			"server dan database berjalan",
			fiber.Map{
				"timestamp": time.Now(),
			},
			nil,
		)
	})

	// Group students
	studentsAPI := api.Group("/students", requireJSON)

	// Endpoint mahasiswa
	studentsAPI.Get("/", func(c *fiber.Ctx) error {
		return getStudents(c, studentRepo)
	})

	studentsAPI.Get("/:id", func(c *fiber.Ctx) error {
		return getStudent(c, studentRepo)
	})

	studentsAPI.Post("/", func(c *fiber.Ctx) error {
		return createStudent(c, studentRepo)
	})

	studentsAPI.Put("/:id", func(c *fiber.Ctx) error {
		return replaceStudent(c, studentRepo)
	})

	studentsAPI.Patch("/:id", func(c *fiber.Ctx) error {
		return patchStudent(c, studentRepo)
	})

	studentsAPI.Delete("/:id", func(c *fiber.Ctx) error {
		return deleteStudent(c, studentRepo)
	})

	// Endpoint tidak ditemukan
	app.Use(func(c *fiber.Ctx) error {
		return errorResponse(
			c,
			fiber.StatusNotFound,
			"endpoint tidak ditemukan",
			nil,
		)
	})

	fmt.Println("Server berjalan di http://localhost:3000")

	log.Fatal(app.Listen(":3000"))
}
