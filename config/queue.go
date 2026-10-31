
package config

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/hibiken/asynq"

	"api-students/app/model"
)

// RedisOpt membaca konfigurasi Redis dari environment.
func RedisOpt() asynq.RedisClientOpt {
	env := LoadEnv()

	return asynq.RedisClientOpt{
		Addr:         env.RedisAddr,
		Password:     env.RedisPassword,
		DB:           env.RedisDB,
		DialTimeout:  2 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	}
}

// NewWorkerServer menyiapkan server Asynq.
func NewWorkerServer(logger *slog.Logger) *asynq.Server {
	env := LoadEnv()

	return asynq.NewServer(RedisOpt(), asynq.Config{
		Concurrency: env.WorkerConcurrency,

		Queues: map[string]int{
			model.QueueNotifications: 1,
		},

		RetryDelayFunc: func(n int, err error, t *asynq.Task) time.Duration {
			return time.Duration((n+1)*2) * time.Second
		},

		ShutdownTimeout: 10 * time.Second,

		ErrorHandler: asynq.ErrorHandlerFunc(
			func(ctx context.Context, t *asynq.Task, err error) {
				retry, _ := asynq.GetRetryCount(ctx)
				maxRetry, _ := asynq.GetMaxRetry(ctx)

				logger.Warn(
					"task_failed",
					slog.String("type", t.Type()),
					slog.String("attempt", fmt.Sprintf("%d/%d", retry+1, maxRetry+1)),
					slog.String("error", err.Error()),
				)
			},
		),

		Logger:   newAsynqLogger(logger),
		LogLevel: asynq.WarnLevel,
	})
}

// asynqLogger menyalurkan log internal Asynq ke slog.
type asynqLogger struct {
	l *slog.Logger
}

func newAsynqLogger(l *slog.Logger) asynqLogger {
	return asynqLogger{
		l: l.With(slog.String("component", "asynq")),
	}
}

func (a asynqLogger) Debug(args ...any) {
	a.l.Debug(fmt.Sprint(args...))
}

func (a asynqLogger) Info(args ...any) {
	a.l.Info(fmt.Sprint(args...))
}

func (a asynqLogger) Warn(args ...any) {
	a.l.Warn(fmt.Sprint(args...))
}

func (a asynqLogger) Error(args ...any) {
	a.l.Error(fmt.Sprint(args...))
}

func (a asynqLogger) Fatal(args ...any) {
	a.l.Error(fmt.Sprint(args...))
	os.Exit(1)
}
