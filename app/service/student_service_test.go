package service

import (
	"testing"

	"api-students/app/model"
)
func TestValidateCreateStudent(t *testing.T) {
	req := model.CreateStudentRequest{
		NIM:      "",
		Name:     "",
		Grade:    101,
		IsActive: true,
	}
	errors := ValidateCreateStudent(req)

	if errors["nim"] == "" {
		t.Error("seharusnya NIM menghasilkan error")
	}

	if errors["name"] == "" {
		t.Error("seharusnya name menghasilkan error")
	}

	if errors["grade"] == "" {
		t.Error("seharusnya grade menghasilkan error")
	}
}
func TestValidateUpdateStudent(t *testing.T) {
	req := model.UpdateStudentRequest{
		NIM:      "",
		Name:     "",
		Grade:    101,
		IsActive: true,
	}

	errors := ValidateUpdateStudent(req)

	if errors["nim"] == "" {
		t.Error("seharusnya NIM menghasilkan error")
	}

	if errors["name"] == "" {
		t.Error("seharusnya name menghasilkan error")
	}

	if errors["grade"] == "" {
		t.Error("seharusnya grade menghasilkan error")
	}
}

func TestApplyPatch(t *testing.T) {
	current := model.Student{
		ID:       1,
		NIM:      "434241101",
		Name:     "Sania",
		Grade:    91.5,
		IsActive: true,
	}

	newName := "Warda"
	newGrade := 95.0

	req := model.PatchStudentRequest{
		Name:  &newName,
		Grade: &newGrade,
	}

	errors := ApplyPatch(&current, req)

	if len(errors) != 0 {
		t.Errorf("seharusnya tidak ada error, tetapi mendapat: %v", errors)
	}

	if current.Name != "Warda" {
		t.Errorf("Name tidak berubah, mendapatkan: %s", current.Name)
	}

	if current.Grade != 95.0 {
		t.Errorf("Grade tidak berubah, mendapatkan: %f", current.Grade)
	}

	// Field yang tidak diubah harus tetap sama.
	if current.NIM != "434241101" {
		t.Errorf("NIM seharusnya tetap, mendapatkan: %s", current.NIM)
	}

	if current.IsActive != true {
		t.Errorf("IsActive seharusnya tetap true")
	}
}
