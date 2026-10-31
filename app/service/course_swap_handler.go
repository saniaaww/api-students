package service

import (
	"strconv"

	"api-students/app/model"

	"github.com/gofiber/fiber/v2"
)

// SwapCourse menangani HTTP request untuk menukar mata kuliah.
func SwapCourse(c *fiber.Ctx, courseService *CourseService) error {
	// Ambil ID pengguna dari middleware RequireAuth.
	userIDValue := c.Locals("user_id")
	if userIDValue == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "unauthorized",
		})
	}

	// Middleware menyimpan user_id sebagai int64.
	userID, ok := userIDValue.(int64)
	if !ok {
		// Antisipasi jika tipe ID yang disimpan berbeda.
		userID, err := strconv.ParseInt(
			strconv.FormatInt(userIDValue.(int64), 10), 10, 64,
		)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "user ID tidak valid",
			})
		}
		_ = userID
	}

	var req model.SwapRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "format request tidak valid",
		})
	}

	if req.FromCourseID == req.ToCourseID {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "mata kuliah asal dan tujuan tidak boleh sama",
		})
	}

	// Konversi ID user dari int64 ke int sesuai signature service.
	_, err := courseService.Swap(
		c.UserContext(),
		int(userID),
		req,
	)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "berhasil menukar mata kuliah",
	})
}
