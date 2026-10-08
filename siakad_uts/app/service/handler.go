package service

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

// ============================================================
// DEPENDENCIES
// ============================================================

type Handler struct {
	Students    repository.StudentRepository
	Users       repository.UserRepository
	Courses     repository.CourseRepository
	Enrollments repository.EnrollmentRepository
}

// ============================================================
// CREATE STUDENT
// POST /api/v1/students
// ============================================================

func (h *Handler) CreateStudent(c *fiber.Ctx) error {

	// Hanya admin yang boleh menambahkan mahasiswa
	role, ok := c.Locals("role").(string)
	if !ok || role != "admin" {
		return helper.Forbidden("akses hanya diperbolehkan untuk admin")
	}

	var req model.CreateStudentRequest

	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("format JSON tidak valid")
	}

	// Validasi field
	if fields := helper.ValidateStruct(req); len(fields) > 0 {
		return helper.Validation(fields)
	}

	// Validasi angkatan
	currentYear := time.Now().Year()

	if req.Angkatan < 1900 || req.Angkatan > currentYear {
		return helper.Validation(map[string]string{
			"angkatan": "angkatan harus 4 digit dan tidak boleh melebihi tahun berjalan",
		})
	}

	// Password awal mahasiswa = NIM
	// Password disimpan dalam bentuk hash
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(req.NIM),
		bcrypt.DefaultCost,
	)

	if err != nil {
		return helper.Internal(err)
	}

	user := model.User{
		Email:    strings.TrimSpace(req.Email),
		Password: string(hashedPassword),
		Role:     "mahasiswa",
	}

	student := model.Student{
		NIM:         strings.TrimSpace(req.NIM),
		Nama:        strings.TrimSpace(req.Nama),
		Prodi:       strings.TrimSpace(req.Prodi),
		Angkatan:    req.Angkatan,
		IPKTerakhir: req.IPKTerakhir,
	}

	// Buat user + student dalam satu transaction
	created, err := h.Students.CreateWithUser(
		c.Context(),
		user,
		student,
	)

	if err != nil {

		if errors.Is(err, repository.ErrDuplicate) {
			return helper.Validation(map[string]string{
				"nim": "NIM sudah terdaftar",
			})
		}

		if errors.Is(err, repository.ErrDuplicateEmail) {
			return helper.Validation(map[string]string{
				"email": "email sudah terdaftar",
			})
		}

		return helper.Internal(err)
	}

	return helper.SuccessResponse(
		c,
		fiber.StatusCreated,
		"mahasiswa berhasil ditambahkan",
		created,
		nil,
	)
}

// ============================================================
// GET STUDENTS
// GET /api/v1/students
// ============================================================

