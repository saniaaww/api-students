package route

import (
	"api-students/app/repository"
	"api-students/app/service"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Register(
	app *fiber.App,
	pool *pgxpool.Pool,
	repo repository.StudentRepository,
) {
	app.Get("/", service.Home)

	api := app.Group("/api/v1")

	api.Get("/health", func(c *fiber.Ctx) error {
		return service.Health(c, pool)
	})

	students := api.Group("/students")

	students.Get("/", func(c *fiber.Ctx) error {
		return service.GetStudents(c, repo)
	})

	students.Get("/:id", func(c *fiber.Ctx) error {
		return service.GetStudent(c, repo)
	})

	students.Post("/", func(c *fiber.Ctx) error {
		return service.CreateStudent(c, repo)
	})

	students.Put("/:id", func(c *fiber.Ctx) error {
		return service.ReplaceStudent(c, repo)
	})

	students.Patch("/:id", func(c *fiber.Ctx) error {
		return service.PatchStudent(c, repo)
	})

	students.Delete("/:id", func(c *fiber.Ctx) error {
		return service.DeleteStudent(c, repo)
	})
}
