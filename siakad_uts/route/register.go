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
	courseRepo repository.CourseRepository,
	enrollmentRepo repository.EnrollmentRepository,
	jwtSecret string,
	refreshRepo repository.RefreshTokenRepository,
) {

	// =====================================================
	// HANDLER
	// =====================================================

	handler := &service.Handler{
		Students:   repo,
		Users:      userRepo,
		Courses:    courseRepo,
		Enrollments: enrollmentRepo,
	}

	// =====================================================
	// HOME
	// =====================================================

	app.Get("/", service.Home)

	// =====================================================
	// API
	// =====================================================

	api := app.Group("/api/v1")

	// =====================================================
	// HEALTH
	// =====================================================

	api.Get("/health", func(c *fiber.Ctx) error {
		return service.Health(c, pool)
	})

	// =====================================================
	// AUTH
	// =====================================================

	auth := api.Group("/auth")

	// =====================================================
	// LOGIN
	// =====================================================

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

	// =====================================================
	// STUDENTS
	// =====================================================

	students := api.Group(
		"/students",
		middleware.RequireAuth(jwtSecret),
	)

	// GET /api/v1/students
	students.Get(
		"/",
		handler.GetStudents,
	)

	// GET /api/v1/students/:id
	students.Get(
		"/:id",
		handler.GetStudentByID,
	)

	// POST /api/v1/students
	students.Post(
		"/",
		handler.CreateStudent,
	)

	// PUT /api/v1/students/:id
	students.Put(
		"/:id",
		handler.UpdateStudent,
	)

	// DELETE /api/v1/students/:id
	students.Delete(
		"/:id",
		handler.DeleteStudent,
	)

	// =====================================================
	// COURSES
	// =====================================================

	courses := api.Group(
		"/courses",
		middleware.RequireAuth(jwtSecret),
	)

	courses.Get(
		"/",
		handler.GetCourses,
	)

	// =====================================================
	// ENROLLMENTS
	// =====================================================

	enrollments := api.Group(
		"/enrollments",
		middleware.RequireAuth(jwtSecret),
	)

	enrollments.Post(
		"/",
		handler.CreateEnrollment,
	)

	enrollments.Delete(
		"/:id",
		handler.DeleteEnrollment,
	)
}