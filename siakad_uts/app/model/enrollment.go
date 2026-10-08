package model

import "time"

type Enrollment struct {
	ID             int       `json:"id"`
	StudentID      int       `json:"student_id"`
	CourseID       int       `json:"course_id"`
	TahunAkademik  string    `json:"tahun_akademik"`
	CreatedAt      time.Time `json:"created_at"`
}

type CreateEnrollmentRequest struct {
	CourseID      int    `json:"course_id" validate:"required,gt=0"`
	TahunAkademik string `json:"tahun_akademik" validate:"required"`
}

type EnrollmentDetail struct {
	ID             int     `json:"id"`
	CourseID       int     `json:"course_id"`
	KodeMK         string  `json:"kode_mk"`
	NamaMK         string  `json:"nama_mk"`
	SKS            int     `json:"sks"`
	Semester       int     `json:"semester"`
	TahunAkademik  string  `json:"tahun_akademik"`
}