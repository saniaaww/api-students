package repository

import (
	"context"
	"errors"

	"api-students/app/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrUserNotFound = errors.New("user tidak ditemukan")
var ErrUserDuplicateEmail = errors.New("email sudah digunakan")

type UserRepository interface {
	FindByEmail(
		ctx context.Context,
		email string,
	) (*model.User, error)

	FindByID(
		ctx context.Context,
		id int64,
	) (*model.User, error)

	Create(
		ctx context.Context,
		user *model.User,
	) (*model.User, error)
}

type userRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(
	pool *pgxpool.Pool,
) UserRepository {
	return &userRepository{
		pool: pool,
	}
}

// =====================================================
// FIND BY EMAIL
// =====================================================

func (r *userRepository) FindByEmail(
	ctx context.Context,
	email string,
) (*model.User, error) {

	user := &model.User{}

	err := r.pool.QueryRow(
		ctx,
		`
		SELECT
			id,
			email,
			password,
			role,
			created_at
		FROM users
		WHERE email = $1
		`,
		email,
	).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}

	if err != nil {
		return nil, err
	}

	return user, nil
}

// =====================================================
// FIND BY ID
// =====================================================

func (r *userRepository) FindByID(
	ctx context.Context,
	id int64,
) (*model.User, error) {

	user := &model.User{}

	err := r.pool.QueryRow(
		ctx,
		`
		SELECT
			id,
			email,
			password,
			role,
			created_at
		FROM users
		WHERE id = $1
		`,
		id,
	).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}

	if err != nil {
		return nil, err
	}

	return user, nil
}

// =====================================================
// CREATE
// =====================================================

func (r *userRepository) Create(
	ctx context.Context,
	user *model.User,
) (*model.User, error) {

	result := &model.User{}

	err := r.pool.QueryRow(
		ctx,
		`
		INSERT INTO users (
			email,
			password,
			role
		)
		VALUES ($1, $2, $3)
		RETURNING
			id,
			email,
			password,
			role,
			created_at
		`,
		user.Email,
		user.Password,
		user.Role,
	).Scan(
		&result.ID,
		&result.Email,
		&result.Password,
		&result.Role,
		&result.CreatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) &&
			pgErr.Code == "23505" {

			return nil, ErrUserDuplicateEmail
		}

		return nil, err
	}

	return result, nil
}