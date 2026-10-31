package repository

import (
	"context"
	"errors"
	"fmt"

	"api-students/app/model"

	"github.com/jackc/pgx/v5"
)

var (
	ErrCourseFull       = errors.New("kuota kelas sudah penuh")
	ErrNotEnrolled      = errors.New("mahasiswa belum terdaftar di kelas")
	ErrVersionConflict  = errors.New("versi kelas sudah berubah")
	ErrQuotaBelowFilled = errors.New("kuota tidak boleh lebih kecil dari jumlah peserta")
)

type CourseRepository interface {
	WithTx(tx DBTX) CourseRepository

	FindAll(ctx context.Context) ([]model.Course, error)
	FindByID(ctx context.Context, id int) (model.Course, error)
	FindByStudent(ctx context.Context, userID int) ([]model.Course, error)
	FindEnrolledUserIDs(ctx context.Context, courseID int) ([]int, error)

	LockStudent(ctx context.Context, userID int) error
	LockCourses(ctx context.Context, ids ...int) error
	SumCredits(ctx context.Context, userID int) (int, error)

	InsertEnrollment(ctx context.Context, courseID, userID int) error
	SetEnrolledCount(ctx context.Context, courseID, count int) error

	IncrementEnrolled(ctx context.Context, courseID int) (model.Course, error)
	DecrementEnrolled(ctx context.Context, courseID int) (model.Course, error)
	DeleteEnrollment(ctx context.Context, courseID, userID int) error

	UpdateQuota(
		ctx context.Context,
		id, quota, version int,
	) (model.Course, error)
}

const courseColumns = `id, code, name, credits, quota, enrolled_count, version, created_at`

type coursePostgresRepository struct {
	db DBTX
}

func NewCourseRepository(db DBTX) CourseRepository {
	return &coursePostgresRepository{db: db}
}

func (r *coursePostgresRepository) WithTx(tx DBTX) CourseRepository {
	return &coursePostgresRepository{db: tx}
}

func scanCourse(row pgx.Row) (model.Course, error) {
	var c model.Course

	err := row.Scan(
		&c.ID,
		&c.Code,
		&c.Name,
		&c.Credits,
		&c.Quota,
		&c.EnrolledCount,
		&c.Version,
		&c.CreatedAt,
	)

	return c, err
}

func (r *coursePostgresRepository) collect(
	rows pgx.Rows,
	err error,
) ([]model.Course, error) {
	if err != nil {
		return nil, fmt.Errorf("mengambil daftar kelas: %w", err)
	}
	defer rows.Close()

	result := []model.Course{}

	for rows.Next() {
		c, err := scanCourse(rows)
		if err != nil {
			return nil, fmt.Errorf("membaca data kelas: %w", err)
		}

		result = append(result, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("membaca daftar kelas: %w", err)
	}

	return result, nil
}

func (r *coursePostgresRepository) FindAll(
	ctx context.Context,
) ([]model.Course, error) {
	return r.collect(r.db.Query(
		ctx,
		"SELECT "+courseColumns+" FROM courses ORDER BY code",
	))
}

func (r *coursePostgresRepository) FindByID(
	ctx context.Context,
	id int,
) (model.Course, error) {
	c, err := scanCourse(r.db.QueryRow(
		ctx,
		"SELECT "+courseColumns+" FROM courses WHERE id = $1",
		id,
	))

	if errors.Is(err, pgx.ErrNoRows) {
		return model.Course{}, ErrNotFound
	}
	if err != nil {
		return model.Course{}, fmt.Errorf("mengambil kelas: %w", err)
	}

	return c, nil
}

func (r *coursePostgresRepository) FindEnrolledUserIDs(
	ctx context.Context,
	courseID int,
) ([]int, error) {
	rows, err := r.db.Query(ctx, `SELECT user_id FROM enrollments WHERE course_id = $1 ORDER BY user_id`, courseID)
	if err != nil {
		return nil, fmt.Errorf("mengambil peserta kelas: %w", err)
	}
	defer rows.Close()

	ids := make([]int, 0)
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("membaca peserta kelas: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("membaca daftar peserta kelas: %w", err)
	}
	return ids, nil
}

func (r *coursePostgresRepository) FindByStudent(
	ctx context.Context,
	userID int,
) ([]model.Course, error) {
	return r.collect(r.db.Query(
		ctx,
		`SELECT `+courseColumns+`
		 FROM courses c
		 JOIN enrollments e ON e.course_id = c.id
		 WHERE e.user_id = $1
		 ORDER BY c.code`,
		userID,
	))
}

// LockStudent mengunci baris mahasiswa sampai transaksi selesai.
// FOR UPDATE mencegah transaksi lain memproses mahasiswa yang sama
// secara bersamaan ketika menghitung total SKS.
func (r *coursePostgresRepository) LockStudent(
	ctx context.Context,
	userID int,
) error {
	var id int

	err := r.db.QueryRow(
		ctx,
		`SELECT id
		 FROM users
		 WHERE id = $1
		 FOR UPDATE`,
		userID,
	).Scan(&id)

	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("mengunci data mahasiswa: %w", err)
	}

	return nil
}

