package service

import (
	"context"
	"errors"
	"strings"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"

	"github.com/gofiber/fiber/v2"
)

func LoginHandler(
	repo repository.UserRepository,
	refreshRepo repository.RefreshTokenRepository,
	jwtSecret string,
) fiber.Handler {

	return func(c *fiber.Ctx) error {

		var req model.LoginRequest

		// ==============================
		// PARSE REQUEST
		// ==============================

		if err := c.BodyParser(&req); err != nil {
			return helper.Validation(map[string]string{
				"body": "format request tidak valid",
			})
		}

		// ==============================
		// BERSIHKAN EMAIL
		// ==============================

		req.Email = strings.TrimSpace(
			strings.ToLower(req.Email),
		)

		// ==============================
		// VALIDASI
		// ==============================

		validationErrors := helper.ValidateStruct(req)

		if validationErrors != nil {
			return helper.Validation(validationErrors)
		}

		// ==============================
		// LOGIN
		// ==============================

		accessToken, expiresIn, user, err := LoginUser(
			context.Background(),
			repo,
			refreshRepo,
			req.Email,
			req.Password,
			jwtSecret,
		)

		if err != nil {

			if errors.Is(err, ErrInvalidLogin) {
				return helper.Unauthorized(
					"email atau password salah",
				)
			}

			return helper.Internal(err)
		}

		// ==============================
		// RESPONSE
		// ==============================

		return helper.SuccessResponse(
			c,
			fiber.StatusOK,
			"login berhasil",
			model.LoginResponse{
				AccessToken: accessToken,
				TokenType:   "Bearer",
				ExpiresIn:   expiresIn,
				User: model.UserSummary{
					ID:    user.ID,
					Email: user.Email,
					Role:  user.Role,
				},
			},
			nil,
		)
	}
}