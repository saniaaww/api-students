package route

import (
	"time"

	"api-students/app/repository"
	"api-students/app/service"
	"api-students/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Register(
	app *fiber.App,
	pool *pgxpool.Pool,
	repo repository.StudentRepository,
	userRepo repository.UserRepository,
	jwtSecret string,
	refreshRepo repository.RefreshTokenRepository,
) {
	app.Get("/", service.Home)

	api := app.Group("/api/v1")

	api.Get("/health", func(c *fiber.Ctx) error {
		return service.Health(c, pool)
	})

	auth := api.Group("/auth")

	auth.Post("/register", service.RegisterHandler(userRepo))

	// Rate limiter untuk login:
	// maksimal 5 percobaan dalam 1 menit.
	// Percobaan ke-6 akan mendapatkan 429.
	loginLimiter := limiter.New(limiter.Config{
		Max:        5,
		Expiration: 1 * time.Minute,
		LimitReached: func(c *fiber.Ctx) error {
			c.Set("Retry-After", "60")

			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"message": "terlalu banyak percobaan login, coba lagi nanti",
			})
		},
	})

	auth.Post(
		"/login",
		loginLimiter,
		service.LoginHandler(
			userRepo,
			refreshRepo,
			jwtSecret,
		),
	)

	auth.Post("/refresh", service.RefreshHandler(
		userRepo,
		refreshRepo,
		jwtSecret,
	))

	auth.Post("/logout", service.LogoutHandler(refreshRepo))

	auth.Get(
		"/me",
		middleware.RequireAuth(jwtSecret),
		service.MeHandler(userRepo),
	)

	students := api.Group(
		"/students",
		middleware.RequireAuth(jwtSecret),
	)

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
