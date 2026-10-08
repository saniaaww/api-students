package middleware

import (
	"github.com/gofiber/fiber/v2"

	"api-students/helper"
)

// RequirePermission memastikan user memiliki permission tertentu.
func RequirePermission(
	perms *helper.PermissionSet,
	permission string,
) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Pastikan user sudah melewati authentication.
		userID := c.Locals("user_id")
		role := c.Locals("role")

		if userID == nil || role == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "unauthorized",
			})
		}

		roleName, ok := role.(string)
		if !ok || roleName == "" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"message": "forbidden",
			})
		}

		// Cek permission.
		if !perms.Can(roleName, permission) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"message": "forbidden",
			})
		}

		return c.Next()
	}
}
