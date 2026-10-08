package model

import "time"

type Student struct {
	ID        int       `json:"id"`
	NIM       string    `json:"nim"`
	Name      string    `json:"name"`
	Grade     float64   `json:"grade"`
	IsActive  bool      `json:"is_active"`
	OwnerID   int       `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateStudentRequest struct {
	NIM      string  `json:"nim" validate:"required,min=8,max=20"`
	Name     string  `json:"name" validate:"required,min=2,max=100"`
	Grade    float64 `json:"grade" validate:"gte=0,lte=100"`
	IsActive bool    `json:"is_active"`
}

type UpdateStudentRequest struct {
	NIM      string  `json:"nim" validate:"required,min=8,max=20"`
	Name     string  `json:"name" validate:"required,min=2,max=100"`
	Grade    float64 `json:"grade" validate:"gte=0,lte=100"`
	IsActive bool    `json:"is_active"`
}

type PatchStudentRequest struct {
	NIM      *string  `json:"nim,omitempty" validate:"omitnil,min=8,max=20"`
	Name     *string  `json:"name,omitempty" validate:"omitnil,min=2,max=100"`
	Grade    *float64 `json:"grade,omitempty" validate:"omitnil,gte=0,lte=100"`
	IsActive *bool    `json:"is_active,omitempty"`
}