package repository

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"api-students/app/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CourseRepository interface {
	FindAll(
		ctx context.Context,
		semester *int,
		search string,
		available bool,
	) ([]model.Course, error)

	FindByID(
		ctx context.Context,
		id int,
	) (model.Course, error)
}

type courseRepository struct {
	db *pgxpool.Pool
}

func NewCourseRepository(db *pgxpool.Pool) CourseRepository {
	return &courseRepository{
		db: db,
	}
}

// =====================================================
// FIND ALL
// GET /api/v1/courses
// =====================================================

func (r *courseRepository) FindAll(
	ctx context.Context,
	semester *int,
	search string,
	available bool,
) ([]model.Course, error) {

	conditions := []string{
		"TRUE",
	}

	args := make([]interface{}, 0)
	argNumber := 1

	if semester != nil {
		conditions = append(
			conditions,
			"c.semester = $" + strconv.Itoa(argNumber),
		)

		args = append(args, *semester)
		argNumber++
	}

	if search != "" {
		conditions = append(
			conditions,
			"(c.kode_mk ILIKE $" + strconv.Itoa(argNumber) +
	" OR c.nama_mk ILIKE $" + strconv.Itoa(argNumber) + ")",
		)

		args = append(args, "%"+search+"%")
		argNumber++
	}

	query := `
		SELECT
			c.id,
			c.kode_mk,
			c.nama_mk,
			c.sks,
			c.semester,
			c.kuota,
			COUNT(e.id)::int AS terisi,
			(c.kuota - COUNT(e.id))::int AS sisa_kuota
		FROM courses c
		LEFT JOIN enrollments e
			ON e.course_id = c.id
		WHERE ` + strings.Join(conditions, " AND ") + `
		GROUP BY
			c.id,
			c.kode_mk,
			c.nama_mk,
			c.sks,
			c.semester,
			c.kuota
	`

	if available {
		query += `
		HAVING COUNT(e.id) < c.kuota
		`
	}

	query += `
		ORDER BY c.semester ASC, c.kode_mk ASC
		`

	rows, err := r.db.Query(
		ctx,
		query,
		args...,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	courses := make([]model.Course, 0)

	for rows.Next() {
		var course model.Course

		if err := rows.Scan(
			&course.ID,
			&course.KodeMK,
			&course.NamaMK,
			&course.SKS,
			&course.Semester,
			&course.Kuota,
			&course.Terisi,
			&course.SisaKuota,
		); err != nil {
			return nil, err
		}

		courses = append(courses, course)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return courses, nil
}

// =====================================================
// FIND BY ID
// =====================================================

func (r *courseRepository) FindByID(
	ctx context.Context,
	id int,
) (model.Course, error) {

	var course model.Course

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			id,
			kode_mk,
			nama_mk,
			sks,
			semester,
			kuota
		FROM courses
		WHERE id = $1
		`,
		id,
	).Scan(
		&course.ID,
		&course.KodeMK,
		&course.NamaMK,
		&course.SKS,
		&course.Semester,
		&course.Kuota,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return model.Course{}, ErrNotFound
	}

	if err != nil {
		return model.Course{}, err
	}

	return course, nil
}