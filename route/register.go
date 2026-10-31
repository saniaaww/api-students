package route

import (
	"time"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/app/service"
	"api-students/config"
	"api-students/helper"
	"api-students/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/hibiken/asynq"
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
	courseService *service.CourseService,
	supervisorService *service.SupervisorService,
	portfolioRepo repository.PortfolioRepository,
) {
	// =====================================================
	// Home
	// =====================================================
	app.Get("/", service.Home)

	// =====================================================
	// API
	// =====================================================
	api := app.Group("/api/v1")

	// =====================================================
	// Admin - Queue Monitoring (Modul 11)
	// GET /admin/queue
	// =====================================================
	app.Get(
		"/admin/queue",
		middleware.RequireAuth(jwtSecret),
		middleware.RequireAdmin(),
		func(c *fiber.Ctx) error {
			// Mengambil statistik Outbox dari PostgreSQL.
			outboxRepo := repository.NewOutboxRepository(pool)

			pending, published, oldest, err := outboxRepo.Stats(
				c.UserContext(),
			)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(
					fiber.Map{
						"success": false,
						"message": "gagal mengambil statistik outbox",
						"error":   err.Error(),
					},
				)
			}

			// Memeriksa koneksi Redis melalui Asynq Inspector.
			inspector := asynq.NewInspector(config.RedisOpt())
			defer inspector.Close()

			queues, queueErr := inspector.Queues()

			response := fiber.Map{
				"success":          true,
				"outbox_pending":   pending,
				"outbox_published": published,
				"outbox_oldest":    oldest.String(),
				"queue_error":      nil,
				"queues":           queues,
			}

			if queueErr != nil {
				response["queue_error"] = queueErr.Error()
				response["queues"] = nil
			}

			return c.JSON(response)
		},
	)

	// =====================================================
	// Courses
	// =====================================================
	api.Get("/courses", func(c *fiber.Ctx) error {
		return service.GetCourses(c, courseService)
	})

	// =====================================================
	// Enrollments - Enroll dan Drop Course
	// =====================================================
	courses := api.Group(
		"/courses",
		middleware.RequireAuth(jwtSecret),
	)

	// POST /api/v1/courses/:id/enroll
	courses.Post("/:id/enroll", func(c *fiber.Ctx) error {
		courseID, err := c.ParamsInt("id")
		if err != nil || courseID <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "ID mata kuliah tidak valid",
			})
		}

		userID, ok := c.Locals("user_id").(int64)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "User tidak terautentikasi",
			})
		}

		err = courseService.Enroll(
			c.UserContext(),
			courseID,
			int(userID),
		)
		if err != nil {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"success": true,
			"message": "Berhasil mendaftar mata kuliah",
		})
	})

	// DELETE /api/v1/courses/:id/enroll
	courses.Delete("/:id/enroll", func(c *fiber.Ctx) error {
		courseID, err := c.ParamsInt("id")
		if err != nil || courseID <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "ID mata kuliah tidak valid",
			})
		}

		userID, ok := c.Locals("user_id").(int64)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "User tidak terautentikasi",
			})
		}

		err = courseService.Drop(
			c.UserContext(),
			courseID,
			int(userID),
		)
		if err != nil {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		}

		return c.SendStatus(fiber.StatusNoContent)
	})

	// =====================================================
	// D.2 - Update Quota Course
	// =====================================================
	api.Patch(
		"/courses/:id",
		middleware.RequireAuth(jwtSecret),
		middleware.RequirePermission(permissionSet, "course:manage"),
		func(c *fiber.Ctx) error {
			id, err := c.ParamsInt("id")
			if err != nil || id <= 0 {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"message": "ID kelas tidak valid",
				})
			}

			var req model.UpdateQuotaRequest
			if err := c.BodyParser(&req); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"message": "format JSON tidak valid",
				})
			}

			updated, err := courseService.UpdateQuota(
				c.UserContext(),
				id,
				req.Quota,
				req.Version,
			)
			if err != nil {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{
					"success": false,
					"message": err.Error(),
				})
			}

			return c.JSON(fiber.Map{
				"success": true,
				"message": "kuota kelas berhasil diperbarui",
				"data":    updated,
			})
		},
	)

	// =====================================================
	// Health
	// =====================================================
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
	// Portfolio Service
	// =====================================================
	portfolioService := service.NewPortfolioService(
		repo,
		portfolioRepo,
	)

	// =====================================================
	// Students
	// =====================================================
	students := api.Group(
		"/students",
		middleware.RequireAuth(jwtSecret),
	)

	students.Get(
		"/",
		middleware.RequirePermission(permissionSet, "student:list"),
		func(c *fiber.Ctx) error {
			return service.GetStudents(c, repo)
		},
	)

	students.Get(
		"/cursor",
		middleware.RequirePermission(permissionSet, "student:list"),
		func(c *fiber.Ctx) error {
			return service.GetStudentsCursor(c, repo)
		},
	)

	students.Get(
		"/:id",
		func(c *fiber.Ctx) error {
			return service.GetStudent(c, repo, permissionSet)
		},
	)

	students.Post(
		"/",
		middleware.RequirePermission(permissionSet, "student:create"),
		func(c *fiber.Ctx) error {
			return service.CreateStudent(c, repo)
		},
	)

	students.Put(
		"/:id",
		func(c *fiber.Ctx) error {
			return service.ReplaceStudent(c, repo, permissionSet)
		},
	)

	students.Patch(
		"/:id",
		func(c *fiber.Ctx) error {
			return service.PatchStudent(c, repo, permissionSet)
		},
	)

	students.Delete(
		"/:id",
		middleware.RequirePermission(permissionSet, "student:delete"),
		func(c *fiber.Ctx) error {
			return service.DeleteStudent(c, repo)
		},
	)

	// =====================================================
	// Portfolio MongoDB
	// =====================================================
	students.Get(
		"/:id/portfolio",
		portfolioService.Get,
	)

	students.Put(
		"/:id/portfolio",
		portfolioService.Put,
	)

	students.Post(
		"/:id/portfolio/projects",
		portfolioService.AddProject,
	)

	students.Delete(
		"/:id/portfolio/projects/:projectId",
		portfolioService.DeleteProject,
	)

	// =====================================================
	// Portfolio Statistics
	// =====================================================
	portfolios := api.Group(
		"/portfolios",
		middleware.RequireAuth(jwtSecret),
	)

	portfolios.Get(
		"/stats/technologies",
		portfolioService.TechnologyStats,
	)

	// =====================================================
	// Notifications - Notifikasi Mahasiswa (Modul 11)
	// =====================================================
	notifications := api.Group(
		"/notifications",
		middleware.RequireAuth(jwtSecret),
	)

	// GET /api/v1/notifications/mine
	notifications.Get("/mine", func(c *fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(int64)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "User tidak terautentikasi",
			})
		}

		notificationRepo := repository.NewNotificationRepository(pool)

		data, err := notificationRepo.FindByUser(
			c.UserContext(),
			int(userID),
			50,
		)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(
				fiber.Map{
					"success": false,
					"message": "Gagal mengambil notifikasi",
					"error":   err.Error(),
				},
			)
		}

		return c.JSON(fiber.Map{
			"success": true,
			"message": "Berhasil mengambil notifikasi",
			"count":   len(data),
			"data":    data,
		})
	})

	// =====================================================
	// Enrollments - Swap Course
	// =====================================================
	enrollments := api.Group(
		"/enrollments",
		middleware.RequireAuth(jwtSecret),
	)

	enrollments.Post(
		"/swap",
		func(c *fiber.Ctx) error {
			return service.SwapCourse(c, courseService)
		},
	)

	// =====================================================
	// Supervisors - Pembimbing Tugas Akhir
	// =====================================================
	supervisors := api.Group(
		"/supervisors",
		middleware.RequireAuth(jwtSecret),
		middleware.RequirePermission(permissionSet, "supervision:manage"),
	)

	supervisors.Post("/:id/assign", func(c *fiber.Ctx) error {
		supervisorID, err := c.ParamsInt("id")
		if err != nil || supervisorID <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "ID pembimbing tidak valid",
			})
		}

		var req model.AssignSupervisorRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "format JSON tidak valid",
			})
		}

		if req.StudentID <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "student_id harus berupa angka positif",
			})
		}

		if err := supervisorService.Assign(
			c.UserContext(),
			supervisorID,
			req.StudentID,
		); err != nil {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"success": true,
			"message": "pembimbing berhasil ditetapkan",
		})
	})

	// =====================================================
	// Supervisions - Switch Supervisor
	// =====================================================
	supervisions := api.Group(
		"/supervisions",
		middleware.RequireAuth(jwtSecret),
		middleware.RequirePermission(permissionSet, "supervision:manage"),
	)

	supervisions.Post("/switch", func(c *fiber.Ctx) error {
		var req model.SwitchSupervisorRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "format JSON tidak valid",
			})
		}

		if err := supervisorService.Switch(
			c.UserContext(),
			req,
		); err != nil {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"success": true,
			"message": "pembimbing berhasil diganti",
		})
	})

	// Mengubah kuota dengan optimistic locking.
	supervisors.Patch("/:id", func(c *fiber.Ctx) error {
		supervisorID, err := c.ParamsInt("id")
		if err != nil || supervisorID <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "ID pembimbing tidak valid",
			})
		}

		var req model.UpdateSupervisorQuotaRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "format JSON tidak valid",
			})
		}

		updated, err := supervisorService.UpdateQuota(
			c.UserContext(),
			supervisorID,
			req.Quota,
			req.Version,
		)
		if err != nil {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"success": true,
			"message": "kuota pembimbing berhasil diperbarui",
			"data":    updated,
		})
	})
}
