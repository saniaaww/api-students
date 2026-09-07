package service

import (
	"strings"

	"api-students/app/model"
)

func ValidateCreateStudent(req model.CreateStudentRequest) map[string]string {
	errors := map[string]string{}

	req.NIM = strings.TrimSpace(req.NIM)
	req.Name = strings.TrimSpace(req.Name)

	if req.NIM == "" {
		errors["nim"] = "NIM wajib diisi"
	}

	if req.Name == "" {
		errors["name"] = "nama wajib diisi"
	}

	if req.Grade < 0 || req.Grade > 100 {
		errors["grade"] = "grade harus berada di antara 0 dan 100"
	}

	return errors
}

func ValidateUpdateStudent(req model.UpdateStudentRequest) map[string]string {
	errors := map[string]string{}

	req.NIM = strings.TrimSpace(req.NIM)
	req.Name = strings.TrimSpace(req.Name)

	if req.NIM == "" {
		errors["nim"] = "NIM wajib diisi pada PUT"
	}

	if req.Name == "" {
		errors["name"] = "nama wajib diisi pada PUT"
	}

	if req.Grade < 0 || req.Grade > 100 {
		errors["grade"] = "grade harus berada di antara 0 dan 100"
	}

	return errors
}

func ApplyPatch(
	student *model.Student,
	req model.PatchStudentRequest,
) map[string]string {

	errors := map[string]string{}

	if req.NIM != nil {
		nim := strings.TrimSpace(*req.NIM)

		if nim == "" {
			errors["nim"] = "NIM tidak boleh kosong"
		} else {
			student.NIM = nim
		}
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)

		if name == "" {
			errors["name"] = "nama tidak boleh kosong"
		} else {
			student.Name = name
		}
	}

	if req.Grade != nil {
		if *req.Grade < 0 || *req.Grade > 100 {
			errors["grade"] = "grade harus berada di antara 0 dan 100"
		} else {
			student.Grade = *req.Grade
		}
	}

	if req.IsActive != nil {
		student.IsActive = *req.IsActive
	}

	return errors
}
