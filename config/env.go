
package config

import (
	"os"
	"strconv"
)

type Env struct {
	DBHost        string
	DBPort        string
	DBUser        string
	DBPassword    string
	DBName        string
	JWTSecret     string
	MongoURI      string
	MongoDatabase string

	// Modul 11: Redis dan Background Worker
	RedisAddr       string
	RedisPassword   string
	RedisDB         int
	WorkerConcurrency int
	RelayIntervalMS   int
	RelayBatch        int
	MailerDelayMS     int
	MailerFailEvery   int
}

func LoadEnv() Env {
	return Env{
		DBHost:        os.Getenv("DB_HOST"),
		DBPort:        os.Getenv("DB_PORT"),
		DBUser:        os.Getenv("DB_USER"),
		DBPassword:    os.Getenv("DB_PASSWORD"),
		DBName:        os.Getenv("DB_NAME"),
		JWTSecret:     os.Getenv("JWT_SECRET"),
		MongoURI:      os.Getenv("MONGO_URI"),
		MongoDatabase: os.Getenv("MONGO_DATABASE"),

		RedisAddr:        getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:    getEnv("REDIS_PASSWORD", ""),
		RedisDB:          GetEnvInt("REDIS_DB", 0),
		WorkerConcurrency: GetEnvInt("WORKER_CONCURRENCY", 5),
		RelayIntervalMS:   GetEnvInt("RELAY_INTERVAL_MS", 500),
		RelayBatch:        GetEnvInt("RELAY_BATCH", 50),
		MailerDelayMS:     GetEnvInt("MAILER_DELAY_MS", 0),
		MailerFailEvery:   GetEnvInt("MAILER_FAIL_EVERY", 0),
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func GetEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	result, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return result
}
