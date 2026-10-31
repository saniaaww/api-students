
package repository

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"api-students/database"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func TestRollbackBeforeAfter(t *testing.T) {
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir("../../"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Chdir(originalDir)
	}()

	if err := godotenv.Load(".env"); err != nil {
		t.Fatalf("Gagal memuat .env: %v", err)
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	pool, err := database.ConnectDB()
	if err != nil {
		t.Fatalf("Gagal menghubungkan database: %v", err)
	}
	defer pool.Close()

	monitorPool, err := database.ConnectDB()
	if err != nil {
		t.Fatalf("Gagal membuat pool pemeriksaan: %v", err)
	}
	defer monitorPool.Close()

	manager := NewTxManager(pool)
	callbackErr := errors.New("simulasi callback gagal")

	var backendPID int

	err = manager.runOnce(ctx, func(tx DBTX) error {
		if err := tx.QueryRow(
			ctx,
			"SELECT pg_backend_pid()",
		).Scan(&backendPID); err != nil {
			return err
		}

		if _, err := tx.Exec(ctx, "SELECT 1"); err != nil {
			return err
		}

		return callbackErr
	})

	if !errors.Is(err, callbackErr) {
		t.Fatalf("Error tidak sesuai, didapat: %v", err)
	}

	// Periksa status dari pool yang berbeda.
	var state string
	queryErr := monitorPool.QueryRow(
		ctx,
		`SELECT state
		 FROM pg_stat_activity
		 WHERE pid = $1`,
		backendPID,
	).Scan(&state)

	if errors.Is(queryErr, pgx.ErrNoRows) {
		t.Log("PASS: backend transaksi sudah tidak ditemukan")
		return
	}

	if queryErr != nil {
		t.Fatalf("Gagal memeriksa status transaksi: %v", queryErr)
	}

	if state == "idle in transaction" {
		t.Errorf(
			"BUG TERDETEKSI: transaksi masih terbuka (PID=%d, state=%s)",
			backendPID,
			state,
		)
		return
	}

	t.Logf(
		"PASS: transaksi tidak lagi idle in transaction (PID=%d, state=%s)",
		backendPID,
		state,
	)
}
