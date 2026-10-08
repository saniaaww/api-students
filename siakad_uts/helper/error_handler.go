package helper

import (
	"errors"

	"github.com/gofiber/fiber/v2"
)

func ErrorHandler(c *fiber.Ctx, err error) error {

	// AppError dari aplikasi
	var appErr *AppError

	if errors.As(err, &appErr) {

		response := fiber.Map{
			"success": false,
			"message": appErr.Message,
		}

		// Khusus validasi
		if appErr.Fields != nil {
			errorsMap := make(map[string][]string)

			for field, message := range appErr.Fields {
				errorsMap[field] = []string{message}
			}

			response["errors"] = errorsMap
		}

		return c.Status(appErr.Status).JSON(response)
	}

	// Error Fiber
	var fiberErr *fiber.Error

	if errors.As(err, &fiberErr) {
		return c.Status(fiberErr.Code).JSON(fiber.Map{
			"success": false,
			"message": fiberErr.Message,
		})
	}

	// Error yang tidak dikenal
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"success": false,
		"message": "terjadi kesalahan pada server",
	})
}
