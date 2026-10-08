package middleware

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

type LoginAttempt struct {
	Count       int
	WindowStart time.Time
}

var (
	loginAttempts = make(map[string]*LoginAttempt)
	loginMutex    sync.Mutex
)

const (
	maxLoginAttempts = 5
	loginWindow      = time.Minute
)

func LoginRateLimit() fiber.Handler {
	return func(c *fiber.Ctx) error {

		ip := c.IP()

		loginMutex.Lock()

		attempt, exists := loginAttempts[ip]

		if !exists || time.Since(attempt.WindowStart) >= loginWindow {
			loginAttempts[ip] = &LoginAttempt{
				Count:       0,
				WindowStart: time.Now(),
			}

			attempt = loginAttempts[ip]
		}

		// Jika sudah 5 kali gagal
		if attempt.Count >= maxLoginAttempts {
			loginMutex.Unlock()

			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"message": "terlalu banyak percobaan login, silakan coba lagi nanti",
			})
		}

		loginMutex.Unlock()

		// Jalankan handler login
		err := c.Next()

		// Hanya hitung percobaan yang benar-benar gagal
		// karena kredensial salah.
		if c.Response().StatusCode() == fiber.StatusUnauthorized {

			loginMutex.Lock()

			attempt, exists = loginAttempts[ip]

			if !exists || time.Since(attempt.WindowStart) >= loginWindow {
				loginAttempts[ip] = &LoginAttempt{
					Count:       1,
					WindowStart: time.Now(),
				}
			} else {
				attempt.Count++
			}

			loginMutex.Unlock()
		}

		return err
	}
}
