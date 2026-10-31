
package repository

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"api-students/app/model"

	"github.com/jackc/pgx/v5"
)

var (
	ErrSupervisorFull = errors.New("kuota pembimbing sudah penuh")
	ErrAlreadyAssigned = errors.New("mahasiswa sudah memiliki pembimbing")
	ErrSupervisorVersionConflict = errors.New(
		"versi data pembimbing sudah berubah",
	)
	ErrQuotaBelowSupervised = errors.New(
		"kuota tidak boleh lebih kecil dari jumlah mahasiswa bimbingan",
	)
)

type SupervisorRepository interface {
	WithTx(tx DBTX) SupervisorRepository

	FindByID(ctx context.Context, id int) (model.Supervisor, error)
	LockStudent(ctx context.Context, studentID int) error
	LockSupervisors(ctx context.Context, ids ...int) error

	Assign(ctx context.Context, studentID, supervisorID int) error
	Switch(ctx context.Context, studentID, newSupervisorID int) error

	UpdateQuota(
		ctx context.Context,
		id, quota, version int,
	) (model.Supervisor, error)
}

const supervisorColumns = `id, name, quota, supervised_count, version`

type supervisorPostgresRepository struct {
	db DBTX
}

func NewSupervisorRepository(db DBTX) SupervisorRepository {
	return &supervisorPostgresRepository{db: db}
}

func (r *supervisorPostgresRepository) WithTx(
	tx DBTX,
) SupervisorRepository {
	return &supervisorPostgresRepository{db: tx}
}

func scanSupervisor(row pgx.Row) (model.Supervisor, error) {
	var s model.Supervisor
	err := row.Scan(
		&s.ID,
		&s.Name,
		&s.Quota,
		&s.SupervisedCount,
		&s.Version,
	)
	return s, err
}

func (r *supervisorPostgresRepository) FindByID(
	ctx context.Context,
	id int,
) (model.Supervisor, error) {
	s, err := scanSupervisor(r.db.QueryRow(
		ctx,
		`SELECT `+supervisorColumns+`
		 FROM supervisors
		 WHERE id = $1`,
		id,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Supervisor{}, ErrNotFound
	}
	if err != nil {
		return model.Supervisor{}, fmt.Errorf(
			"mengambil data pembimbing: %w", err,
		)
	}
	return s, nil
}

func (r *supervisorPostgresRepository) LockStudent(
	ctx context.Context,
	studentID int,
) error {
	var id int
	err := r.db.QueryRow(
		ctx,
		`SELECT id
		 FROM students
		 WHERE id = $1
		 FOR UPDATE`,
		studentID,
	).Scan(&id)

	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("mengunci mahasiswa: %w", err)
	}
	return nil
}

// LockSupervisors mengunci baris pembimbing berdasarkan urutan ID.
func (r *supervisorPostgresRepository) LockSupervisors(
	ctx context.Context,
	ids ...int,
) error {
	if len(ids) == 0 {
		return nil
	}

	ordered := append([]int(nil), ids...)
	sort.Ints(ordered)

	// Hilangkan ID duplikat agar jumlah baris yang diharapkan tepat.
	unique := ordered[:0]
	for _, id := range ordered {
		if len(unique) == 0 || unique[len(unique)-1] != id {
			unique = append(unique, id)
		}
	}

	rows, err := r.db.Query(
		ctx,
		`SELECT id
		 FROM supervisors
		 WHERE id = ANY($1)
		 ORDER BY id
		 FOR UPDATE`,
		unique,
	)
	if err != nil {
		return fmt.Errorf("mengunci pembimbing: %w", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return fmt.Errorf("membaca pembimbing terkunci: %w", err)
		}
		count++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("membaca pembimbing terkunci: %w", err)
	}
	if count != len(unique) {
		return ErrNotFound
	}
	return nil
}

