package config
import (
	"io"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"gopkg.in/natefinch/lumberjack.v2"
)
func NewLogger() fiber.Handler {
	// Membuat folder logs jika belum ada.
	if err := os.MkdirAll("logs", 0755); err != nil {
		panic(err)
	}
	// File log dengan fitur rotation.
	fileLogger := &lumberjack.Logger{
		Filename:   "logs/app.log",
		MaxSize:    10,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   true,
	}
	// Log ditulis ke terminal dan file.
	output := io.MultiWriter(
		os.Stdout,
		fileLogger,
	)
	return logger.New(logger.Config{
		Output:        output,
		DisableColors: true,
		Format:        "{\"request_id\":\"${locals:requestid}\",\"method\":\"${method}\",\"path\":\"${path}\",\"status\":${status},\"duration\":\"${latency}\"}\n",
	})
}
