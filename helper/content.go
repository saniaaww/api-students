package helper

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// NegotiateFormat menentukan format response berdasarkan header Accept.
//
// Didukung:
// - application/json
// - text/csv
// - */*
func NegotiateFormat(c *fiber.Ctx) (string, error) {
	accept := strings.TrimSpace(c.Get("Accept"))

	// Accept kosong atau wildcard menggunakan JSON.
	if accept == "" || accept == "*/*" {
		return "json", nil
	}

	for _, mediaType := range strings.Split(accept, ",") {
		mediaType = strings.TrimSpace(
			strings.Split(mediaType, ";")[0],
		)

		switch mediaType {
		case "application/json":
			return "json", nil

		case "text/csv":
			return "csv", nil

		case "*/*":
			return "json", nil
		}
	}

	return "", NotAcceptable(
		"format response tidak didukung",
	)
}