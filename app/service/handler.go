package service

import (
	"errors"
	"strconv"
	"strings"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"

	"github.com/gofiber/fiber/v2"
)

// =====================================================
// Helper
// =====================================================

// parseStudentID mengambil ID dari URL parameter.
func parseStudentID(c *fiber.Ctx) (int, bool) {
	id, err := strconv.Atoi(c.Params("id"))

	if err != nil || id < 1 {
		return 0, false
	}

	return id, true
}

// =====================================================
// GET /api/v1/students
// =====================================================

func GetStudents(
	c *fiber.Ctx,
	repo repository.StudentRepository,
) error {

	page := 1
	limit := 10

	// Pagination: page
	if value := c.Query("page"); value != "" {
		parsed, err := strconv.Atoi(value)

		if err != nil || parsed < 1 {
			return helper.ErrorResponse(
				c,
				fiber.StatusBadRequest,
				"page harus berupa angka positif",
				nil,
			)
		}

		page = parsed
	}

	// Pagination: limit
	if value := c.Query("limit"); value != "" {
		parsed, err := strconv.Atoi(value)

		if err != nil || parsed < 1 {
			return helper.ErrorResponse(
				c,
				fiber.StatusBadRequest,
				"limit harus berupa angka positif",
				nil,
			)
		}

		if parsed > 50 {
			parsed = 50
		}

		limit = parsed
	}

	// Search berdasarkan nama.
	search := strings.TrimSpace(
		c.Query("search"),
	)

	// Filter is_active.
	var isActive *bool

	if value := c.Query("is_active"); value != "" {
		parsed, err := strconv.ParseBool(value)

		if err != nil {
			return helper.ErrorResponse(
				c,
				fiber.StatusBadRequest,
				"is_active harus bernilai true atau false",
				nil,
			)
		}

		isActive = &parsed
	}

	// Filter grade minimum.
	var gradeMin *float64

	if value := c.Query("grade_min"); value != "" {
		parsed, err := strconv.ParseFloat(value, 64)

		if err != nil {
			return helper.ErrorResponse(
				c,
				fiber.StatusBadRequest,
				"grade_min harus berupa angka",
				nil,
			)
		}

		gradeMin = &parsed
	}

	// Filter grade maksimum.
	var gradeMax *float64

	if value := c.Query("grade_max"); value != "" {
		parsed, err := strconv.ParseFloat(value, 64)

		if err != nil {
			return helper.ErrorResponse(
				c,
				fiber.StatusBadRequest,
				"grade_max harus berupa angka",
				nil,
			)
		}

		gradeMax = &parsed
	}

	// Sort dan order.
	sortField := strings.ToLower(
		strings.TrimSpace(c.Query("sort", "id")),
	)

	order := strings.ToLower(
		strings.TrimSpace(c.Query("order", "asc")),
	)

	// Whitelist field sorting.
	allowedSort := map[string]bool{
		"id":        true,
		"nim":       true,
		"name":      true,
		"grade":     true,
		"is_active": true,
	}

	if !allowedSort[sortField] {
		return helper.ErrorResponse(
			c,
			fiber.StatusBadRequest,
			"field sort tidak diperbolehkan",
			nil,
		)
	}

	if order != "asc" && order != "desc" {
		return helper.ErrorResponse(
			c,
			fiber.StatusBadRequest,
			"order harus asc atau desc",
			nil,
		)
	}

	// Menghitung offset untuk PostgreSQL.
	offset := (page - 1) * limit

	students, total, err := repo.FindAll(
		c.Context(),
		search,
		isActive,
		gradeMin,
		gradeMax,
		sortField,
		order,
		limit,
		offset,
	)

	if err != nil {
		return helper.ErrorResponse(
			c,
			fiber.StatusInternalServerError,
			"gagal mengambil data student",
			nil,
		)
	}

	totalPages := 0

	if total > 0 {
		totalPages = (total + limit - 1) / limit
	}

	return helper.SuccessResponse(
		c,
		fiber.StatusOK,
		"Data student berhasil diambil",
		students,
		map[string]interface{}{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
		},
	)
}

