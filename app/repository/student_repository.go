package repository

import (
	"context"
	"errors"
	"fmt"

	"api-students/app/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound  = errors.New("student tidak ditemukan")
	ErrDuplicate = errors.New("NIM sudah digunakan")
)

type StudentRepository interface {

	// =====================================================
	// Pagination lama (OFFSET)
	// Tetap dipertahankan agar handler lama tetap berjalan.
	// =====================================================

	FindAll(
		ctx context.Context,
		search string,
		isActive *bool,
		gradeMin *float64,
		gradeMax *float64,
		sort string,
		order string,
		limit int,
		offset int,
	) ([]model.Student, int, error)

	// =====================================================
	// Cursor Pagination
	// =====================================================

	FindAllCursor(
		ctx context.Context,
		search string,
		isActive *bool,
		gradeMin *float64,
		gradeMax *float64,
		limit int,
		cursor int,
	) ([]model.Student, error)

	FindByID(
		ctx context.Context,
		id int,
	) (model.Student, error)

	Create(
		ctx context.Context,
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

// =====================================================
// FIND ALL - PAGINATION LAMA
// =====================================================

func (r *studentRepository) FindAll(
	ctx context.Context,
	search string,
	isActive *bool,
	gradeMin *float64,
	gradeMax *float64,
	sort string,
	order string,
	limit int,
	offset int,
) ([]model.Student, int, error) {

	where := "WHERE 1=1"
	args := []interface{}{}
	argNumber := 1

	if search != "" {
		where += fmt.Sprintf(
			" AND name ILIKE $%d",
			argNumber,
		)

		args = append(
			args,
			"%"+search+"%",
		)

		argNumber++
	}

	if isActive != nil {
		where += fmt.Sprintf(
			" AND is_active = $%d",
			argNumber,
		)

		args = append(
			args,
			*isActive,
		)

		argNumber++
	}

	if gradeMin != nil {
		where += fmt.Sprintf(
			" AND grade >= $%d",
			argNumber,
		)

		args = append(
			args,
			*gradeMin,
		)

		argNumber++
	}

	if gradeMax != nil {
		where += fmt.Sprintf(
			" AND grade <= $%d",
			argNumber,
		)

		args = append(
			args,
			*gradeMax,
		)

		argNumber++
	}

	allowedSort := map[string]string{
		"id":         "id",
		"nim":        "nim",
		"name":       "name",
		"grade":      "grade",
		"is_active":  "is_active",
		"created_at": "created_at",
	}

	sortColumn, ok := allowedSort[sort]

	if !ok {
		sortColumn = "id"
	}

	if order != "desc" {
		order = "asc"
	}

	query := fmt.Sprintf(`
		SELECT
			id,
			nim,
			name,
			grade,
			is_active,
			owner_id,
			created_at
		FROM students
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`,
		where,
		sortColumn,
		order,
		argNumber,
		argNumber+1,
	)

	args = append(
		args,
		limit,
		offset,
	)

	rows, err := r.db.Query(
		ctx,
		query,
		args...,
	)

	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()

	students := []model.Student{}

	for rows.Next() {

		var student model.Student

		err := rows.Scan(
			&student.ID,
			&student.NIM,
			&student.Name,
			&student.Grade,
			&student.IsActive,
			&student.OwnerID,
			&student.CreatedAt,
		)

		if err != nil {
			return nil, 0, err
		}

		students = append(
			students,
			student,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var total int

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM students
		%s
	`, where)

	if err := r.db.QueryRow(
		ctx,
		countQuery,
		args[:len(args)-2]...,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	return students, total, nil
}

// =====================================================
// FIND ALL CURSOR
// =====================================================

func (r *studentRepository) FindAllCursor(
	ctx context.Context,
	search string,
	isActive *bool,
	gradeMin *float64,
	gradeMax *float64,
	limit int,
	cursor int,
) ([]model.Student, error) {

	// Cursor menggunakan ID.
	// Data yang diambil adalah ID setelah cursor.
	where := "WHERE id > $1"

	args := []interface{}{
		cursor,
	}

	argNumber := 2

	// Search berdasarkan nama
	if search != "" {

		where += fmt.Sprintf(
			" AND name ILIKE $%d",
			argNumber,
		)

		args = append(
			args,
			"%"+search+"%",
		)

		argNumber++
	}

	// Filter is_active
	if isActive != nil {

		where += fmt.Sprintf(
			" AND is_active = $%d",
			argNumber,
		)

		args = append(
			args,
			*isActive,
		)

		argNumber++
	}

	// Filter grade minimum
	if gradeMin != nil {

		where += fmt.Sprintf(
			" AND grade >= $%d",
			argNumber,
		)

		args = append(
			args,
			*gradeMin,
		)

		argNumber++
	}

	// Filter grade maksimum
	if gradeMax != nil {

		where += fmt.Sprintf(
			" AND grade <= $%d",
			argNumber,
		)

		args = append(
			args,
			*gradeMax,
		)

		argNumber++
	}

	query := fmt.Sprintf(`
		SELECT
			id,
			nim,
			name,
			grade,
			is_active,
			owner_id,
			created_at
		FROM students
		%s
		ORDER BY id ASC
		LIMIT $%d
	`,
		where,
		argNumber,
	)

	args = append(
		args,
		limit,
	)

	rows, err := r.db.Query(
		ctx,
		query,
		args...,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	students := []model.Student{}

	for rows.Next() {

		var student model.Student

		err := rows.Scan(
			&student.ID,
			&student.NIM,
			&student.Name,
			&student.Grade,
			&student.IsActive,
			&student.OwnerID,
			&student.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		students = append(
			students,
			student,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return students, nil
}

// =====================================================
// FIND BY ID
// =====================================================

func (r *studentRepository) FindByID(
	ctx context.Context,
	id int,
) (model.Student, error) {

	var student model.Student

	query := `
		SELECT
			id,
			nim,
			name,
			grade,
			is_active,
			owner_id,
			created_at
		FROM students
		WHERE id = $1
	`

	err := r.db.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&student.ID,
		&student.NIM,
		&student.Name,
		&student.Grade,
		&student.IsActive,
		&student.OwnerID,
		&student.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return model.Student{}, ErrNotFound
	}

	if err != nil {
		return model.Student{}, err
	}

	return student, nil
}

// =====================================================
// CREATE
// =====================================================

func (r *studentRepository) Create(
	ctx context.Context,
	student model.Student,
) (model.Student, error) {

	query := `
		INSERT INTO students (
			nim,
			name,
			grade,
			is_active,
			owner_id
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING
			id,
			nim,
			name,
			grade,
			is_active,
			owner_id,
			created_at
	`

	var result model.Student

	err := r.db.QueryRow(
		ctx,
		query,
		student.NIM,
		student.Name,
		student.Grade,
		student.IsActive,
		student.OwnerID,
	).Scan(
		&result.ID,
		&result.NIM,
		&result.Name,
		&result.Grade,
		&result.IsActive,
		&result.OwnerID,
		&result.CreatedAt,
	)

	if err != nil {

		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) &&
			pgErr.Code == "23505" {

			return model.Student{}, ErrDuplicate
		}

		return model.Student{}, err
	}

	return result, nil
}

// =====================================================
// UPDATE
// =====================================================

func (r *studentRepository) Update(
	ctx context.Context,
	id int,
	student model.Student,
) (model.Student, error) {

	query := `
		UPDATE students
		SET
			nim = $1,
			name = $2,
			grade = $3,
			is_active = $4
		WHERE id = $5
		RETURNING
			id,
			nim,
			name,
			grade,
			is_active,
			owner_id,
			created_at
	`

	var result model.Student

	err := r.db.QueryRow(
		ctx,
		query,
		student.NIM,
		student.Name,
		student.Grade,
		student.IsActive,
		id,
	).Scan(
		&result.ID,
		&result.NIM,
		&result.Name,
		&result.Grade,
		&result.IsActive,
		&result.OwnerID,
		&result.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return model.Student{}, ErrNotFound
	}

	if err != nil {

		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) &&
			pgErr.Code == "23505" {

			return model.Student{}, ErrDuplicate
		}

		return model.Student{}, err
	}

	return result, nil
}

// =====================================================
// DELETE
// =====================================================

func (r *studentRepository) Delete(
	ctx context.Context,
	id int,
) error {

	query := `
		DELETE FROM students
		WHERE id = $1
	`

	result, err := r.db.Exec(
		ctx,
		query,
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