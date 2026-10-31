package service

import (
	"bytes"
	"encoding/csv"
	"errors"
	"strconv"
	"strings"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"

	"github.com/gofiber/fiber/v2"
)

// =====================================================
// HELPER
// =====================================================

// parseStudentID mengambil ID dari URL parameter.
func parseStudentID(c *fiber.Ctx) (int, bool) {
	id, err := strconv.Atoi(c.Params("id"))

	if err != nil || id < 1 {
		return 0, false
	}

	return id, true
}

// getCurrentAuthUser mengambil identitas user yang sedang login.
func getCurrentAuthUser(c *fiber.Ctx) (model.AuthUser, bool) {
	userID, ok := c.Locals("user_id").(int64)
	if !ok {
		return model.AuthUser{}, false
	}

	role, ok := c.Locals("role").(string)
	if !ok || role == "" {
		return model.AuthUser{}, false
	}

	username, _ := c.Locals("username").(string)

	return model.AuthUser{
		UserID:   userID,
		Username: username,
		Role:     role,
	}, true
}

// =====================================================
// CONTENT NEGOTIATION
// =====================================================

// wantsCSV menentukan format response berdasarkan header Accept.
//
// application/json -> JSON
// text/csv         -> CSV
// */*              -> JSON
// kosong           -> JSON
// format lain      -> 406 Not Acceptable
func wantsCSV(c *fiber.Ctx) (bool, error) {

	accept := strings.ToLower(
		strings.TrimSpace(c.Get("Accept")),
	)

	// Jika Accept kosong -> JSON
	if accept == "" {
		return false, nil
	}

	// CSV memiliki prioritas jika diminta secara eksplisit.
	if strings.Contains(accept, "text/csv") {
		return true, nil
	}

	// JSON
	if strings.Contains(accept, "application/json") {
		return false, nil
	}

	// Wildcard -> JSON
	if strings.Contains(accept, "*/*") {
		return false, nil
	}

	// Format tidak didukung.
	return false, helper.NotAcceptable(
		"format response tidak didukung",
	)
}

// =====================================================
// CSV RESPONSE
// =====================================================

