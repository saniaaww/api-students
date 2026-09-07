package service

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Home(c *fiber.Ctx) error {
	return c.SendString("API Students berjalan")
}

func Health(c *fiber.Ctx, pool *pgxpool.Pool) error {
	if err := pool.Ping(c.Context()); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "database tidak terhubung",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "API dan database berjalan",
	})
}
