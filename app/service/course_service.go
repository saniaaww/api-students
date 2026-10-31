package service

import (
	"context"
	"errors"
	"fmt"

	"api-students/app/model"
	"api-students/app/repository"
)

const MaxCreditsPerSemester = 24

var errCreditLimit = errors.New(
	"batas maksimal SKS per semester terlampaui",
)

type CourseService struct {
	courses repository.CourseRepository
	outbox  repository.OutboxRepository
	tx      *repository.TxManager
}

func NewCourseService(
	courses repository.CourseRepository,
	outbox repository.OutboxRepository,
	tx *repository.TxManager,
) *CourseService {
	return &CourseService{
		courses: courses,
		outbox:  outbox,
		tx:      tx,
	}
}

// EnrollNaive adalah versi awal untuk mengamati race condition.
func (s *CourseService) EnrollNaive(
	ctx context.Context,
	courseID int,
	userID int,
) error {
	course, err := s.courses.FindByID(ctx, courseID)
	if err != nil {
		return err
	}

	if course.EnrolledCount >= course.Quota {
		return repository.ErrCourseFull
	}

	if err := s.courses.InsertEnrollment(
		ctx, courseID, userID,
	); err != nil {
		return err
	}

	return s.courses.SetEnrolledCount(
		ctx,
		courseID,
		course.EnrolledCount+1,
	)
}

// Enroll mendaftarkan mahasiswa menggunakan transaction.
func (s *CourseService) Enroll(
	ctx context.Context,
	courseID int,
	userID int,
) error {
	return s.tx.WithTx(ctx, func(tx repository.DBTX) error {
		repo := s.courses.WithTx(tx)

		if err := repo.LockStudent(ctx, userID); err != nil {
			return err
		}

		if err := repo.LockCourses(ctx, courseID); err != nil {
			return err
		}

		course, err := repo.FindByID(ctx, courseID)
		if err != nil {
			return err
		}

		taken, err := repo.SumCredits(ctx, userID)
		if err != nil {
			return err
		}

		if !WithinCreditLimit(taken, course.Credits) {
			return errCreditLimit
		}

		// Tambahkan peserta secara atomik.
		enrolled, err := repo.IncrementEnrolled(ctx, courseID)
		if err != nil {
			return err
		}

		// Catat event ke outbox dalam transaction yang sama.
		outboxRepo := s.outbox.WithTx(tx)
		if err := outboxRepo.Insert(
			ctx,
			model.EventEnrollmentCreated,
			model.EnrollmentEvent{
				UserID:     userID,
				CourseID:   enrolled.ID,
				CourseCode: enrolled.Code,
				CourseName: enrolled.Name,
			},
		); err != nil {
			return err
		}

		// Simpan pendaftaran mahasiswa.
		return repo.InsertEnrollment(ctx, courseID, userID)
	})
}

// Drop membatalkan pendaftaran mahasiswa menggunakan transaction.
func (s *CourseService) Drop(
	ctx context.Context,
	courseID int,
	userID int,
) error {
	return s.tx.WithTx(ctx, func(tx repository.DBTX) error {
		repo := s.courses.WithTx(tx)

		if err := repo.LockStudent(ctx, userID); err != nil {
			return err
		}

		if err := repo.LockCourses(ctx, courseID); err != nil {
			return err
		}

		if err := repo.DeleteEnrollment(
			ctx,
			courseID,
			userID,
		); err != nil {
			return err
		}

		// Kurangi jumlah peserta.
		course, err := repo.DecrementEnrolled(ctx, courseID)
		if err != nil {
			return err
		}

		// Catat event pembatalan pendaftaran.
		outboxRepo := s.outbox.WithTx(tx)
		return outboxRepo.Insert(
			ctx,
			model.EventEnrollmentDropped,
			model.EnrollmentEvent{
				UserID:     userID,
				CourseID:   course.ID,
				CourseCode: course.Code,
				CourseName: course.Name,
			},
		)
	})
}

