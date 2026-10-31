
package service

import (
	"context"
	"errors"
	"fmt"

	"api-students/app/model"
	"api-students/app/repository"
)

type SupervisorService struct {
	supervisors repository.SupervisorRepository
	tx          *repository.TxManager
}

func NewSupervisorService(
	supervisors repository.SupervisorRepository,
	tx *repository.TxManager,
) *SupervisorService {
	return &SupervisorService{
		supervisors: supervisors,
		tx:          tx,
	}
}

// Assign menetapkan pembimbing kepada mahasiswa dalam satu transaksi.
func (s *SupervisorService) Assign(
	ctx context.Context,
	supervisorID int,
	studentID int,
) error {
	if supervisorID <= 0 || studentID <= 0 {
		return errors.New("ID pembimbing dan mahasiswa harus positif")
	}

	return s.tx.WithTx(ctx, func(tx repository.DBTX) error {
		repo := s.supervisors.WithTx(tx)

		// Kunci mahasiswa agar dua permintaan penetapan
		// untuk mahasiswa yang sama tidak berjalan bersamaan.
		if err := repo.LockStudent(ctx, studentID); err != nil {
			return err
		}

		// Pastikan pembimbing tersedia dan terkunci.
		if err := repo.LockSupervisors(ctx, supervisorID); err != nil {
			return err
		}

		// Tambahkan jumlah bimbingan dan catat penetapan.
		// Keduanya harus berhasil dalam transaksi yang sama.
		return repo.Assign(ctx, studentID, supervisorID)
	})
}

// Switch mengganti pembimbing secara atomik.
func (s *SupervisorService) Switch(
	ctx context.Context,
	req model.SwitchSupervisorRequest,
) error {
	if req.StudentID <= 0 || req.NewSupervisorID <= 0 {
		return errors.New("ID mahasiswa dan pembimbing baru harus positif")
	}

	return s.tx.WithTx(ctx, func(tx repository.DBTX) error {
		repo := s.supervisors.WithTx(tx)

		// Urutan penguncian dimulai dari mahasiswa.
		if err := repo.LockStudent(ctx, req.StudentID); err != nil {
			return err
		}

		// Switch pada repository menentukan pembimbing lama,
		// kemudian mengunci kedua supervisor berdasarkan ID menaik.
		return repo.Switch(
			ctx,
			req.StudentID,
			req.NewSupervisorID,
		)
	})
}

// UpdateQuota mengubah kuota menggunakan optimistic locking.
func (s *SupervisorService) UpdateQuota(
	ctx context.Context,
	supervisorID int,
	quota int,
	version int,
) (model.Supervisor, error) {
	if supervisorID <= 0 {
		return model.Supervisor{}, errors.New(
			"ID pembimbing harus berupa angka positif",
		)
	}

	if quota < 0 {
		return model.Supervisor{}, errors.New(
			"kuota tidak boleh negatif",
		)
	}

	if version < 1 {
		return model.Supervisor{}, errors.New(
			"version harus berupa angka positif",
		)
	}

	var updated model.Supervisor

	err := s.tx.WithTx(ctx, func(tx repository.DBTX) error {
		repo := s.supervisors.WithTx(tx)

		var err error
		updated, err = repo.UpdateQuota(
			ctx,
			supervisorID,
			quota,
			version,
		)
		return err
	})
	if err != nil {
		return model.Supervisor{}, fmt.Errorf(
			"memperbarui kuota pembimbing: %w",
			err,
		)
	}

	return updated, nil
}

// SupervisorByID mengambil data pembimbing berdasarkan ID.
func (s *SupervisorService) SupervisorByID(
	ctx context.Context,
	id int,
) (model.Supervisor, error) {
	if id <= 0 {
		return model.Supervisor{}, errors.New(
			"ID pembimbing harus berupa angka positif",
		)
	}

	supervisor, err := s.supervisors.FindByID(ctx, id)
	if err != nil {
		return model.Supervisor{}, fmt.Errorf(
			"mengambil data pembimbing: %w",
			err,
		)
	}

	return supervisor, nil
}
