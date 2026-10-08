package service

import (
	"strings"

	"api-students/app/model"
)

func ApplyPatch(
	student *model.Student,
	req model.PatchStudentRequest,
) {
	if req.NIM != nil {
		student.NIM = strings.TrimSpace(*req.NIM)
	}

	if req.Name != nil {
		student.Name = strings.TrimSpace(*req.Name)
	}

	if req.Grade != nil {
		student.Grade = *req.Grade
	}

	if req.IsActive != nil {
		student.IsActive = *req.IsActive
	}
}