func (h *Handler) GetStudents(c *fiber.Ctx) error {

	page := 1
	perPage := 10

	// Pagination: page
	if value := c.Query("page"); value != "" {

		parsed, err := strconv.Atoi(value)

		if err != nil || parsed < 1 {
			return helper.Validation(map[string]string{
				"page": "page harus berupa angka minimal 1",
			})
		}

		page = parsed
	}

	// Pagination: per_page
	if value := c.Query("per_page"); value != "" {

		parsed, err := strconv.Atoi(value)

		if err != nil || parsed < 1 || parsed > 50 {
			return helper.Validation(map[string]string{
				"per_page": "per_page harus antara 1 sampai 50",
			})
		}

		perPage = parsed
	}

	// Filter angkatan
	var angkatan *int

	if value := c.Query("angkatan"); value != "" {

		parsed, err := strconv.Atoi(value)

		if err != nil {
			return helper.Validation(map[string]string{
				"angkatan": "angkatan harus berupa angka",
			})
		}

		angkatan = &parsed
	}

	offset := (page - 1) * perPage

	students, total, err := h.Students.FindAll(
		c.Context(),
		c.Query("search"),
		c.Query("prodi"),
		angkatan,
		c.Query("sort"),
		perPage,
		offset,
	)

	if err != nil {
		return helper.Internal(err)
	}

	lastPage := 0

	if total > 0 {
		lastPage = (total + perPage - 1) / perPage
	}

	meta := fiber.Map{
		"current_page": page,
		"per_page":     perPage,
		"total":        total,
		"last_page":    lastPage,
	}

	return helper.SuccessResponse(
		c,
		fiber.StatusOK,
		"Data mahasiswa berhasil diambil",
		students,
		meta,
	)
}
func (h *Handler) GetStudentByID(c *fiber.Ctx) error {

	// Ambil role dari token
	role, ok := c.Locals("role").(string)
	if !ok {
		return helper.Unauthorized("token tidak valid")
	}

	// Ambil ID mahasiswa dari URL
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return helper.BadRequest("id mahasiswa harus berupa angka")
	}

	// Cari mahasiswa berdasarkan ID
	student, err := h.Students.FindByID(
		c.Context(),
		id,
	)

	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound("mahasiswa tidak ditemukan")
		}

		return helper.Internal(err)
	}

	// =====================================================
	// ADMIN
	// Admin boleh melihat mahasiswa mana pun
	// =====================================================

	if role == "admin" {
		return helper.SuccessResponse(
			c,
			fiber.StatusOK,
			"data mahasiswa berhasil diambil",
			student,
			nil,
		)
	}

	// =====================================================
	// MAHASISWA
	// Hanya boleh melihat dirinya sendiri
	// =====================================================

	if role == "mahasiswa" {

		userID, ok := c.Locals("user_id").(int64)

		if !ok {
			return helper.Unauthorized(
				"identitas user tidak valid",
			)
		}

		// Repository membutuhkan int,
		// sedangkan JWT menyimpan user_id sebagai int64.
		ownStudent, err := h.Students.FindByUserID(
			c.Context(),
			int(userID),
		)

		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return helper.NotFound(
					"data mahasiswa tidak ditemukan",
				)
			}

			return helper.Internal(err)
		}

		// Cek apakah ID yang diminta adalah miliknya sendiri
		if ownStudent.ID != id {
			return helper.Forbidden(
				"anda hanya dapat mengakses data sendiri",
			)
		}

		return helper.SuccessResponse(
			c,
			fiber.StatusOK,
			"data mahasiswa berhasil diambil",
			student,
			nil,
		)
	}

	// Role selain admin dan mahasiswa ditolak
	return helper.Forbidden("akses ditolak")
}
// ============================================================
// UPDATE STUDENT
// PUT /api/v1/students/:id
//
// NIM tidak boleh diubah.
// Hanya admin.
// ============================================================

func (h *Handler) UpdateStudent(c *fiber.Ctx) error {

	// Hanya admin
	role, ok := c.Locals("role").(string)

	if !ok || role != "admin" {
		return helper.Forbidden(
			"akses hanya diperbolehkan untuk admin",
		)
	}

	id, err := strconv.Atoi(c.Params("id"))

	if err != nil || id < 1 {
		return helper.NotFound(
			"mahasiswa tidak ditemukan",
		)
	}

	var req model.UpdateStudentRequest

	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest(
			"format JSON tidak valid",
		)
	}

	if fields := helper.ValidateStruct(req); len(fields) > 0 {
		return helper.Validation(fields)
	}

	currentYear := time.Now().Year()

	if req.Angkatan < 1900 || req.Angkatan > currentYear {
		return helper.Validation(map[string]string{
			"angkatan": "angkatan harus 4 digit dan tidak boleh melebihi tahun berjalan",
		})
	}

	student := model.Student{
		ID:          id,
		Nama:        strings.TrimSpace(req.Nama),
		Prodi:       strings.TrimSpace(req.Prodi),
		Angkatan:    req.Angkatan,
		IPKTerakhir: req.IPKTerakhir,
	}

	updated, err := h.Students.Update(
		c.Context(),
		id,
		student,
	)

	if err != nil {

		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound(
				"mahasiswa tidak ditemukan",
			)
		}

		return helper.Internal(err)
	}

	return helper.SuccessResponse(
		c,
		fiber.StatusOK,
		"data mahasiswa berhasil diperbarui",
		updated,
		nil,
	)
}

