package main

import (
	"errors"
	"strconv"
	"strings"

	"api-students/app/repository"
	"api-students/model"

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

func getStudents(
	c *fiber.Ctx,
	repo repository.StudentRepository,
) error {

	page := 1
	limit := 10

	// Pagination: page
	if value := c.Query("page"); value != "" {
		parsed, err := strconv.Atoi(value)

		if err != nil || parsed < 1 {
			return errorResponse(
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
			return errorResponse(
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
			return errorResponse(
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
			return errorResponse(
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
			return errorResponse(
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
		return errorResponse(
			c,
			fiber.StatusBadRequest,
			"field sort tidak diperbolehkan",
			nil,
		)
	}

	if order != "asc" && order != "desc" {
		return errorResponse(
			c,
			fiber.StatusBadRequest,
			"order harus asc atau desc",
			nil,
		)
	}

	// Menghitung offset untuk PostgreSQL.
	offset := (page - 1) * limit

	// Seluruh filtering, sorting, dan pagination
	// sekarang dikerjakan oleh repository/SQL.
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
		return errorResponse(
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

	return successResponse(
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

func getStudent(
	c *fiber.Ctx,
	repo repository.StudentRepository,
) error {

	id, valid := parseStudentID(c)

	if !valid {
		return errorResponse(
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
		return errorResponse(
			c,
			fiber.StatusNotFound,
			"student tidak ditemukan",
			nil,
		)
	}

	if err != nil {
		return errorResponse(
			c,
			fiber.StatusInternalServerError,
			"gagal mengambil data student",
			nil,
		)
	}

	return successResponse(
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

func createStudent(
	c *fiber.Ctx,
	repo repository.StudentRepository,
) error {

	var req CreateStudentRequest

	if err := c.BodyParser(&req); err != nil {
		return errorResponse(
			c,
			fiber.StatusBadRequest,
			"body harus berupa JSON yang valid",
			nil,
		)
	}

	req.NIM = strings.TrimSpace(req.NIM)
	req.Name = strings.TrimSpace(req.Name)

	validationErrors := map[string]string{}

	if req.NIM == "" {
		validationErrors["nim"] = "NIM wajib diisi"
	}

	if req.Name == "" {
		validationErrors["name"] = "nama wajib diisi"
	}

	if req.Grade < 0 || req.Grade > 100 {
		validationErrors["grade"] =
			"grade harus berada di antara 0 dan 100"
	}

	if len(validationErrors) > 0 {
		return errorResponse(
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
		return errorResponse(
			c,
			fiber.StatusConflict,
			"NIM sudah digunakan",
			map[string]string{
				"nim": "NIM harus unik",
			},
		)
	}

	if err != nil {
		return errorResponse(
			c,
			fiber.StatusInternalServerError,
			"gagal menambahkan student",
			nil,
		)
	}

	location := "/api/v1/students/" +
		strconv.Itoa(result.ID)

	c.Set("Location", location)

	return successResponse(
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

func replaceStudent(
	c *fiber.Ctx,
	repo repository.StudentRepository,
) error {

	id, valid := parseStudentID(c)

	if !valid {
		return errorResponse(
			c,
			fiber.StatusBadRequest,
			"id harus berupa angka positif",
			nil,
		)
	}

	var req ReplaceStudentRequest

	if err := c.BodyParser(&req); err != nil {
		return errorResponse(
			c,
			fiber.StatusBadRequest,
			"body harus berupa JSON yang valid",
			nil,
		)
	}

	req.NIM = strings.TrimSpace(req.NIM)
	req.Name = strings.TrimSpace(req.Name)

	validationErrors := map[string]string{}

	// Seluruh field wajib dikirim pada PUT.
	if req.NIM == "" {
		validationErrors["nim"] =
			"NIM wajib diisi pada PUT"
	}

	if req.Name == "" {
		validationErrors["name"] =
			"nama wajib diisi pada PUT"
	}

	if req.Grade < 0 || req.Grade > 100 {
		validationErrors["grade"] =
			"grade harus berada di antara 0 dan 100"
	}

	if len(validationErrors) > 0 {
		return errorResponse(
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
		return errorResponse(
			c,
			fiber.StatusNotFound,
			"student tidak ditemukan",
			nil,
		)
	}

	if errors.Is(err, repository.ErrDuplicate) {
		return errorResponse(
			c,
			fiber.StatusConflict,
			"NIM sudah digunakan",
			map[string]string{
				"nim": "NIM harus unik",
			},
		)
	}

	if err != nil {
		return errorResponse(
			c,
			fiber.StatusInternalServerError,
			"gagal mengganti data student",
			nil,
		)
	}

	return successResponse(
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

func patchStudent(
	c *fiber.Ctx,
	repo repository.StudentRepository,
) error {

	id, valid := parseStudentID(c)

	if !valid {
		return errorResponse(
			c,
			fiber.StatusBadRequest,
			"id harus berupa angka positif",
			nil,
		)
	}

	// Ambil data lama terlebih dahulu.
	current, err := repo.FindByID(
		c.Context(),
		id,
	)

	if errors.Is(err, repository.ErrNotFound) {
		return errorResponse(
			c,
			fiber.StatusNotFound,
			"student tidak ditemukan",
			nil,
		)
	}

	if err != nil {
		return errorResponse(
			c,
			fiber.StatusInternalServerError,
			"gagal mengambil data student",
			nil,
		)
	}

	var req PatchStudentRequest

	if err := c.BodyParser(&req); err != nil {
		return errorResponse(
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

		return errorResponse(
			c,
			fiber.StatusBadRequest,
			"tidak ada field yang diubah",
			nil,
		)
	}

	// Pertahankan nilai lama.
	updated := current

	// Update NIM jika dikirim.
	if req.NIM != nil {

		nim := strings.TrimSpace(*req.NIM)

		if nim == "" {
			return errorResponse(
				c,
				fiber.StatusUnprocessableEntity,
				"validasi gagal",
				map[string]string{
					"nim": "NIM tidak boleh kosong",
				},
			)
		}

		updated.NIM = nim
	}

	// Update Name jika dikirim.
	if req.Name != nil {

		name := strings.TrimSpace(*req.Name)

		if name == "" {
			return errorResponse(
				c,
				fiber.StatusUnprocessableEntity,
				"validasi gagal",
				map[string]string{
					"name": "nama tidak boleh kosong",
				},
			)
		}

		updated.Name = name
	}

	// Update Grade jika dikirim.
	if req.Grade != nil {

		if *req.Grade < 0 || *req.Grade > 100 {
			return errorResponse(
				c,
				fiber.StatusUnprocessableEntity,
				"validasi gagal",
				map[string]string{
					"grade":
						"grade harus berada di antara 0 dan 100",
				},
			)
		}

		updated.Grade = *req.Grade
	}

	// Update IsActive jika dikirim.
	if req.IsActive != nil {
		updated.IsActive = *req.IsActive
	}

	result, err := repo.Update(
		c.Context(),
		id,
		updated,
	)

	if errors.Is(err, repository.ErrNotFound) {
		return errorResponse(
			c,
			fiber.StatusNotFound,
			"student tidak ditemukan",
			nil,
		)
	}

	if errors.Is(err, repository.ErrDuplicate) {
		return errorResponse(
			c,
			fiber.StatusConflict,
			"NIM sudah digunakan",
			map[string]string{
				"nim": "NIM harus unik",
			},
		)
	}

	if err != nil {
		return errorResponse(
			c,
			fiber.StatusInternalServerError,
			"gagal memperbarui student",
			nil,
		)
	}

	return successResponse(
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

func deleteStudent(
	c *fiber.Ctx,
	repo repository.StudentRepository,
) error {

	id, valid := parseStudentID(c)

	if !valid {
		return errorResponse(
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
		return errorResponse(
			c,
			fiber.StatusNotFound,
			"student tidak ditemukan",
			nil,
		)
	}

	if err != nil {
		return errorResponse(
			c,
			fiber.StatusInternalServerError,
			"gagal menghapus student",
			nil,
		)
	}

	// DELETE berhasil tanpa body.
	return c.SendStatus(
		fiber.StatusNoContent,
	)
}