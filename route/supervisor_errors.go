package route

import (
	"errors"

	"api-students/app/repository"

	"github.com/gofiber/fiber/v2"
)

func handleSupervisorError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "data mahasiswa atau pembimbing tidak ditemukan",
		})

	case errors.Is(err, repository.ErrSupervisorFull):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})

	case errors.Is(err, repository.ErrAlreadyAssigned):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})

	case errors.Is(err, repository.ErrSupervisorVersionConflict):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})

	case errors.Is(err, repository.ErrQuotaBelowSupervised):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})

	case errors.Is(err, repository.ErrDuplicate):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"success": false,
			"message": "data penetapan pembimbing sudah ada",
		})

	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "terjadi kesalahan internal",
		})
	}
}