// LockCourses mengunci baris kelas berdasarkan urutan ID.
// Urutan yang konsisten membantu mengurangi risiko deadlock.
func (r *coursePostgresRepository) LockCourses(
	ctx context.Context,
	ids ...int,
) error {
	if len(ids) == 0 {
		return nil
	}

	rows, err := r.db.Query(
		ctx,
		`SELECT id
		 FROM courses
		 WHERE id = ANY($1)
		 ORDER BY id
		 FOR UPDATE`,
		ids,
	)
	if err != nil {
		return fmt.Errorf("mengunci data kelas: %w", err)
	}
	defer rows.Close()

	count := 0

	for rows.Next() {
		var id int

		if err := rows.Scan(&id); err != nil {
			return fmt.Errorf("membaca kelas yang dikunci: %w", err)
		}

		count++
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("membaca data kelas yang dikunci: %w", err)
	}

	if count != len(ids) {
		return ErrNotFound
	}

	return nil
}

// SumCredits menghitung total SKS yang telah diambil mahasiswa.
func (r *coursePostgresRepository) SumCredits(
	ctx context.Context,
	userID int,
) (int, error) {
	var total int

	err := r.db.QueryRow(
		ctx,
		`SELECT COALESCE(SUM(c.credits), 0)
		 FROM enrollments e
		 JOIN courses c ON c.id = e.course_id
		 WHERE e.user_id = $1`,
		userID,
	).Scan(&total)

	if err != nil {
		return 0, fmt.Errorf("menghitung SKS: %w", err)
	}

	return total, nil
}

// InsertEnrollment mencatat pendaftaran mahasiswa ke kelas.
func (r *coursePostgresRepository) InsertEnrollment(
	ctx context.Context,
	courseID, userID int,
) error {
	_, err := r.db.Exec(
		ctx,
		`INSERT INTO enrollments (course_id, user_id)
		 VALUES ($1, $2)`,
		courseID,
		userID,
	)

	if err != nil {
		if pgErrorCode(err) == codeUniqueViolation {
			return ErrDuplicate
		}

		return fmt.Errorf("mencatat pendaftaran: %w", err)
	}

	return nil
}

// SetEnrolledCount dipertahankan untuk eksperimen EnrollNaive.
func (r *coursePostgresRepository) SetEnrolledCount(
	ctx context.Context,
	courseID, count int,
) error {
	_, err := r.db.Exec(
		ctx,
		`UPDATE courses
		 SET enrolled_count = $1
		 WHERE id = $2`,
		count,
		courseID,
	)
	if err != nil {
		return fmt.Errorf("menyimpan jumlah peserta: %w", err)
	}

	return nil
}

