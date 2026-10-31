package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/hibiken/asynq"

	"api-students/app/model"
	"api-students/config"
	"api-students/database"
)

func main() {
	outboxID := flag.Int64("outbox", 0, "ID event outbox")
	copies := flag.Int("copies", 1, "Jumlah salinan task")
	flag.Parse()

	if *outboxID <= 0 || *copies <= 0 {
		log.Fatal("Gunakan: go run ./tools/redeliver -outbox <id> -copies <jumlah>")
	}

	config.LoadEnv()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Gunakan fungsi koneksi PostgreSQL milik project.
	pool, err := database.ConnectDB()
	if err != nil {
		log.Fatalf("Gagal terhubung ke PostgreSQL: %v", err)
	}
	defer pool.Close()

	var event model.OutboxEvent

	err = pool.QueryRow(ctx, `
		SELECT id, event_type, payload, created_at
		FROM outbox
		WHERE id = $1
	`, *outboxID).Scan(
		&event.ID,
		&event.EventType,
		&event.Payload,
		&event.CreatedAt,
	)
	if err != nil {
		log.Fatalf("Event outbox %d tidak ditemukan: %v", *outboxID, err)
	}

	job := model.NotificationJob{
		OutboxID:  event.ID,
		EventType: event.EventType,
		Payload:   json.RawMessage(event.Payload),
	}

	client := asynq.NewClient(config.RedisOpt())
	defer client.Close()

	task := asynq.NewTask(
		model.TaskSendNotification,
		mustMarshal(job),
		asynq.Queue(model.QueueNotifications),
		asynq.MaxRetry(5),
		asynq.Timeout(30*time.Second),
		asynq.Retention(24*time.Hour),
	)

	for i := 1; i <= *copies; i++ {
		info, err := client.EnqueueContext(ctx, task)
		if err != nil {
			log.Fatalf("Gagal mengirim salinan ke-%d: %v", i, err)
		}

		fmt.Printf(
			"Salinan %d/%d berhasil: task_id=%s\n",
			i, *copies, info.ID,
		)
	}

	fmt.Printf(
		"Selesai: %d salinan event outbox %d dikirim ke Redis.\n",
		*copies, *outboxID,
	)
}

func mustMarshal(v any) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		log.Fatalf("Gagal membuat payload task: %v", err)
	}
	return data
}