// =====================================================
// GET /api/v1/students/:id
// =====================================================

func GetStudent(
	c *fiber.Ctx,
	repo repository.StudentRepository,
) error {

	id, valid := parseStudentID(c)

	if !valid {
		return helper.ErrorResponse(
			c,
			fiber.StatusBadRequest,
			"id harus berupa angka positif",
			nil,
		)
	}

	student, err := repo.FindByID(
		c.Context(),
		id,
	)

	if errors.Is(err, repository.ErrNotFound) {
		return helper.ErrorResponse(
			c,
			fiber.StatusNotFound,
			"student tidak ditemukan",
			nil,
		)
	}

	if err != nil {
		return helper.ErrorResponse(
			c,
			fiber.StatusInternalServerError,
			"gagal mengambil data student",
			nil,
		)
	}

	return helper.SuccessResponse(
		c,
		fiber.StatusOK,
		"Data student berhasil ditemukan",
		student,
		nil,
	)
}

// =====================================================
// POST /api/v1/students
// =====================================================

func CreateStudent(
	c *fiber.Ctx,
	repo repository.StudentRepository,
) error {

	var req model.CreateStudentRequest

	if err := c.BodyParser(&req); err != nil {
		return helper.ErrorResponse(
			c,
			fiber.StatusBadRequest,
			"body harus berupa JSON yang valid",
			nil,
		)
	}

	req.NIM = strings.TrimSpace(req.NIM)
	req.Name = strings.TrimSpace(req.Name)

	// Validasi menggunakan pure business rules.
	validationErrors := ValidateCreateStudent(req)

	if len(validationErrors) > 0 {
		return helper.ErrorResponse(
			c,
			fiber.StatusUnprocessableEntity,
			"validasi gagal",
			validationErrors,
		)
	}

	student := model.Student{
		NIM:      req.NIM,
		Name:     req.Name,
		Grade:    req.Grade,
		IsActive: req.IsActive,
	}

	result, err := repo.Create(
		c.Context(),
		student,
	)

	if errors.Is(err, repository.ErrDuplicate) {
		return helper.ErrorResponse(
			c,
			fiber.StatusConflict,
			"NIM sudah digunakan",
			map[string]string{
				"nim": "NIM harus unik",
			},
		)
	}

	if err != nil {
		return helper.ErrorResponse(
			c,
			fiber.StatusInternalServerError,
			"gagal menambahkan student",
			nil,
		)
	}

	location := "/api/v1/students/" +
		strconv.Itoa(result.ID)

	c.Set("Location", location)

	return helper.SuccessResponse(
		c,
		fiber.StatusCreated,
		"Student berhasil dibuat",
		result,
		nil,
	)
}

// =====================================================
// PUT /api/v1/students/:id
// =====================================================

func ReplaceStudent(
	c *fiber.Ctx,
	repo repository.StudentRepository,
) error {

	id, valid := parseStudentID(c)

	if !valid {
		return helper.ErrorResponse(
			c,
			fiber.StatusBadRequest,
			"id harus berupa angka positif",
			nil,
		)
	}

	var req model.UpdateStudentRequest

	if err := c.BodyParser(&req); err != nil {
		return helper.ErrorResponse(
			c,
			fiber.StatusBadRequest,
			"body harus berupa JSON yang valid",
			nil,
		)
	}

	req.NIM = strings.TrimSpace(req.NIM)
	req.Name = strings.TrimSpace(req.Name)

	// Validasi PUT menggunakan pure business rules.
	validationErrors := ValidateUpdateStudent(req)

	if len(validationErrors) > 0 {
		return helper.ErrorResponse(
			c,
			fiber.StatusUnprocessableEntity,
			"validasi gagal",
			validationErrors,
		)
	}

	student := model.Student{
		ID:       id,
		NIM:      req.NIM,
		Name:     req.Name,
		Grade:    req.Grade,
		IsActive: req.IsActive,
	}

	result, err := repo.Update(
		c.Context(),
		id,
		student,
	)

	if errors.Is(err, repository.ErrNotFound) {
		return helper.ErrorResponse(
			c,
			fiber.StatusNotFound,
			"student tidak ditemukan",
			nil,
		)
	}

	if errors.Is(err, repository.ErrDuplicate) {
		return helper.ErrorResponse(
			c,
			fiber.StatusConflict,
			"NIM sudah digunakan",
			map[string]string{
				"nim": "NIM harus unik",
			},
		)
	}

	if err != nil {
		return helper.ErrorResponse(
			c,
			fiber.StatusInternalServerError,
			"gagal mengganti data student",
			nil,
		)
	}

	return helper.SuccessResponse(
		c,
		fiber.StatusOK,
		"Student berhasil diganti seluruhnya",
		result,
		nil,
	)
}

