package route

import (
	"time"

	"api-students/app/repository"
	"api-students/app/service"
	"api-students/helper"
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
	permissionSet *helper.PermissionSet,
) {
	// =====================================================
	// Home
	// =====================================================

	app.Get("/", service.Home)

	// =====================================================
	// API
	// =====================================================

	api := app.Group("/api/v1")

	api.Get("/health", func(c *fiber.Ctx) error {
		return service.Health(c, pool)
	})

	// =====================================================
	// Authentication
	// =====================================================

	auth := api.Group("/auth")

	auth.Post(
		"/register",
		service.RegisterHandler(userRepo),
	)

	// Rate limiter untuk login:
	// maksimal 5 percobaan dalam 1 menit.
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

	auth.Post(
		"/refresh",
		service.RefreshHandler(
			userRepo,
			refreshRepo,
			jwtSecret,
		),
	)

	auth.Post(
		"/logout",
		service.LogoutHandler(refreshRepo),
	)

	auth.Get(
		"/me",
		middleware.RequireAuth(jwtSecret),
		service.MeHandler(userRepo),
	)

	// =====================================================
	// Students
	// =====================================================

	// Semua endpoint students wajib login terlebih dahulu.
	students := api.Group(
		"/students",
		middleware.RequireAuth(jwtSecret),
	)

	// -----------------------------------------------------
	// GET /api/v1/students
	// Permission: student:list
	// -----------------------------------------------------

	students.Get(
		"/",
		middleware.RequirePermission(
			permissionSet,
			"student:list",
		),
		func(c *fiber.Ctx) error {
			return service.GetStudents(
				c,
				repo,
			)
		},
	)

	// -----------------------------------------------------
	// GET /api/v1/students/:id
	// Ownership:
	// - Owner boleh melihat data sendiri.
	// - Non-owner membutuhkan student:read:any.
	// -----------------------------------------------------

	students.Get(
		"/cursor",
		middleware.RequirePermission(
			permissionSet,
			"student:list",
		),
		func(c *fiber.Ctx) error {
			return service.GetStudentsCursor(
				c,
				repo,
			)
		},
	)

	students.Get(
		"/:id",
		func(c *fiber.Ctx) error {
			return service.GetStudent(
				c,
				repo,
				permissionSet,
			)
		},
	)

	// -----------------------------------------------------
	// POST /api/v1/students
	// Permission: student:create
	// OwnerID diambil otomatis dari user yang login.
	// -----------------------------------------------------

	students.Post(
		"/",
		middleware.RequirePermission(
			permissionSet,
			"student:create",
		),
		func(c *fiber.Ctx) error {
			return service.CreateStudent(
				c,
				repo,
			)
		},
	)

	// -----------------------------------------------------
	// PUT /api/v1/students/:id
	// Ownership:
	// - Owner boleh mengubah data sendiri.
	// - Non-owner membutuhkan student:update:any.
	// -----------------------------------------------------

	students.Put(
		"/:id",
		func(c *fiber.Ctx) error {
			return service.ReplaceStudent(
				c,
				repo,
				permissionSet,
			)
		},
	)

	// -----------------------------------------------------
	// PATCH /api/v1/students/:id
	// Ownership:
	// - Owner boleh mengubah data sendiri.
	// - Non-owner membutuhkan student:update:any.
	// -----------------------------------------------------

	students.Patch(
		"/:id",
		func(c *fiber.Ctx) error {
			return service.PatchStudent(
				c,
				repo,
				permissionSet,
			)
		},
	)

	// -----------------------------------------------------
	// DELETE /api/v1/students/:id
	// Permission: student:delete
	// -----------------------------------------------------

	students.Delete(
		"/:id",
		middleware.RequirePermission(
			permissionSet,
			"student:delete",
		),
		func(c *fiber.Ctx) error {
			return service.DeleteStudent(
				c,
				repo,
			)
		},
	)
}
