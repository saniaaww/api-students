package model

import "time"

type Student struct {
	ID          int        `json:"id"`
	UserID      int        `json:"user_id"`
	NIM         string     `json:"nim"`
	Nama        string     `json:"nama"`
	Prodi       string     `json:"prodi"`
	Angkatan    int        `json:"angkatan"`
	IPKTerakhir *float64   `json:"ipk_terakhir,omitempty"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

type CreateStudentRequest struct {
	NIM         string   `json:"nim" validate:"required,len=12,numeric"`
	Nama        string   `json:"nama" validate:"required,min=2,max=100"`
	Email       string   `json:"email" validate:"required,email,max=120"`
	Prodi       string   `json:"prodi" validate:"required,min=2,max=100"`
	Angkatan    int      `json:"angkatan" validate:"required"`
	IPKTerakhir *float64 `json:"ipk_terakhir,omitempty" validate:"omitempty,gte=0,lte=4"`
}

type UpdateStudentRequest struct {
	Nama        string   `json:"nama" validate:"required,min=2,max=100"`
	Prodi       string   `json:"prodi" validate:"required,min=2,max=100"`
	Angkatan    int      `json:"angkatan" validate:"required"`
	IPKTerakhir *float64 `json:"ipk_terakhir,omitempty" validate:"omitempty,gte=0,lte=4"`
}

type PatchStudentRequest struct {
	Nama        *string  `json:"nama,omitempty" validate:"omitempty,min=2,max=100"`
	Prodi       *string  `json:"prodi,omitempty" validate:"omitempty,min=2,max=100"`
	Angkatan    *int     `json:"angkatan,omitempty"`
	IPKTerakhir *float64 `json:"ipk_terakhir,omitempty" validate:"omitempty,gte=0,lte=4"`
}