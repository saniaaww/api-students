package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"api-students/app/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound       = errors.New("data tidak ditemukan")
	ErrDuplicate      = errors.New("NIM sudah digunakan")
	ErrDuplicateEmail = errors.New("email sudah digunakan")
)

type StudentRepository interface {
	FindAll(
		ctx context.Context,
		search string,
		prodi string,
		angkatan *int,
		sort string,
		limit int,
		offset int,
	) ([]model.Student, int, error)

	FindByID(
		ctx context.Context,
		id int,
	) (model.Student, error)

	FindByUserID(
		ctx context.Context,
		userID int,
	) (model.Student, error)

	CreateWithUser(
		ctx context.Context,
		user model.User,
		student model.Student,
	) (model.Student, error)

	Update(
		ctx context.Context,
		id int,
		student model.Student,
	) (model.Student, error)

	Delete(
		ctx context.Context,
		id int,
	) error
}

type studentRepository struct {
	db *pgxpool.Pool
}

func NewStudentRepository(db *pgxpool.Pool) StudentRepository {
	return &studentRepository{
		db: db,
	}
}

func (r *studentRepository) FindAll(
	ctx context.Context,
	search string,
	prodi string,
	angkatan *int,
	sort string,
	limit int,
	offset int,
) ([]model.Student, int, error) {

	where := []string{
		"deleted_at IS NULL",
	}

	args := make([]interface{}, 0)
	argNumber := 1

	// Search NIM atau nama.
	if search != "" {
		where = append(
			where,
			fmt.Sprintf(
				"(nim ILIKE $%d OR nama ILIKE $%d)",
				argNumber,
				argNumber,
			),
		)

		args = append(args, "%"+search+"%")
		argNumber++
	}

	// Filter prodi.
	if prodi != "" {
		where = append(
			where,
			fmt.Sprintf("prodi ILIKE $%d", argNumber),
		)

		args = append(args, prodi)
		argNumber++
	}

	// Filter angkatan.
	if angkatan != nil {
		where = append(
			where,
			fmt.Sprintf("angkatan = $%d", argNumber),
		)

		args = append(args, *angkatan)
		argNumber++
	}

	whereClause := "WHERE " + strings.Join(where, " AND ")

	// Sort hanya boleh berasal dari pilihan yang
	// ditentukan aplikasi, bukan input SQL bebas.
	orderBy := "id ASC"

	switch sort {
	case "nama":
		orderBy = "nama ASC, id ASC"

	case "-ipk_terakhir":
		orderBy = "ipk_terakhir DESC NULLS LAST, id ASC"
	}

	dataQuery := fmt.Sprintf(
		`
		SELECT
			id,
			user_id,
			nim,
			nama,
			prodi,
			angkatan,
			ipk_terakhir,
			deleted_at
		FROM students
		%s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
		`,
		whereClause,
		orderBy,
		argNumber,
		argNumber+1,
	)

	dataArgs := append(
		append([]interface{}{}, args...),
		limit,
		offset,
	)

	rows, err := r.db.Query(
		ctx,
		dataQuery,
		dataArgs...,
	)

	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()

	students := make([]model.Student, 0)

	for rows.Next() {
		var student model.Student

		if err := rows.Scan(
			&student.ID,
			&student.UserID,
			&student.NIM,
			&student.Nama,
			&student.Prodi,
			&student.Angkatan,
			&student.IPKTerakhir,
			&student.DeletedAt,
		); err != nil {
			return nil, 0, err
		}

		students = append(students, student)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	// Hitung total data menggunakan filter yang sama.
	countQuery := fmt.Sprintf(
		`
		SELECT COUNT(*)
		FROM students
		%s
		`,
		whereClause,
	)

	var total int

	if err := r.db.QueryRow(
		ctx,
		countQuery,
		args...,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	return students, total, nil
}

func (r *studentRepository) FindByID(
	ctx context.Context,
	id int,
) (model.Student, error) {

	var student model.Student

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			id,
			user_id,
			nim,
			nama,
			prodi,
			angkatan,
			ipk_terakhir,
			deleted_at
		FROM students
		WHERE id = $1
		  AND deleted_at IS NULL
		`,
		id,
	).Scan(
		&student.ID,
		&student.UserID,
		&student.NIM,
		&student.Nama,
		&student.Prodi,
		&student.Angkatan,
		&student.IPKTerakhir,
		&student.DeletedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return model.Student{}, ErrNotFound
	}

	if err != nil {
		return model.Student{}, err
	}

	return student, nil
}

func (r *studentRepository) FindByUserID(
	ctx context.Context,
	userID int,
) (model.Student, error) {

	var student model.Student

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			id,
			user_id,
			nim,
			nama,
			prodi,
			angkatan,
			ipk_terakhir,
			deleted_at
		FROM students
		WHERE user_id = $1
		  AND deleted_at IS NULL
		`,
		userID,
	).Scan(
		&student.ID,
		&student.UserID,
		&student.NIM,
		&student.Nama,
		&student.Prodi,
		&student.Angkatan,
		&student.IPKTerakhir,
		&student.DeletedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return model.Student{}, ErrNotFound
	}

	if err != nil {
		return model.Student{}, err
	}

	return student, nil
}

func (r *studentRepository) CreateWithUser(
	ctx context.Context,
	user model.User,
	student model.Student,
) (model.Student, error) {

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return model.Student{}, err
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()


	var userID int64

	err = tx.QueryRow(
		ctx,
		`
		INSERT INTO users (
			email,
			password,
			role
		)
		VALUES ($1, $2, $3)
		RETURNING id
		`,
		user.Email,
		user.Password,
		user.Role,
	).Scan(&userID)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) &&
			pgErr.Code == "23505" {

			return model.Student{}, ErrDuplicateEmail
		}

		return model.Student{}, err
	}

	// -------------------------------------------------
	// 2. User ID menjadi students.user_id
	// -------------------------------------------------

	student.UserID = int(userID)

	// -------------------------------------------------
	// 3. Buat STUDENT
	// -------------------------------------------------

	var result model.Student

	err = tx.QueryRow(
		ctx,
		`
		INSERT INTO students (
			user_id,
			nim,
			nama,
			prodi,
			angkatan,
			ipk_terakhir
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING
			id,
			user_id,
			nim,
			nama,
			prodi,
			angkatan,
			ipk_terakhir,
			deleted_at
		`,
		student.UserID,
		student.NIM,
		student.Nama,
		student.Prodi,
		student.Angkatan,
		student.IPKTerakhir,
	).Scan(
		&result.ID,
		&result.UserID,
		&result.NIM,
		&result.Nama,
		&result.Prodi,
		&result.Angkatan,
		&result.IPKTerakhir,
		&result.DeletedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) &&
			pgErr.Code == "23505" {

			return model.Student{}, ErrDuplicate
		}

		return model.Student{}, err
	}

	// -------------------------------------------------
	// 4. COMMIT
	// -------------------------------------------------

	if err := tx.Commit(ctx); err != nil {
		return model.Student{}, err
	}

	return result, nil
}

