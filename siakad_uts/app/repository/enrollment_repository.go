package repository

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"api-students/app/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrEnrollmentNotFound = errors.New("enrollment tidak ditemukan")
	ErrEnrollmentDuplicate = errors.New("mata kuliah sudah diambil")
	ErrCourseFull          = errors.New("kuota mata kuliah penuh")
	ErrSKSLimit            = errors.New("batas SKS terlampaui")
	ErrEnrollmentForbidden = errors.New("enrollment bukan milik mahasiswa")
)

type EnrollmentRepository interface {
	FindByStudentID(
		ctx context.Context,
		studentID int,
	) ([]model.EnrollmentDetail, error)

	TotalSKS(
		ctx context.Context,
		studentID int,
	) (int, error)

	Create(
		ctx context.Context,
		studentID int,
		courseID int,
		tahunAkademik string,
		maxSKS int,
	) (model.Enrollment, error)

	DeleteOwned(
		ctx context.Context,
		enrollmentID int,
		studentID int,
	) error
}

type enrollmentRepository struct {
	db *pgxpool.Pool
}

func NewEnrollmentRepository(
	db *pgxpool.Pool,
) EnrollmentRepository {
	return &enrollmentRepository{
		db: db,
	}
}

// =====================================================
// FIND BY STUDENT
// Untuk GET /students/{id}
// =====================================================

func (r *enrollmentRepository) FindByStudentID(
	ctx context.Context,
	studentID int,
) ([]model.EnrollmentDetail, error) {

	rows, err := r.db.Query(
		ctx,
		`
		SELECT
			e.id,
			e.course_id,
			c.kode_mk,
			c.nama_mk,
			c.sks,
			c.semester,
			e.tahun_akademik
		FROM enrollments e
		JOIN courses c
			ON c.id = e.course_id
		WHERE e.student_id = $1
		ORDER BY e.created_at ASC, e.id ASC
		`,
		studentID,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	result := make([]model.EnrollmentDetail, 0)

	for rows.Next() {
		var item model.EnrollmentDetail

		if err := rows.Scan(
			&item.ID,
			&item.CourseID,
			&item.KodeMK,
			&item.NamaMK,
			&item.SKS,
			&item.Semester,
			&item.TahunAkademik,
		); err != nil {
			return nil, err
		}

		result = append(result, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// =====================================================
// TOTAL SKS
// =====================================================

func (r *enrollmentRepository) TotalSKS(
	ctx context.Context,
	studentID int,
) (int, error) {

	var total int

	err := r.db.QueryRow(
		ctx,
		`
		SELECT COALESCE(SUM(c.sks), 0)::int
		FROM enrollments e
		JOIN courses c
			ON c.id = e.course_id
		WHERE e.student_id = $1
		`,
		studentID,
	).Scan(&total)

	if err != nil {
		return 0, err
	}

	return total, nil
}


func (r *enrollmentRepository) Create(
	ctx context.Context,
	studentID int,
	courseID int,
	tahunAkademik string,
	maxSKS int,
) (model.Enrollment, error) {

	tx, err := r.db.Begin(ctx)

	if err != nil {
		return model.Enrollment{}, err
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	// -------------------------------------------------
	// LOCK COURSE
	// -------------------------------------------------

	var (
		courseSKS int
		courseQuota int
	)

	err = tx.QueryRow(
		ctx,
		`
		SELECT
			sks,
			kuota
		FROM courses
		WHERE id = $1
		FOR UPDATE
		`,
		courseID,
	).Scan(
		&courseSKS,
		&courseQuota,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return model.Enrollment{}, ErrNotFound
	}

	if err != nil {
		return model.Enrollment{}, err
	}

	// -------------------------------------------------
	// CEK DUPLIKASI
	// -------------------------------------------------

	var exists bool

	err = tx.QueryRow(
		ctx,
		`
		SELECT EXISTS (
			SELECT 1
			FROM enrollments
			WHERE student_id = $1
			  AND course_id = $2
			  AND tahun_akademik = $3
		)
		`,
		studentID,
		courseID,
		tahunAkademik,
	).Scan(&exists)

	if err != nil {
		return model.Enrollment{}, err
	}

	if exists {
		return model.Enrollment{}, ErrEnrollmentDuplicate
	}

	// -------------------------------------------------
	// CEK KUOTA
	// -------------------------------------------------

	var terisi int

	err = tx.QueryRow(
		ctx,
		`
		SELECT COUNT(*)
		FROM enrollments
		WHERE course_id = $1
		`,
		courseID,
	).Scan(&terisi)

	if err != nil {
		return model.Enrollment{}, err
	}

	if terisi >= courseQuota {
		return model.Enrollment{}, ErrCourseFull
	}

	// -------------------------------------------------
	// CEK TOTAL SKS
	// -------------------------------------------------

	var currentSKS int

	err = tx.QueryRow(
		ctx,
		`
		SELECT COALESCE(SUM(c.sks), 0)::int
		FROM enrollments e
		JOIN courses c
			ON c.id = e.course_id
		WHERE e.student_id = $1
		`,
		studentID,
	).Scan(&currentSKS)

	if err != nil {
		return model.Enrollment{}, err
	}

	if currentSKS+courseSKS > maxSKS {
		return model.Enrollment{}, fmt.Errorf(
			"%w: sisa SKS %d",
			ErrSKSLimit,
			maxSKS-currentSKS,
		)
	}

	// -------------------------------------------------
	// INSERT
	// -------------------------------------------------

	var result model.Enrollment

	err = tx.QueryRow(
		ctx,
		`
		INSERT INTO enrollments (
			student_id,
			course_id,
			tahun_akademik
		)
		VALUES ($1, $2, $3)
		RETURNING
			id,
			student_id,
			course_id,
			tahun_akademik,
			created_at
		`,
		studentID,
		courseID,
		tahunAkademik,
	).Scan(
		&result.ID,
		&result.StudentID,
		&result.CourseID,
		&result.TahunAkademik,
		&result.CreatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) &&
			pgErr.Code == "23505" {

			return model.Enrollment{}, ErrEnrollmentDuplicate
		}

		return model.Enrollment{}, err
	}

	// -------------------------------------------------
	// COMMIT
	// -------------------------------------------------

	if err := tx.Commit(ctx); err != nil {
		return model.Enrollment{}, err
	}

	return result, nil
}

// =====================================================
// DELETE OWNED ENROLLMENT
// =====================================================

func (r *enrollmentRepository) DeleteOwned(
	ctx context.Context,
	enrollmentID int,
	studentID int,
) error {

	var ownerID int

	err := r.db.QueryRow(
		ctx,
		`
		SELECT student_id
		FROM enrollments
		WHERE id = $1
		`,
		enrollmentID,
	).Scan(&ownerID)

	if errors.Is(err, pgx.ErrNoRows) {
		return ErrEnrollmentNotFound
	}

	if err != nil {
		return err
	}

	if ownerID != studentID {
		return ErrEnrollmentForbidden
	}

	result, err := r.db.Exec(
		ctx,
		`
		DELETE FROM enrollments
		WHERE id = $1
		  AND student_id = $2
		`,
		enrollmentID,
		studentID,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrEnrollmentNotFound
	}

	return nil
}

// compile-time check untuk format tahun akademik.
// Validasi utama tetap akan dilakukan di service.
var academicYearPattern = regexp.MustCompile(
	`^\d{4}/\d{4} (Ganjil|Genap)$`,
)