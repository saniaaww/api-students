package helper

import "github.com/gofiber/fiber/v2"

func RequestID(c *fiber.Ctx) string {
	requestID, _ := c.Locals("requestid").(string)
	return requestID
}
