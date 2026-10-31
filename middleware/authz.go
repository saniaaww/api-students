package middleware

import (
	"strings"

	"api-students/helper"

	"github.com/gofiber/fiber/v2"
)

// RequirePermission memastikan user memiliki permission tertentu.
func RequirePermission(
	perms *helper.PermissionSet,
	permission string,
) fiber.Handler {
	return func(c *fiber.Ctx) error {
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

		if !perms.Can(roleName, permission) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"message": "forbidden",
			})
		}

		return c.Next()
	}
}

// RequireAdmin memastikan hanya user dengan role admin yang dapat mengakses.
func RequireAdmin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("user_id")
		role := c.Locals("role")

		if userID == nil || role == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "unauthorized",
			})
		}

		roleName, ok := role.(string)
		if !ok || !strings.EqualFold(roleName, "admin") {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"message": "admin access required",
			})
		}

		return c.Next()
	}
}