// ============================================================
// DELETE STUDENT
// DELETE /api/v1/students/:id
//
// Soft delete.
// Hanya admin.
// ============================================================

func (h *Handler) DeleteStudent(c *fiber.Ctx) error {

	// Hanya admin
	role, ok := c.Locals("role").(string)

	if !ok || role != "admin" {
		return helper.Forbidden(
			"akses hanya diperbolehkan untuk admin",
		)
	}

	id, err := strconv.Atoi(c.Params("id"))

	if err != nil || id < 1 {
		return helper.NotFound(
			"mahasiswa tidak ditemukan",
		)
	}

	if err := h.Students.Delete(
		c.Context(),
		id,
	); err != nil {

		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound(
				"mahasiswa tidak ditemukan",
			)
		}

		return helper.Internal(err)
	}

	return c.SendStatus(
		fiber.StatusNoContent,
	)
}

// ============================================================
// HELPER
// ============================================================

func contextWithTimeout(
	parent context.Context,
	duration time.Duration,
) (context.Context, context.CancelFunc) {

	return context.WithTimeout(
		parent,
		duration,
	)
}

// ============================================================
// GET COURSES
// GET /api/v1/courses
//
// Bisa diakses semua user yang sudah login.
//
// Query:
// - semester
// - search
// - available=true
// ============================================================

func (h *Handler) GetCourses(c *fiber.Ctx) error {

	var semester *int

	if value := c.Query("semester"); value != "" {

		parsed, err := strconv.Atoi(value)

		if err != nil || parsed < 1 {
			return helper.Validation(map[string]string{
				"semester": "semester harus berupa angka minimal 1",
			})
		}

		semester = &parsed
	}

	available := strings.EqualFold(
		c.Query("available"),
		"true",
	)

	courses, err := h.Courses.FindAll(
		c.Context(),
		semester,
		c.Query("search"),
		available,
	)

	if err != nil {
		return helper.Internal(err)
	}

	return helper.SuccessResponse(
		c,
		fiber.StatusOK,
		"data mata kuliah berhasil diambil",
		courses,
		nil,
	)
}

// ============================================================
// CREATE ENROLLMENT
// POST /api/v1/enrollments
//
// Hanya mahasiswa.
// ============================================================