func studentsToCSV(students []model.Student) ([]byte, error) {

	var buffer bytes.Buffer

	writer := csv.NewWriter(&buffer)

	// Header CSV
	if err := writer.Write([]string{
		"id",
		"nim",
		"name",
		"grade",
		"is_active",
		"owner_id",
		"created_at",
	}); err != nil {
		return nil, err
	}

	// Data student
	for _, student := range students {

		err := writer.Write([]string{
			strconv.Itoa(student.ID),
			student.NIM,
			student.Name,
			strconv.FormatFloat(student.Grade, 'f', 2, 64),
			strconv.FormatBool(student.IsActive),
			strconv.Itoa(student.OwnerID),
			student.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})

		if err != nil {
			return nil, err
		}
	}

	writer.Flush()

	if err := writer.Error(); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// =====================================================
// GET /api/v1/students
// =====================================================

func GetStudents(
	c *fiber.Ctx,
	repo repository.StudentRepository,
) error {

	// =====================================================
	// CONTENT NEGOTIATION
	// =====================================================

	useCSV, err := wantsCSV(c)

	if err != nil {
		return err
	}

	// =====================================================
	// LIMIT
	// =====================================================

	limit := 10

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

	// =====================================================
	// SEARCH
	// =====================================================

	search := strings.TrimSpace(
		c.Query("search"),
	)

	// =====================================================
	// FILTER IS_ACTIVE
	// =====================================================

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

	// =====================================================
	// FILTER GRADE MIN
	// =====================================================

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

	// =====================================================
	// FILTER GRADE MAX
	// =====================================================

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

	// =====================================================
	// CURSOR PAGINATION
	// =====================================================

	if cursorValue := c.Query("cursor"); cursorValue != "" {

		cursor, err := strconv.Atoi(cursorValue)

		if err != nil || cursor < 0 {
			return helper.ErrorResponse(
				c,
				fiber.StatusBadRequest,
				"cursor harus berupa angka nol atau positif",
				nil,
			)
		}

		students, err := repo.FindAllCursor(
			c.Context(),
			search,
			isActive,
			gradeMin,
			gradeMax,
			limit,
			cursor,
		)

		if err != nil {
			return helper.ErrorResponse(
				c,
				fiber.StatusInternalServerError,
				"gagal mengambil data student",
				nil,
			)
		}

		// =================================================
		// CSV
		// =================================================

		if useCSV {

			csvData, err := studentsToCSV(students)

			if err != nil {
				return helper.ErrorResponse(
					c,
					fiber.StatusInternalServerError,
					"gagal membuat response CSV",
					nil,
				)
			}

			c.Set("Content-Type", "text/csv; charset=utf-8")
			c.Set(
				"Content-Disposition",
				`attachment; filename="students.csv"`,
			)

			return c.Send(csvData)
		}

		// =================================================
		// JSON
		// =================================================

		nextCursor := 0

		if len(students) > 0 {
			nextCursor = students[len(students)-1].ID
		}

		hasNext := len(students) == limit

		return helper.SuccessResponse(
			c,
			fiber.StatusOK,
			"Data student berhasil diambil",
			students,
			map[string]interface{}{
				"limit":       limit,
				"cursor":      cursor,
				"next_cursor": nextCursor,
				"has_next":    hasNext,
			},
		)
	}

	// =====================================================
	// PAGINATION OFFSET
	// =====================================================

	page := 1

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

	// =====================================================
	// SORT & ORDER
	// =====================================================

	sortField := strings.ToLower(
		strings.TrimSpace(
			c.Query("sort", "id"),
		),
	)

	order := strings.ToLower(
		strings.TrimSpace(
			c.Query("order", "asc"),
		),
	)

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

	// =====================================================
	// OFFSET
	// =====================================================

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

	// =====================================================
	// CSV
	// =====================================================

	if useCSV {

		csvData, err := studentsToCSV(students)

		if err != nil {
			return helper.ErrorResponse(
				c,
				fiber.StatusInternalServerError,
				"gagal membuat response CSV",
				nil,
			)
		}

		c.Set("Content-Type", "text/csv; charset=utf-8")
		c.Set(
			"Content-Disposition",
			`attachment; filename="students.csv"`,
		)

		return c.Send(csvData)
	}

	// =====================================================
	// JSON
	// =====================================================

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
// GET /api/v1/students/cursor
// =====================================================

func GetStudentsCursor(
	c *fiber.Ctx,
	repo repository.StudentRepository,
) error {

	// =====================================================
	// CONTENT NEGOTIATION
	// =====================================================

	useCSV, err := wantsCSV(c)

	if err != nil {
		return err
	}

	// =====================================================
	// LIMIT
	// =====================================================

	limit := 10

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

	// =====================================================
	// CURSOR
	// =====================================================

	cursor := 0

	if value := c.Query("cursor"); value != "" {

		parsed, err := strconv.Atoi(value)

		if err != nil || parsed < 0 {
			return helper.ErrorResponse(
				c,
				fiber.StatusBadRequest,
				"cursor harus berupa angka 0 atau lebih",
				nil,
			)
		}

		cursor = parsed
	}

	// =====================================================
	// SEARCH
	// =====================================================

	search := strings.TrimSpace(
		c.Query("search"),
	)

	// =====================================================
	// FILTER IS_ACTIVE
	// =====================================================

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

	// =====================================================
	// FILTER GRADE MIN
	// =====================================================

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

	// =====================================================
	// FILTER GRADE MAX
	// =====================================================

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

	// =====================================================
	// QUERY DATABASE
	// =====================================================

	students, err := repo.FindAllCursor(
		c.Context(),
		search,
		isActive,
		gradeMin,
		gradeMax,
		limit,
		cursor,
	)

	if err != nil {
		return helper.ErrorResponse(
			c,
			fiber.StatusInternalServerError,
			"gagal mengambil data student",
			nil,
		)
	}

	// =====================================================
	// CSV
	// =====================================================

	if useCSV {

		csvData, err := studentsToCSV(students)

		if err != nil {
			return helper.ErrorResponse(
				c,
				fiber.StatusInternalServerError,
				"gagal membuat response CSV",
				nil,
			)
		}

		c.Set("Content-Type", "text/csv; charset=utf-8")
		c.Set(
			"Content-Disposition",
			`attachment; filename="students.csv"`,
		)

		return c.Send(csvData)
	}

	// =====================================================
	// JSON
	// =====================================================

	nextCursor := 0

	if len(students) > 0 {
		nextCursor = students[len(students)-1].ID
	}

	hasNext := len(students) == limit

	return helper.SuccessResponse(
		c,
		fiber.StatusOK,
		"Data student berhasil diambil",
		students,
		map[string]interface{}{
			"limit":       limit,
			"cursor":      cursor,
			"next_cursor": nextCursor,
			"has_next":    hasNext,
		},
	)
}

// =====================================================
// GET /api/v1/students/:id
// =====================================================

func GetStudent(
	c *fiber.Ctx,
	repo repository.StudentRepository,
	perms *helper.PermissionSet,
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

	currentUser, ok := getCurrentAuthUser(c)

	if !ok {
		return helper.ErrorResponse(
			c,
			fiber.StatusUnauthorized,
			"identitas user tidak valid",
			nil,
		)
	}

	if !CanAccessStudent(
		currentUser,
		student.OwnerID,
		perms,
		"student:read:any",
	) {
		return helper.ErrorResponse(
			c,
			fiber.StatusForbidden,
			"forbidden",
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

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	currentUser, ok := getCurrentAuthUser(c)

	if !ok {
		return helper.ErrorResponse(
			c,
			fiber.StatusUnauthorized,
			"identitas user tidak valid",
			nil,
		)
	}

	student := model.Student{
		NIM:      req.NIM,
		Name:     req.Name,
		Grade:    req.Grade,
		IsActive: req.IsActive,
		OwnerID:  int(currentUser.UserID),
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
	perms *helper.PermissionSet,
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

	currentUser, ok := getCurrentAuthUser(c)

	if !ok {
		return helper.ErrorResponse(
			c,
			fiber.StatusUnauthorized,
			"identitas user tidak valid",
			nil,
		)
	}

	if !CanAccessStudent(
		currentUser,
		current.OwnerID,
		perms,
		"student:update:any",
	) {
		return helper.ErrorResponse(
			c,
			fiber.StatusForbidden,
			"forbidden",
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

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	student := model.Student{
		ID:        id,
		NIM:       req.NIM,
		Name:      req.Name,
		Grade:     req.Grade,
		IsActive:  req.IsActive,
		OwnerID:   current.OwnerID,
		CreatedAt: current.CreatedAt,
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
	perms *helper.PermissionSet,
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

	currentUser, ok := getCurrentAuthUser(c)

	if !ok {
		return helper.ErrorResponse(
			c,
			fiber.StatusUnauthorized,
			"identitas user tidak valid",
			nil,
		)
	}

	if !CanAccessStudent(
		currentUser,
		current.OwnerID,
		perms,
		"student:update:any",
	) {
		return helper.ErrorResponse(
			c,
			fiber.StatusForbidden,
			"forbidden",
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

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	updated := current

	ApplyPatch(
		&updated,
		req,
	)

	updated.OwnerID = current.OwnerID

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

func GetCourses(
	c *fiber.Ctx,
	courseService *CourseService,
) error {
	courses, err := courseService.Courses(c.Context())
	if err != nil {
		return helper.ErrorResponse(
			c,
			fiber.StatusInternalServerError,
			"gagal mengambil data kelas",
			nil,
		)
	}

	return helper.SuccessResponse(
		c,
		fiber.StatusOK,
		"Data kelas berhasil diambil",
		courses,
		nil,
	)
}
