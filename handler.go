package main

import (
	"sort"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

var students = []Student{
	{
		ID:       1,
		NIM:      "434241101",
		Name:     "Sania",
		Grade:    91.5,
		IsActive: true,
	},
	{
		ID:       2,
		NIM:      "434241105",
		Name:     "Alda",
		Grade:    90,
		IsActive: true,
	},
	{
		ID:       3,
		NIM:      "434241100",
		Name:     "Warda",
		Grade:    78,
		IsActive: false,
	},
}

var nextID = 4

// findStudentIndex mencari posisi student berdasarkan ID.
func findStudentIndex(id int) int {
	for i := range students {
		if students[i].ID == id {
			return i
		}
	}

	return -1
}

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

func getStudents(c *fiber.Ctx) error {
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

		// Batas maksimal 50 data per halaman.
		if parsed > 50 {
			parsed = 50
		}

		limit = parsed
	}

	// Search berdasarkan nama.
	search := strings.ToLower(
		strings.TrimSpace(c.Query("search")),
	)

	// Filter berdasarkan status aktif.
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

	// Sort dan order.
	sortField := strings.ToLower(
		strings.TrimSpace(c.Query("sort", "id")),
	)

	order := strings.ToLower(
		strings.TrimSpace(c.Query("order", "asc")),
	)

	// Whitelist field yang boleh digunakan untuk sorting.
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

	// Filtering data.
	filtered := make([]Student, 0)

	for _, student := range students {

		// Filter search berdasarkan nama.
		if search != "" &&
			!strings.Contains(
				strings.ToLower(student.Name),
				search,
			) {
			continue
		}

		// Filter is_active.
		if isActive != nil &&
			student.IsActive != *isActive {
			continue
		}

		filtered = append(filtered, student)
	}

	// Sorting.
	sort.SliceStable(filtered, func(i, j int) bool {
		var less bool

		switch sortField {
		case "id":
			less = filtered[i].ID < filtered[j].ID

		case "nim":
			less = filtered[i].NIM < filtered[j].NIM

		case "name":
			less = strings.ToLower(filtered[i].Name) <
				strings.ToLower(filtered[j].Name)

		case "grade":
			less = filtered[i].Grade < filtered[j].Grade

		case "is_active":
			less = !filtered[i].IsActive &&
				filtered[j].IsActive
		}

		if order == "desc" {
			return !less
		}

		return less
	})

	// Pagination.
	total := len(filtered)
	totalPages := 0

	if total > 0 {
		totalPages = (total + limit - 1) / limit
	}

	start := (page - 1) * limit

	// Jika halaman melebihi jumlah halaman.
	if start >= total {
		return successResponse(
			c,
			fiber.StatusOK,
			"Data student berhasil diambil",
			[]Student{},
			map[string]interface{}{
				"page":        page,
				"limit":       limit,
				"total":       total,
				"total_pages": totalPages,
			},
		)
	}

	end := start + limit

	if end > total {
		end = total
	}

	result := filtered[start:end]

	return successResponse(
		c,
		fiber.StatusOK,
		"Data student berhasil diambil",
		result,
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

func getStudent(c *fiber.Ctx) error {
	id, valid := parseStudentID(c)

	if !valid {
		return errorResponse(
			c,
			fiber.StatusBadRequest,
			"id harus berupa angka positif",
			nil,
		)
	}

	index := findStudentIndex(id)

	if index == -1 {
		return errorResponse(
			c,
			fiber.StatusNotFound,
			"student tidak ditemukan",
			nil,
		)
	}

	return successResponse(
		c,
		fiber.StatusOK,
		"Data student berhasil ditemukan",
		students[index],
		nil,
	)
}

// =====================================================
// POST /api/v1/students
// =====================================================

func createStudent(c *fiber.Ctx) error {
	var req CreateStudentRequest

	if err := c.BodyParser(&req); err != nil {
		return errorResponse(
			c,
			fiber.StatusBadRequest,
			"body harus berupa JSON yang valid",
			nil,
		)
	}

	// Membersihkan input.
	req.NIM = strings.TrimSpace(req.NIM)
	req.Name = strings.TrimSpace(req.Name)

	errors := map[string]string{}

	// Validasi NIM.
	if req.NIM == "" {
		errors["nim"] = "NIM wajib diisi"
	}

	// Validasi nama.
	if req.Name == "" {
		errors["name"] = "nama wajib diisi"
	}

	// Validasi grade.
	if req.Grade < 0 || req.Grade > 100 {
		errors["grade"] = "grade harus berada di antara 0 dan 100"
	}

	// Cek NIM duplikat.
	for _, student := range students {
		if student.NIM == req.NIM {
			return errorResponse(
				c,
				fiber.StatusConflict,
				"NIM sudah digunakan",
				map[string]string{
					"nim": "NIM harus unik",
				},
			)
		}
	}

	// Jika validasi gagal.
	if len(errors) > 0 {
		return errorResponse(
			c,
			fiber.StatusUnprocessableEntity,
			"validasi gagal",
			errors,
		)
	}

	// Membuat student baru.
	newStudent := Student{
		ID:       nextID,
		NIM:      req.NIM,
		Name:     req.Name,
		Grade:    req.Grade,
		IsActive: req.IsActive,
	}

	students = append(students, newStudent)

	nextID++

	// Header Location.
	location := "/api/v1/students/" +
		strconv.Itoa(newStudent.ID)

	c.Set("Location", location)

	return successResponse(
		c,
		fiber.StatusCreated,
		"Student berhasil dibuat",
		newStudent,
		nil,
	)
}

// =====================================================
// PUT /api/v1/students/:id
// =====================================================

func replaceStudent(c *fiber.Ctx) error {
	id, valid := parseStudentID(c)

	if !valid {
		return errorResponse(
			c,
			fiber.StatusBadRequest,
			"id harus berupa angka positif",
			nil,
		)
	}

	index := findStudentIndex(id)

	if index == -1 {
		return errorResponse(
			c,
			fiber.StatusNotFound,
			"student tidak ditemukan",
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

	errors := map[string]string{}

	// Semua field wajib untuk PUT.
	if req.NIM == "" {
		errors["nim"] = "NIM wajib diisi pada PUT"
	}

	if req.Name == "" {
		errors["name"] = "nama wajib diisi pada PUT"
	}

	if req.Grade < 0 || req.Grade > 100 {
		errors["grade"] = "grade harus berada di antara 0 dan 100"
	}

	// Cek NIM duplikat dengan student lain.
	for i, student := range students {
		if i != index && student.NIM == req.NIM {
			return errorResponse(
				c,
				fiber.StatusConflict,
				"NIM sudah digunakan",
				map[string]string{
					"nim": "NIM harus unik",
				},
			)
		}
	}

	if len(errors) > 0 {
		return errorResponse(
			c,
			fiber.StatusUnprocessableEntity,
			"validasi gagal",
			errors,
		)
	}

	// PUT mengganti seluruh isi.
	students[index].NIM = req.NIM
	students[index].Name = req.Name
	students[index].Grade = req.Grade
	students[index].IsActive = req.IsActive

	return successResponse(
		c,
		fiber.StatusOK,
		"Student berhasil diganti seluruhnya",
		students[index],
		nil,
	)
}

// =====================================================
// PATCH /api/v1/students/:id
// =====================================================

func patchStudent(c *fiber.Ctx) error {
	id, valid := parseStudentID(c)

	if !valid {
		return errorResponse(
			c,
			fiber.StatusBadRequest,
			"id harus berupa angka positif",
			nil,
		)
	}

	index := findStudentIndex(id)

	if index == -1 {
		return errorResponse(
			c,
			fiber.StatusNotFound,
			"student tidak ditemukan",
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

	// Tidak boleh PATCH tanpa field.
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

	// PATCH NIM.
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

		// Cek NIM duplikat.
		for i, student := range students {
			if i != index && student.NIM == nim {
				return errorResponse(
					c,
					fiber.StatusConflict,
					"NIM sudah digunakan",
					map[string]string{
						"nim": "NIM harus unik",
					},
				)
			}
		}

		students[index].NIM = nim
	}

	// PATCH Name.
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

		students[index].Name = name
	}

	// PATCH Grade.
	if req.Grade != nil {
		if *req.Grade < 0 || *req.Grade > 100 {
			return errorResponse(
				c,
				fiber.StatusUnprocessableEntity,
				"validasi gagal",
				map[string]string{
					"grade": "grade harus berada di antara 0 dan 100",
				},
			)
		}

		students[index].Grade = *req.Grade
	}

	// PATCH IsActive.
	if req.IsActive != nil {
		students[index].IsActive = *req.IsActive
	}

	return successResponse(
		c,
		fiber.StatusOK,
		"Student berhasil diperbarui sebagian",
		students[index],
		nil,
	)
}

// =====================================================
// DELETE /api/v1/students/:id
// =====================================================

func deleteStudent(c *fiber.Ctx) error {
	id, valid := parseStudentID(c)

	if !valid {
		return errorResponse(
			c,
			fiber.StatusBadRequest,
			"id harus berupa angka positif",
			nil,
		)
	}

	index := findStudentIndex(id)

	if index == -1 {
		return errorResponse(
			c,
			fiber.StatusNotFound,
			"student tidak ditemukan",
			nil,
		)
	}

	// Menghapus student dari slice.
	students = append(
		students[:index],
		students[index+1:]...,
	)

	// DELETE berhasil tanpa response body.
	return c.SendStatus(fiber.StatusNoContent)
}