func (h *Handler) CreateEnrollment(c *fiber.Ctx) error {

	role, ok := c.Locals("role").(string)
	if !ok {
		return helper.Unauthorized("token tidak valid")
	}

	if role != "mahasiswa" {
		return helper.Forbidden("hanya mahasiswa yang dapat mengambil mata kuliah")
	}

	// kode CreateEnrollment yang sudah ada di bawahnya lanjutkan di sini...
	// =====================================================
	// AMBIL USER ID
	// =====================================================

	userID, ok := c.Locals("user_id").(int64)

	if !ok {
		return helper.Unauthorized(
			"identitas user tidak ditemukan",
		)
	}

	// =====================================================
	// CARI DATA MAHASISWA
	// =====================================================

	student, err := h.Students.FindByUserID(
		c.Context(),
		int(userID),
	)

	if err != nil {

		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound(
				"data mahasiswa tidak ditemukan",
			)
		}

		return helper.Internal(err)
	}

	// =====================================================
	// PARSE REQUEST
	// =====================================================

	var req model.CreateEnrollmentRequest

	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest(
			"format JSON tidak valid",
		)
	}

	if fields := helper.ValidateStruct(req); len(fields) > 0 {
		return helper.Validation(fields)
	}

	// =====================================================
	// VALIDASI TAHUN AKADEMIK
	// Contoh: 2026/2027 Ganjil
	// =====================================================

	tahunAkademik := strings.TrimSpace(
		req.TahunAkademik,
	)

	parts := strings.Fields(tahunAkademik)

	if len(parts) != 2 {
		return helper.Validation(map[string]string{
			"tahun_akademik": "format harus seperti 2026/2027 Ganjil",
		})
	}

	yearParts := strings.Split(
		parts[0],
		"/",
	)

	if len(yearParts) != 2 ||
		len(yearParts[0]) != 4 ||
		len(yearParts[1]) != 4 {

		return helper.Validation(map[string]string{
			"tahun_akademik": "format harus seperti 2026/2027 Ganjil",
		})
	}

	if parts[1] != "Ganjil" &&
		parts[1] != "Genap" {

		return helper.Validation(map[string]string{
			"tahun_akademik": "semester akademik harus Ganjil atau Genap",
		})
	}

	// =====================================================
	// BATAS SKS BERDASARKAN IPK
	// =====================================================

	var maxSKS int

	ipk := float64(0)

	if student.IPKTerakhir != nil {
		ipk = *student.IPKTerakhir
	}

	switch {

	case ipk >= 3.00:
		maxSKS = 24

	case ipk >= 2.50:
		maxSKS = 21

	default:
		maxSKS = 18
	}

	// =====================================================
	// SIMPAN KRS
	// =====================================================

	enrollment, err := h.Enrollments.Create(
		c.Context(),
		student.ID,
		req.CourseID,
		tahunAkademik,
		maxSKS,
	)

	if err != nil {

		if errors.Is(
			err,
			repository.ErrEnrollmentDuplicate,
		) {
			return helper.Conflict(
				"mata kuliah sudah diambil pada tahun akademik tersebut",
			)
		}

		if errors.Is(
			err,
			repository.ErrCourseFull,
		) {
			return helper.Validation(map[string]string{
				"course_id": "kuota mata kuliah sudah penuh",
			})
		}

		if errors.Is(
			err,
			repository.ErrSKSLimit,
		) {
			return helper.Validation(map[string]string{
				"course_id": err.Error(),
			})
		}

		return helper.Internal(err)
	}

	return helper.SuccessResponse(
		c,
		fiber.StatusCreated,
		"mata kuliah berhasil ditambahkan ke KRS",
		enrollment,
		nil,
	)
}

// ============================================================
// DELETE ENROLLMENT
// DELETE /api/v1/enrollments/:id
//
// Mahasiswa hanya boleh menghapus KRS miliknya sendiri.
// ============================================================

func (h *Handler) DeleteEnrollment(c *fiber.Ctx) error {

	userID, ok := c.Locals("user_id").(int64)

	if !ok {
		return helper.Unauthorized(
			"identitas user tidak ditemukan",
		)
	}

	student, err := h.Students.FindByUserID(
		c.Context(),
		int(userID),
	)

	if err != nil {

		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound(
				"data mahasiswa tidak ditemukan",
			)
		}

		return helper.Internal(err)
	}

	enrollmentID, err := strconv.Atoi(
		c.Params("id"),
	)

	if err != nil || enrollmentID < 1 {
		return helper.NotFound(
			"data KRS tidak ditemukan",
		)
	}

	err = h.Enrollments.DeleteOwned(
		c.Context(),
		enrollmentID,
		student.ID,
	)

	if err != nil {

		if errors.Is(
			err,
			repository.ErrEnrollmentNotFound,
		) {
			return helper.NotFound(
				"data KRS tidak ditemukan",
			)
		}

		if errors.Is(
			err,
			repository.ErrEnrollmentForbidden,
		) {
			return helper.Forbidden(
				"tidak dapat menghapus KRS mahasiswa lain",
			)
		}

		return helper.Internal(err)
	}

	return c.SendStatus(
		fiber.StatusNoContent,
	)
}