// Assign mengasumsikan pemanggil sudah mengunci mahasiswa dan
// menjalankan seluruh operasi melalui TxManager.
func (r *supervisorPostgresRepository) Assign(
	ctx context.Context,
	studentID, supervisorID int,
) error {
	// UPDATE bersyarat: tidak mengambil kuota jika sudah penuh.
	tag, err := r.db.Exec(
		ctx,
		`UPDATE supervisors
		 SET supervised_count = supervised_count + 1
		 WHERE id = $1
		   AND supervised_count < quota`,
		supervisorID,
	)
	if err != nil {
		return fmt.Errorf("menambah jumlah bimbingan: %w", err)
	}

	if tag.RowsAffected() == 0 {
		_, findErr := r.FindByID(ctx, supervisorID)
		if findErr != nil {
			return findErr
		}
		return ErrSupervisorFull
	}

	_, err = r.db.Exec(
		ctx,
		`INSERT INTO supervisions (student_id, supervisor_id)
		 VALUES ($1, $2)`,
		studentID,
		supervisorID,
	)
	if err != nil {
		if pgErrorCode(err) == codeUniqueViolation {
			return ErrAlreadyAssigned
		}
		return fmt.Errorf("menetapkan pembimbing: %w", err)
	}
	return nil
}

// Switch mengganti pembimbing. Pemanggil harus mengunci mahasiswa
// dan menjalankan method ini di dalam TxManager.WithTx.
func (r *supervisorPostgresRepository) Switch(
	ctx context.Context,
	studentID, newSupervisorID int,
) error {
	var oldSupervisorID int
	err := r.db.QueryRow(
		ctx,
		`SELECT supervisor_id
		 FROM supervisions
		 WHERE student_id = $1`,
		studentID,
	).Scan(&oldSupervisorID)

	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotEnrolled
	}
	if err != nil {
		return fmt.Errorf("mencari pembimbing lama: %w", err)
	}
	if oldSupervisorID == newSupervisorID {
		return ErrAlreadyAssigned
	}

	// Semua transaksi mengunci kedua pembimbing berdasarkan ID,
	// bukan berdasarkan arah perpindahan.
	if err := r.LockSupervisors(
		ctx, oldSupervisorID, newSupervisorID,
	); err != nil {
		return err
	}

	tag, err := r.db.Exec(
		ctx,
		`UPDATE supervisors
		 SET supervised_count = supervised_count - 1
		 WHERE id = $1
		   AND supervised_count > 0`,
		oldSupervisorID,
	)
	if err != nil {
		return fmt.Errorf("mengurangi hitungan pembimbing lama: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("jumlah bimbingan pembimbing lama tidak valid")
	}

	tag, err = r.db.Exec(
		ctx,
		`UPDATE supervisors
		 SET supervised_count = supervised_count + 1
		 WHERE id = $1
		   AND supervised_count < quota`,
		newSupervisorID,
	)
	if err != nil {
		return fmt.Errorf("menambah hitungan pembimbing baru: %w", err)
	}
	if tag.RowsAffected() == 0 {
		_, findErr := r.FindByID(ctx, newSupervisorID)
		if findErr != nil {
			return findErr
		}
		return ErrSupervisorFull
	}

	tag, err = r.db.Exec(
		ctx,
		`UPDATE supervisions
		 SET supervisor_id = $1
		 WHERE student_id = $2
		   AND supervisor_id = $3`,
		newSupervisorID,
		studentID,
		oldSupervisorID,
	)
	if err != nil {
		return fmt.Errorf("mengganti relasi pembimbing: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotEnrolled
	}
	return nil
}

func (r *supervisorPostgresRepository) UpdateQuota(
	ctx context.Context,
	id, quota, version int,
) (model.Supervisor, error) {
	if quota < 0 {
		return model.Supervisor{}, ErrQuotaBelowSupervised
	}

	s, err := scanSupervisor(r.db.QueryRow(
		ctx,
		`UPDATE supervisors
		 SET quota = $1,
		     version = version + 1
		 WHERE id = $2
		   AND version = $3
		   AND $1 >= supervised_count
		 RETURNING `+supervisorColumns,
		quota,
		id,
		version,
	))
	if err == nil {
		return s, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return model.Supervisor{}, fmt.Errorf(
			"memperbarui kuota pembimbing: %w", err,
		)
	}

	current, findErr := r.FindByID(ctx, id)
	if findErr != nil {
		return model.Supervisor{}, findErr
	}
	if current.Version != version {
		return model.Supervisor{}, ErrSupervisorVersionConflict
	}
	return model.Supervisor{}, ErrQuotaBelowSupervised
}
