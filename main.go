package main

import (
	"fmt"
	"log"
	"strings"
	"time"

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
		return successResponse(
			c,
			fiber.StatusOK,
			"server berjalan",
			fiber.Map{
				"timestamp": time.Now(),
			},
			nil,
		)
	})

	// Group students
	studentsAPI := api.Group("/students", requireJSON)

	// GET daftar mahasiswa
	studentsAPI.Get("/", getStudents)

	// GET satu mahasiswa
	studentsAPI.Get("/:id", getStudent)

	// POST mahasiswa
	studentsAPI.Post("/", createStudent)

	// PUT mahasiswa
	studentsAPI.Put("/:id", replaceStudent)

	// PATCH mahasiswa
	studentsAPI.Patch("/:id", patchStudent)

	// DELETE mahasiswa
	studentsAPI.Delete("/:id", deleteStudent)

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
