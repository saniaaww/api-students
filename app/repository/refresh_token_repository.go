package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RefreshTokenRepository interface {
	Create(
		ctx context.Context,
		userID int64,
		tokenHash string,
		expiresAt time.Time,
	) error

	FindActive(
		ctx context.Context,
		tokenHash string,
	) (int64, error)

	Revoke(
		ctx context.Context,
		tokenHash string,
	) error
}

type refreshTokenRepository struct {
	pool *pgxpool.Pool
}

func NewRefreshTokenRepository(pool *pgxpool.Pool) RefreshTokenRepository {
	return &refreshTokenRepository{
		pool: pool,
	}
}

func (r *refreshTokenRepository) Create(
	ctx context.Context,
	userID int64,
	tokenHash string,
	expiresAt time.Time,
) error {
	_, err := r.pool.Exec(
		ctx,
		`INSERT INTO refresh_tokens
			(user_id, token_hash, expires_at)
		 VALUES ($1, $2, $3)`,
		userID,
		tokenHash,
		expiresAt,
	)

	return err
}

func (r *refreshTokenRepository) FindActive(
	ctx context.Context,
	tokenHash string,
) (int64, error) {
	var userID int64

	err := r.pool.QueryRow(
		ctx,
		`SELECT user_id
		 FROM refresh_tokens
		 WHERE token_hash = $1
		   AND revoked_at IS NULL
		   AND expires_at > CURRENT_TIMESTAMP`,
		tokenHash,
	).Scan(&userID)

	return userID, err
}

func (r *refreshTokenRepository) Revoke(
	ctx context.Context,
	tokenHash string,
) error {
	_, err := r.pool.Exec(
		ctx,
		`UPDATE refresh_tokens
		 SET revoked_at = CURRENT_TIMESTAMP
		 WHERE token_hash = $1
		   AND revoked_at IS NULL`,
		tokenHash,
	)

	return err
}