// =====================================================
// PATCH /api/v1/students/:id
// =====================================================

func PatchStudent(
	c *fiber.Ctx,
	repo repository.StudentRepository,
) error {

	id, valid := parseStudentID(c)

	if !valid {
		return helper.ErrorResponse(
			c,
			fiber.StatusBadRequest,
			"id harus berupa angka positif",
			nil,
		)
	}

	current, err := repo.FindByID(
		c.Context(),
		id,
	)

	if errors.Is(err, repository.ErrNotFound) {
		return helper.ErrorResponse(
			c,
			fiber.StatusNotFound,
			"student tidak ditemukan",
			nil,
		)
	}

	if err != nil {
		return helper.ErrorResponse(
			c,
			fiber.StatusInternalServerError,
			"gagal mengambil data student",
			nil,
		)
	}

	var req model.PatchStudentRequest

	if err := c.BodyParser(&req); err != nil {
		return helper.ErrorResponse(
			c,
			fiber.StatusBadRequest,
			"body harus berupa JSON yang valid",
			nil,
		)
	}

	// PATCH tidak boleh kosong.
	if req.NIM == nil &&
		req.Name == nil &&
		req.Grade == nil &&
		req.IsActive == nil {

		return helper.ErrorResponse(
			c,
			fiber.StatusBadRequest,
			"tidak ada field yang diubah",
			nil,
		)
	}

	// Salin data lama.
	updated := current

	// Terapkan perubahan menggunakan pure business rules.
	validationErrors := ApplyPatch(
		&updated,
		req,
	)

	if len(validationErrors) > 0 {
		return helper.ErrorResponse(
			c,
			fiber.StatusUnprocessableEntity,
			"validasi gagal",
			validationErrors,
		)
	}

	result, err := repo.Update(
		c.Context(),
		id,
		updated,
	)

	if errors.Is(err, repository.ErrNotFound) {
		return helper.ErrorResponse(
			c,
			fiber.StatusNotFound,
			"student tidak ditemukan",
			nil,
		)
	}

	if errors.Is(err, repository.ErrDuplicate) {
		return helper.ErrorResponse(
			c,
			fiber.StatusConflict,
			"NIM sudah digunakan",
			map[string]string{
				"nim": "NIM harus unik",
			},
		)
	}

	if err != nil {
		return helper.ErrorResponse(
			c,
			fiber.StatusInternalServerError,
			"gagal memperbarui student",
			nil,
		)
	}

	return helper.SuccessResponse(
		c,
		fiber.StatusOK,
		"Student berhasil diperbarui sebagian",
		result,
		nil,
	)
}

// =====================================================
// DELETE /api/v1/students/:id
// =====================================================

func DeleteStudent(
	c *fiber.Ctx,
	repo repository.StudentRepository,
) error {

	id, valid := parseStudentID(c)

	if !valid {
		return helper.ErrorResponse(
			c,
			fiber.StatusBadRequest,
			"id harus berupa angka positif",
			nil,
		)
	}

	err := repo.Delete(
		c.Context(),
		id,
	)

	if errors.Is(err, repository.ErrNotFound) {
		return helper.ErrorResponse(
			c,
			fiber.StatusNotFound,
			"student tidak ditemukan",
			nil,
		)
	}

	if err != nil {
		return helper.ErrorResponse(
			c,
			fiber.StatusInternalServerError,
			"gagal menghapus student",
			nil,
		)
	}

	return c.SendStatus(
		fiber.StatusNoContent,
	)
}
