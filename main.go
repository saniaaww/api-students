
package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"api-students/app/repository"
	"api-students/app/service"
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
	// ==================================================
	// 1. MEMUAT ENVIRONMENT
	// ==================================================
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: file .env tidak ditemukan")
	}

	env := config.LoadEnv()

	// Logger untuk komponen background Modul 11.
	workerLogger := slog.New(
		slog.NewJSONHandler(os.Stdout, nil),
	)

	// Context untuk graceful shutdown.
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	// ==================================================
	// 2. KONEKSI POSTGRESQL
	// ==================================================
	pool, err := database.ConnectDB()
	if err != nil {
		log.Fatal("Gagal terhubung ke PostgreSQL: ", err)
	}
	defer pool.Close()

	log.Println("Berhasil terhubung ke PostgreSQL")

	// ==================================================
	// 3. KONEKSI MONGODB
	// ==================================================
	if env.MongoURI == "" {
		log.Fatal("MONGO_URI belum diatur di file .env")
	}

	if env.MongoDatabase == "" {
		log.Fatal("MONGO_DATABASE belum diatur di file .env")
	}

	mongoCtx, cancelMongo := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)

	mongoClient, err := database.NewMongoClient(
		mongoCtx,
		env.MongoURI,
	)
	cancelMongo()

	if err != nil {
		log.Fatal("Gagal terhubung ke MongoDB: ", err)
	}

	defer func() {
		ctx, cancel := context.WithTimeout(
			context.Background(),
			10*time.Second,
		)
		defer cancel()

		if err := mongoClient.Disconnect(ctx); err != nil {
			log.Println("Gagal menutup koneksi MongoDB:", err)
		}
	}()

	mongoDB := mongoClient.Database(env.MongoDatabase)

	log.Printf(
		"Berhasil terhubung ke MongoDB, database: %s",
		mongoDB.Name(),
	)

	// ==================================================
	// 4. MEMBUAT REPOSITORY
	// ==================================================

	// Repository PostgreSQL.
	studentRepo := repository.NewStudentRepository(pool)
	userRepo := repository.NewUserRepository(pool)
	refreshRepo := repository.NewRefreshTokenRepository(pool)
	courseRepo := repository.NewCourseRepository(pool)
	supervisorRepo := repository.NewSupervisorRepository(pool)
	roleRepo := repository.NewRoleRepository(pool)

	// Repository MongoDB.
	portfolioRepo := repository.NewPortfolioRepository(mongoDB)

	// Repository Modul 11: Outbox.
	outboxRepo := repository.NewOutboxRepository(pool)

	// ==================================================
	// 5. TRANSACTION MANAGER
	// ==================================================
	txManager := repository.NewTxManager(pool)

	// ==================================================
	// 6. MEMBUAT SERVICE
	// ==================================================

	// CourseService sekarang menerima tiga parameter:
	// CourseRepository, OutboxRepository, dan TxManager.
	courseService := service.NewCourseService(
		courseRepo,
		outboxRepo,
		txManager,
	)

	supervisorService := service.NewSupervisorService(
		supervisorRepo,
		txManager,
	)

	// ==================================================
	// 7. MEMUAT PERMISSION
	// ==================================================
	rolePermissions, err := roleRepo.LoadPermissions(
		context.Background(),
	)
	if err != nil {
		log.Fatal("Gagal memuat permissions: ", err)
	}

	permissionSet := helper.NewPermissionSet(rolePermissions)

	// ==================================================
	// 8. MEMBUAT APLIKASI FIBER
	// ==================================================
	app := fiber.New(fiber.Config{
		AppName: "API Students - Praktikum Backend Lanjut",
	})

	// Middleware global.
	app.Use(requestid.New())
	app.Use(config.NewLogger())
	app.Use(cors.New())

	// ==================================================
	// 9. REGISTRASI ROUTE
	// ==================================================
	route.Register(
		app,
		pool,
		studentRepo,
		userRepo,
		env.JWTSecret,
		refreshRepo,
		permissionSet,
		courseService,
		supervisorService,
		portfolioRepo,
	)

	// Endpoint tidak ditemukan.
	app.Use(func(c *fiber.Ctx) error {
		return helper.NotFound("endpoint tidak ditemukan")
	})

	// ==================================================
	// 10. MENJALANKAN BACKGROUND WORKER MODUL 11
	// ==================================================
	stopWorker := startWorker(
		ctx,
		workerLogger,
		pool,
	)

	if stopWorker == nil {
		log.Fatal("Gagal menjalankan background worker")
	}

	// ==================================================
	// 11. MENJALANKAN HTTP SERVER
	// ==================================================
	serverErr := make(chan error, 1)

	go func() {
		log.Println("Server berjalan di http://localhost:3000")
		serverErr <- app.Listen(":3000")
	}()

	// Menunggu Ctrl+C atau server berhenti.
	select {
	case <-ctx.Done():
		log.Println("Menerima sinyal shutdown...")

	case err := <-serverErr:
		if err != nil {
			log.Println("HTTP server berhenti:", err)
		}
		stop()
	}

	// ==================================================
	// 12. GRACEFUL SHUTDOWN
	// ==================================================

	// Hentikan HTTP server terlebih dahulu.
	shutdownCtx, cancelShutdown := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancelShutdown()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		log.Println("Gagal melakukan shutdown HTTP:", err)
	}

	// Setelah HTTP berhenti, hentikan relay dan worker.
	stopWorker()

	log.Println("Aplikasi berhenti dengan rapi")
}
