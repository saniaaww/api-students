package config

import (
	"errors"
	"io"
	"log/slog"
	"os"
	"time"

	"api-students/helper"

	"github.com/gofiber/fiber/v2"
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

	logger := slog.New(
		slog.NewJSONHandler(output, nil),
	)

	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		requestID, _ := c.Locals("requestid").(string)

		// Status response.
		status := c.Response().StatusCode()

		// Jika handler mengembalikan error,
		// ambil status dari AppError.
		if err != nil {
			var appErr *helper.AppError

			if errors.As(err, &appErr) {
				status = appErr.Status
			} else {
				status = fiber.StatusInternalServerError
			}
		}

		attrs := []any{
			slog.String("request_id", requestID),
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.Int("status", status),
			slog.Duration("duration", time.Since(start)),
			slog.String("ip", c.IP()),
		}

		// 4xx dianggap request yang ditolak.
		if status >= 400 && status < 500 {
			logger.Warn("http_request", attrs...)
		} else if status >= 500 {
			logger.Error("http_request", attrs...)
		} else {
			logger.Info("http_request", attrs...)
		}

		return err
	}
}
