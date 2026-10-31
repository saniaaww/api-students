package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/hibiken/asynq"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"
)

type NotificationWorker struct {
	users         repository.UserRepository
	notifications repository.NotificationRepository
	tx            *repository.TxManager
	mailer        helper.Mailer
	logger        *slog.Logger
}

func NewNotificationWorker(users repository.UserRepository, notifications repository.NotificationRepository, tx *repository.TxManager, mailer helper.Mailer, logger *slog.Logger) *NotificationWorker {
	return &NotificationWorker{users: users, notifications: notifications, tx: tx, mailer: mailer, logger: logger}
}

func (w *NotificationWorker) HandleSend(ctx context.Context, t *asynq.Task) error {
	var job model.NotificationJob
	if err := json.Unmarshal(t.Payload(), &job); err != nil {
		return fmt.Errorf("%w: payload task rusak: %v", asynq.SkipRetry, err)
	}

	if job.EventType == model.EventCourseQuotaChanged {
		var event model.CourseQuotaChangedEvent
		if err := json.Unmarshal(job.Payload, &event); err != nil {
			return fmt.Errorf("%w: payload quota event rusak: %v", asynq.SkipRetry, err)
		}
		for _, userID := range event.UserIDs {
			if err := w.sendQuotaNotification(ctx, job, event, userID); err != nil {
				return err // retry melanjutkan; penerima yang sudah sukses di-skip oleh UNIQUE key
			}
		}
		w.logger.Info("quota_notifications_processed", slog.Int64("outbox_id", job.OutboxID), slog.Int("recipients", len(event.UserIDs)))
		return nil
	}

	var event model.EnrollmentEvent
	if err := json.Unmarshal(job.Payload, &event); err != nil {
		return fmt.Errorf("%w: payload event %d rusak: %v", asynq.SkipRetry, job.OutboxID, err)
	}
	user, err := w.users.FindByID(ctx, int64(event.UserID))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fmt.Errorf("%w: user %d tidak ditemukan", asynq.SkipRetry, event.UserID)
		}
		return err
	}
	if user == nil {
		return fmt.Errorf("%w: user %d tidak ditemukan", asynq.SkipRetry, event.UserID)
	}
	msg, err := BuildMessage(job.EventType, event, *user)
	if err != nil {
		return fmt.Errorf("%w: %v", asynq.SkipRetry, err)
	}
	return w.sendOne(ctx, job.OutboxID, int(user.ID), msg, job.EventType)
}

func (w *NotificationWorker) sendQuotaNotification(ctx context.Context, job model.NotificationJob, event model.CourseQuotaChangedEvent, userID int) error {
	user, err := w.users.FindByID(ctx, int64(userID))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil
		}
		return err
	}
	if user == nil {
		return nil
	}
	msg := BuildQuotaChangedMessage(event, *user)
	return w.sendOne(ctx, job.OutboxID, int(user.ID), msg, job.EventType)
}

func (w *NotificationWorker) sendOne(ctx context.Context, outboxID int64, userID int, msg model.Message, eventType string) error {
	alreadySent := false
	err := w.tx.WithTxOnce(ctx, func(tx repository.DBTX) error {
		claimed, err := w.notifications.WithTx(tx).Claim(ctx, model.Notification{OutboxID: outboxID, UserID: userID, Channel: "email", Subject: msg.Subject, Body: msg.Body})
		if err != nil {
			return err
		}
		if !claimed {
			alreadySent = true
			return nil
		}
		return w.mailer.Send(ctx, msg)
	})
	if errors.Is(err, helper.ErrPermanentDelivery) {
		return fmt.Errorf("%w: %v", asynq.SkipRetry, err)
	}
	if err != nil {
		return err
	}
	retry, _ := asynq.GetRetryCount(ctx)
	w.logger.Info("notification_processed", slog.Int64("outbox_id", outboxID), slog.String("event", eventType), slog.Int("user_id", userID), slog.Int("retry", retry), slog.Bool("already_sent", alreadySent))
	return nil
}
