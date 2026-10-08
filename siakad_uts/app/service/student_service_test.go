package service

import (
	"testing"

	"api-students/app/model"
	"api-students/helper"
)

func TestCreateStudentValidation(t *testing.T) {
	req := model.CreateStudentRequest{
		NIM:         "",
		Nama:        "",
		Prodi:       "",
		Angkatan:    0,
		IPKTerakhir: nil,
	}

	errs := helper.ValidateStruct(req)

	if errs["nim"] == "" {
		t.Error("seharusnya NIM menghasilkan error")
	}

	if errs["nama"] == "" {
		t.Error("seharusnya nama menghasilkan error")
	}

	if errs["prodi"] == "" {
		t.Error("seharusnya prodi menghasilkan error")
	}
}

func TestUpdateStudentValidation(t *testing.T) {
	req := model.UpdateStudentRequest{
		Nama:        "",
		Prodi:       "",
		Angkatan:    0,
		IPKTerakhir: nil,
	}

	errs := helper.ValidateStruct(req)

	if errs["nama"] == "" {
		t.Error("seharusnya nama menghasilkan error")
	}

	if errs["prodi"] == "" {
		t.Error("seharusnya prodi menghasilkan error")
	}
}

func TestApplyPatch(t *testing.T) {
	ipk := 3.80
	angkatan := 2025
	nama := "Warda"
	prodi := "D4 Teknik Informatika"

	current := model.Student{
		ID:          1,
		NIM:         "434241100000",
		Nama:        "Sania",
		Prodi:       "D4 Teknik Informatika",
		Angkatan:    2024,
		IPKTerakhir: nil,
	}

	req := model.PatchStudentRequest{
		Nama:        &nama,
		Prodi:       &prodi,
		Angkatan:    &angkatan,
		IPKTerakhir: &ipk,
	}

	ApplyPatch(&current, req)

	if current.Nama != "Warda" {
		t.Errorf(
			"Nama tidak berubah, mendapatkan: %s",
			current.Nama,
		)
	}

	if current.Prodi != "D4 Teknik Informatika" {
		t.Errorf(
			"Prodi tidak berubah, mendapatkan: %s",
			current.Prodi,
		)
	}

	if current.Angkatan != 2025 {
		t.Errorf(
			"Angkatan tidak berubah, mendapatkan: %d",
			current.Angkatan,
		)
	}

	if current.IPKTerakhir == nil || *current.IPKTerakhir != 3.80 {
		t.Errorf(
			"IPKTerakhir tidak berubah dengan benar",
		)
	}

	// NIM tidak dikirim melalui PATCH,
	// sehingga nilainya harus tetap.
	if current.NIM != "434241100000" {
		t.Errorf(
			"NIM seharusnya tetap, mendapatkan: %s",
			current.NIM,
		)
	}
}