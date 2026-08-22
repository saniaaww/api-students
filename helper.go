package main

import "github.com/gofiber/fiber/v2"

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
}

func successResponse(c *fiber.Ctx, status int, message string, data interface{}, meta interface{}) error {
	return c.Status(status).JSON(APIResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

func errorResponse(c *fiber.Ctx, status int, message string, data interface{}) error {
	return c.Status(status).JSON(APIResponse{
		Success: false,
		Message: message,
		Data:    data,
	})
}
