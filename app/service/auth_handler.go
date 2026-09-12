package service

import (
	"errors"

	"api-students/app/model"
	"api-students/app/repository"

	"github.com/gofiber/fiber/v2"
)

func RegisterHandler(repo repository.UserRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req model.RegisterRequest

		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "format request tidak valid",
			})
		}

		user, err := RegisterUser(c.Context(), repo, req)
		if err != nil {
			status := fiber.StatusBadRequest

			switch {
			case errors.Is(err, ErrUsernameExists):
				status = fiber.StatusConflict
			case errors.Is(err, ErrEmailExists):
				status = fiber.StatusConflict
			}

			return c.Status(status).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"success": true,
			"message": "registrasi berhasil",
			"data": fiber.Map{
				"id":       user.ID,
				"username": user.Username,
				"email":    user.Email,
				"role":     user.Role,
			},
		})
	}
}

func LoginHandler(
	repo repository.UserRepository,
	refreshRepo repository.RefreshTokenRepository,
	jwtSecret string,
) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "format request tidak valid",
			})
		}

		accessToken, refreshToken, err := LoginUser(
			c.Context(),
			repo,
			refreshRepo,
			req.Username,
			req.Password,
			jwtSecret,
		)

		if err != nil {
			if errors.Is(err, ErrInvalidLogin) {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"success": false,
					"message": "username atau password salah",
				})
			}

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "terjadi kesalahan pada server",
			})
		}

		return c.JSON(fiber.Map{
			"success": true,
			"message": "login berhasil",
			"data": fiber.Map{
				"access_token":  accessToken,
				"refresh_token": refreshToken,
				"token_type":    "Bearer",
			},
		})
	}
}

func RefreshHandler(
	userRepo repository.UserRepository,
	refreshRepo repository.RefreshTokenRepository,
	jwtSecret string,
) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req struct {
			RefreshToken string `json:"refresh_token"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "format request tidak valid",
			})
		}

		if req.RefreshToken == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "refresh token wajib diisi",
			})
		}

		accessToken, newRefreshToken, err := RefreshAccessToken(
			c.Context(),
			userRepo,
			refreshRepo,
			req.RefreshToken,
			jwtSecret,
		)

		if err != nil {
			if errors.Is(err, ErrInvalidRefreshToken) {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"success": false,
					"message": "refresh token tidak valid atau sudah tidak aktif",
				})
			}

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "terjadi kesalahan pada server",
			})
		}

		return c.JSON(fiber.Map{
			"success": true,
			"message": "token berhasil diperbarui",
			"data": fiber.Map{
				"access_token":  accessToken,
				"refresh_token": newRefreshToken,
				"token_type":    "Bearer",
			},
		})
	}
}

func LogoutHandler(refreshRepo repository.RefreshTokenRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req struct {
			RefreshToken string `json:"refresh_token"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "format request tidak valid",
			})
		}

		if req.RefreshToken == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "refresh token wajib diisi",
			})
		}

		tokenHash := HashRefreshToken(req.RefreshToken)

		if err := refreshRepo.Revoke(c.Context(), tokenHash); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "gagal melakukan logout",
			})
		}

		return c.JSON(fiber.Map{
			"success": true,
			"message": "logout berhasil",
		})
	}
}

func MeHandler(repo repository.UserRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(int64)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "identitas user tidak ditemukan",
			})
		}

		user, err := repo.FindByID(c.Context(), userID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "gagal mengambil data user",
			})
		}

		if user == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "user tidak ditemukan",
			})
		}

		return c.JSON(fiber.Map{
			"success": true,
			"data": fiber.Map{
				"id":       user.ID,
				"username": user.Username,
				"email":    user.Email,
				"role":     user.Role,
			},
		})
	}
}