func (r *studentRepository) Update(
	ctx context.Context,
	id int,
	student model.Student,
) (model.Student, error) {

	var result model.Student

	err := r.db.QueryRow(
		ctx,
		`
		UPDATE students
		SET
			nama = $1,
			prodi = $2,
			angkatan = $3,
			ipk_terakhir = $4
		WHERE id = $5
		  AND deleted_at IS NULL
		RETURNING
			id,
			user_id,
			nim,
			nama,
			prodi,
			angkatan,
			ipk_terakhir,
			deleted_at
		`,
		student.Nama,
		student.Prodi,
		student.Angkatan,
		student.IPKTerakhir,
		id,
	).Scan(
		&result.ID,
		&result.UserID,
		&result.NIM,
		&result.Nama,
		&result.Prodi,
		&result.Angkatan,
		&result.IPKTerakhir,
		&result.DeletedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return model.Student{}, ErrNotFound
	}

	if err != nil {
		return model.Student{}, err
	}

	return result, nil
}

func (r *studentRepository) Delete(
	ctx context.Context,
	id int,
) error {

	result, err := r.db.Exec(
		ctx,
		`
		UPDATE students
		SET deleted_at = CURRENT_TIMESTAMP
		WHERE id = $1
		  AND deleted_at IS NULL
		`,
		id,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}