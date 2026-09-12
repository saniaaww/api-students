package repository

import (
	"context"

	"api-students/app/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByID(ctx context.Context, id int64) (*model.User, error)
	Create(ctx context.Context, user *model.User) (*model.User, error)
}

type userRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userRepository{pool: pool}
}

func (r *userRepository) FindByUsername(
	ctx context.Context,
	username string,
) (*model.User, error) {
	user := &model.User{}

	err := r.pool.QueryRow(
		ctx,
		`SELECT id, username, email, password, role, created_at
		 FROM users
		 WHERE username = $1`,
		username,
	).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return user, nil
}

func (r *userRepository) FindByEmail(
	ctx context.Context,
	email string,
) (*model.User, error) {
	user := &model.User{}

	err := r.pool.QueryRow(
		ctx,
		`SELECT id, username, email, password, role, created_at
		 FROM users
		 WHERE email = $1`,
		email,
	).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return user, nil
}

func (r *userRepository) FindByID(
	ctx context.Context,
	id int64,
) (*model.User, error) {
	user := &model.User{}

	err := r.pool.QueryRow(
		ctx,
		`SELECT id, username, email, password, role, created_at
		 FROM users
		 WHERE id = $1`,
		id,
	).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return user, nil
}

func (r *userRepository) Create(
	ctx context.Context,
	user *model.User,
) (*model.User, error) {
	result := &model.User{}

	err := r.pool.QueryRow(
		ctx,
		`INSERT INTO users (username, email, password, role)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, username, email, password, role, created_at`,
		user.Username,
		user.Email,
		user.Password,
		user.Role,
	).Scan(
		&result.ID,
		&result.Username,
		&result.Email,
		&result.Password,
		&result.Role,
		&result.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return result, nil
}