// Swap memindahkan pendaftaran dari mata kuliah lama
// ke mata kuliah baru dalam satu transaction.
func (s *CourseService) Swap(
	ctx context.Context,
	userID int,
	req model.SwapRequest,
) (model.Course, error) {
	if req.FromCourseID <= 0 || req.ToCourseID <= 0 {
		return model.Course{}, errors.New(
			"ID mata kuliah harus berupa angka positif",
		)
	}

	if req.FromCourseID == req.ToCourseID {
		return model.Course{}, errors.New(
			"mata kuliah asal dan tujuan tidak boleh sama",
		)
	}

	var joined model.Course

	err := s.tx.WithTx(ctx, func(tx repository.DBTX) error {
		repo := s.courses.WithTx(tx)

		// Kunci mahasiswa terlebih dahulu.
		if err := repo.LockStudent(ctx, userID); err != nil {
			return err
		}

		// Kunci kedua mata kuliah dalam urutan ID menaik.
		if err := repo.LockCourses(
			ctx,
			req.FromCourseID,
			req.ToCourseID,
		); err != nil {
			return err
		}

		from, err := repo.FindByID(ctx, req.FromCourseID)
		if err != nil {
			return err
		}

		to, err := repo.FindByID(ctx, req.ToCourseID)
		if err != nil {
			return err
		}

		// Hitung total SKS yang sedang diambil.
		taken, err := repo.SumCredits(ctx, userID)
		if err != nil {
			return err
		}

		// Pastikan total SKS setelah pindah tidak melebihi batas.
		if CreditsAfterSwap(
			taken,
			from.Credits,
			to.Credits,
		) > MaxCreditsPerSemester {
			return errCreditLimit
		}

		// Hapus pendaftaran kelas lama.
		if err := repo.DeleteEnrollment(
			ctx,
			req.FromCourseID,
			userID,
		); err != nil {
			return err
		}

		// Kurangi jumlah peserta kelas lama.
		if _, err := repo.DecrementEnrolled(
			ctx,
			req.FromCourseID,
		); err != nil {
			return err
		}

		// Tambah peserta kelas baru secara atomik.
		joined, err = repo.IncrementEnrolled(
			ctx,
			req.ToCourseID,
		)
		if err != nil {
			return err
		}

		// Simpan pendaftaran kelas baru.
		if err := repo.InsertEnrollment(
			ctx,
			req.ToCourseID,
			userID,
		); err != nil {
			return err
		}

		// Catat satu event perpindahan kelas.
		outboxRepo := s.outbox.WithTx(tx)
		return outboxRepo.Insert(
			ctx,
			model.EventEnrollmentSwapped,
			model.EnrollmentEvent{
				UserID:         userID,
				CourseID:       to.ID,
				CourseCode:     to.Code,
				CourseName:     to.Name,
				FromCourseCode: from.Code,
				FromCourseName: from.Name,
			},
		)
	})

	if err != nil {
		return model.Course{}, err
	}

	return joined, nil
}

// UpdateQuota mengubah kuota dengan optimistic locking.
func (s *CourseService) UpdateQuota(
	ctx context.Context,
	courseID int,
	quota int,
	version int,
) (model.Course, error) {
	if courseID <= 0 {
		return model.Course{}, errors.New(
			"ID mata kuliah harus berupa angka positif",
		)
	}

	if quota < 1 || quota > 500 {
		return model.Course{}, errors.New(
			"kuota harus berada antara 1 dan 500",
		)
	}

	if version < 1 {
		return model.Course{}, errors.New(
			"version harus berupa angka positif",
		)
	}

	var updated model.Course

	err := s.tx.WithTx(ctx, func(tx repository.DBTX) error {
		courseRepo := s.courses.WithTx(tx)
		outboxRepo := s.outbox.WithTx(tx)

		if err := courseRepo.LockCourses(ctx, courseID); err != nil {
			return err
		}

		var err error
		updated, err = courseRepo.UpdateQuota(
			ctx,
			courseID,
			quota,
			version,
		)
		if err != nil {
			return err
		}

		// Snapshot penerima dalam transaction yang sama.
		userIDs, err := courseRepo.FindEnrolledUserIDs(
			ctx,
			courseID,
		)
		if err != nil {
			return err
		}

		event := model.CourseQuotaChangedEvent{
			CourseID:   updated.ID,
			CourseCode: updated.Code,
			CourseName: updated.Name,
			Quota:      updated.Quota,
			UserIDs:    userIDs,
		}

		return outboxRepo.Insert(
			ctx,
			model.EventCourseQuotaChanged,
			event,
		)
	})

	if err != nil {
		return model.Course{}, fmt.Errorf(
			"memperbarui kuota dan mencatat event: %w",
			err,
		)
	}

	return updated, nil
}

// Courses mengambil daftar seluruh mata kuliah.
func (s *CourseService) Courses(
	ctx context.Context,
) ([]model.Course, error) {
	return s.courses.FindAll(ctx)
}

// CourseByID mengambil mata kuliah berdasarkan ID.
func (s *CourseService) CourseByID(
	ctx context.Context,
	id int,
) (model.Course, error) {
	course, err := s.courses.FindByID(ctx, id)
	if err != nil {
		return model.Course{}, fmt.Errorf(
			"gagal mengambil mata kuliah: %w",
			err,
		)
	}

	return course, nil
}

// WithinCreditLimit memeriksa batas maksimal SKS.
func WithinCreditLimit(taken, adding int) bool {
	return taken+adding <= MaxCreditsPerSemester
}

// CreditsAfterSwap menghitung total SKS setelah pindah kelas.
func CreditsAfterSwap(
	taken int,
	leaving int,
	joining int,
) int {
	return taken - leaving + joining
}
