package service

import (
	"strings"

	"api-students/app/model"
)

func ApplyPatch(
	student *model.Student,
	req model.PatchStudentRequest,
) {
	if req.Nama != nil {
		student.Nama = strings.TrimSpace(*req.Nama)
	}

	if req.Prodi != nil {
		student.Prodi = strings.TrimSpace(*req.Prodi)
	}

	if req.Angkatan != nil {
		student.Angkatan = *req.Angkatan
	}

	if req.IPKTerakhir != nil {
		student.IPKTerakhir = req.IPKTerakhir
	}
}