// IncrementEnrolled menambah peserta secara atomik.
// Pemeriksaan kuota dan penambahan peserta dilakukan dalam satu UPDATE.
func (r *coursePostgresRepository) IncrementEnrolled(
	ctx context.Context,
	courseID int,
) (model.Course, error) {
	c, err := scanCourse(r.db.QueryRow(
		ctx,
		`UPDATE courses
		 SET enrolled_count = enrolled_count + 1
		 WHERE id = $1
		   AND enrolled_count < quota
		 RETURNING `+courseColumns,
		courseID,
	))

	if errors.Is(err, pgx.ErrNoRows) {
		var exists bool

		err = r.db.QueryRow(
			ctx,
			`SELECT EXISTS (
				SELECT 1 FROM courses WHERE id = $1
			)`,
			courseID,
		).Scan(&exists)

		if err != nil {
			return model.Course{}, fmt.Errorf("memeriksa kelas: %w", err)
		}

		if !exists {
			return model.Course{}, ErrNotFound
		}

		return model.Course{}, ErrCourseFull
	}

	if err != nil {
		return model.Course{}, fmt.Errorf("menambah jumlah peserta: %w", err)
	}

	return c, nil
}

// DecrementEnrolled mengurangi peserta dan mengembalikan data kelas terbaru.
func (r *coursePostgresRepository) DecrementEnrolled(
	ctx context.Context,
	courseID int,
) (model.Course, error) {
	c, err := scanCourse(r.db.QueryRow(
		ctx,
		`UPDATE courses
		 SET enrolled_count = enrolled_count - 1
		 WHERE id = $1
		   AND enrolled_count > 0
		 RETURNING `+courseColumns,
		courseID,
	))

	if errors.Is(err, pgx.ErrNoRows) {
		var exists bool

		err = r.db.QueryRow(
			ctx,
			`SELECT EXISTS (
				SELECT 1 FROM courses WHERE id = $1
			)`,
			courseID,
		).Scan(&exists)

		if err != nil {
			return model.Course{}, fmt.Errorf("memeriksa kelas: %w", err)
		}

		if !exists {
			return model.Course{}, ErrNotFound
		}

		return model.Course{}, ErrNotEnrolled
	}

	if err != nil {
		return model.Course{}, fmt.Errorf("mengurangi jumlah peserta: %w", err)
	}

	return c, nil
}

// DeleteEnrollment memastikan pendaftaran memang ada.
func (r *coursePostgresRepository) DeleteEnrollment(
	ctx context.Context,
	courseID, userID int,
) error {
	tag, err := r.db.Exec(
		ctx,
		`DELETE FROM enrollments
		 WHERE course_id = $1
		   AND user_id = $2`,
		courseID,
		userID,
	)
	if err != nil {
		return fmt.Errorf("menghapus pendaftaran: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrNotEnrolled
	}

	return nil
}

// UpdateQuota menggunakan optimistic locking melalui kolom version.
// Pembaruan hanya berhasil jika versi masih sama dan kuota tidak
// lebih kecil daripada jumlah peserta yang sudah terdaftar.
func (r *coursePostgresRepository) UpdateQuota(
	ctx context.Context,
	id, quota, version int,
) (model.Course, error) {
	c, err := scanCourse(r.db.QueryRow(
		ctx,
		`UPDATE courses
		 SET quota = $1,
		     version = version + 1
		 WHERE id = $2
		   AND version = $3
		   AND $1 >= enrolled_count
		 RETURNING `+courseColumns,
		quota,
		id,
		version,
	))

	if err == nil {
		return c, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return model.Course{}, fmt.Errorf("memperbarui kuota kelas: %w", err)
	}

	// Periksa kondisi terkini untuk membedakan konflik versi
	// dari kuota yang lebih kecil daripada jumlah peserta.
	current, findErr := r.FindByID(ctx, id)
	if findErr != nil {
		return model.Course{}, findErr
	}

	if current.Version != version {
		return model.Course{}, ErrVersionConflict
	}

	return model.Course{}, ErrQuotaBelowFilled
}
