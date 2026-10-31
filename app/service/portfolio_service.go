package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"api-students/app/model"
	"api-students/app/repository"

	"github.com/gofiber/fiber/v2"
)

type PortfolioService struct {
	students   repository.StudentRepository
	portfolios repository.PortfolioRepository
}

func NewPortfolioService(
	students repository.StudentRepository,
	portfolios repository.PortfolioRepository,
) *PortfolioService {
	return &PortfolioService{
		students:   students,
		portfolios: portfolios,
	}
}

func portfolioResponse(
	c *fiber.Ctx,
	status int,
	message string,
	data interface{},
) error {
	return c.Status(status).JSON(fiber.Map{
		"success": status < 400,
		"message": message,
		"data":    data,
	})
}

func portfolioID(c *fiber.Ctx) (int, error) {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("id mahasiswa harus berupa angka positif")
	}
	return id, nil
}

func decodePortfolioJSON(c *fiber.Ctx, target interface{}) error {
	if !strings.HasPrefix(
		strings.ToLower(c.Get("Content-Type")),
		"application/json",
	) {
		return fmt.Errorf("Content-Type harus application/json")
	}

	decoder := json.NewDecoder(bytes.NewReader(c.Body()))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf(
			"body JSON tidak valid atau berisi field yang tidak dikenal: %w",
			err,
		)
	}

	var extra interface{}
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("body hanya boleh berisi satu objek JSON")
		}
		return fmt.Errorf("JSON tambahan tidak valid: %w", err)
	}

	return nil
}

func validatePortfolio(req model.SavePortfolioRequest) error {
	if strings.TrimSpace(req.Headline) == "" || len(req.Headline) > 150 {
		return fmt.Errorf("headline wajib diisi dan maksimal 150 karakter")
	}

	if strings.TrimSpace(req.Bio) == "" || len(req.Bio) > 2000 {
		return fmt.Errorf("bio wajib diisi dan maksimal 2000 karakter")
	}

	if len(req.Projects) > 20 {
		return fmt.Errorf("portfolio maksimal berisi 20 project")
	}

	for _, p := range req.Projects {
		if err := validateProject(
			p.Title,
			p.Technologies,
			p.Link,
			p.Year,
		); err != nil {
			return err
		}
	}

	for _, cert := range req.Certificates {
		if strings.TrimSpace(cert.Name) == "" ||
			strings.TrimSpace(cert.Issuer) == "" ||
			strings.TrimSpace(cert.Date) == "" {
			return fmt.Errorf(
				"nama, penerbit, dan tanggal sertifikat wajib diisi",
			)
		}

		if _, err := time.Parse("2006-01-02", cert.Date); err != nil {
			return fmt.Errorf("tanggal sertifikat harus berformat YYYY-MM-DD")
		}
	}

	return nil
}

func validateProject(
	title string,
	technologies []string,
	link string,
	year int,
) error {
	if strings.TrimSpace(title) == "" || len(title) > 150 {
		return fmt.Errorf("judul project wajib diisi dan maksimal 150 karakter")
	}

	if len(technologies) == 0 || len(technologies) > 20 {
		return fmt.Errorf("teknologi harus berisi 1 sampai 20 item")
	}

	for _, technology := range technologies {
		if strings.TrimSpace(technology) == "" || len(technology) > 50 {
			return fmt.Errorf(
				"nama teknologi tidak boleh kosong dan maksimal 50 karakter",
			)
		}
	}

	if link != "" {
		u, err := url.ParseRequestURI(link)
		if err != nil ||
			(u.Scheme != "https" && u.Scheme != "http") ||
			u.Host == "" {
			return fmt.Errorf(
				"link project harus berupa URL HTTP atau HTTPS yang valid",
			)
		}
	}

	if year < 1900 || year > time.Now().Year()+1 {
		return fmt.Errorf("tahun project tidak valid")
	}

	return nil
}

func newProjectID() (string, error) {
	var b [16]byte

	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}

	return hex.EncodeToString(b[:]), nil
}

func (s *PortfolioService) ensureStudent(
	ctx context.Context,
	studentID int,
) error {
	_, err := s.students.FindByID(ctx, studentID)
	return err
}

func portfolioError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return portfolioResponse(
			c, fiber.StatusNotFound, "mahasiswa tidak ditemukan", nil,
		)

	case errors.Is(err, repository.ErrPortfolioNotFound):
		return portfolioResponse(
			c, fiber.StatusNotFound, "portfolio tidak ditemukan", nil,
		)

	case errors.Is(err, repository.ErrPortfolioExists):
		return portfolioResponse(
			c, fiber.StatusConflict, "portfolio sudah ada", nil,
		)

	case errors.Is(err, repository.ErrProjectNotFound):
		return portfolioResponse(
			c, fiber.StatusNotFound, "project tidak ditemukan", nil,
		)

	case errors.Is(err, repository.ErrProjectLimit),
		errors.Is(err, repository.ErrProjectLimitReached):
		return portfolioResponse(
			c, fiber.StatusConflict,
			"portfolio maksimal berisi 20 project", nil,
		)

	default:
		return portfolioResponse(
			c, fiber.StatusInternalServerError,
			"terjadi kesalahan pada server", nil,
		)
	}
}

// GET /api/v1/students/:id/portfolio
func (s *PortfolioService) Get(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Second)
	defer cancel()

	id, err := portfolioID(c)
	if err != nil {
		return portfolioResponse(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := s.ensureStudent(ctx, id); err != nil {
		return portfolioError(c, err)
	}

	portfolio, err := s.portfolios.FindByStudentID(ctx, id)
	if err != nil {
		return portfolioError(c, err)
	}

	return portfolioResponse(
		c, fiber.StatusOK, "portfolio berhasil diambil", portfolio,
	)
}

// PUT /api/v1/students/:id/portfolio
func (s *PortfolioService) Put(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Second)
	defer cancel()

	id, err := portfolioID(c)
	if err != nil {
		return portfolioResponse(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := s.ensureStudent(ctx, id); err != nil {
		return portfolioError(c, err)
	}

	var req model.SavePortfolioRequest
	if err := decodePortfolioJSON(c, &req); err != nil {
		return portfolioResponse(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := validatePortfolio(req); err != nil {
		return portfolioResponse(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	projects := req.Projects
	if projects == nil {
		projects = []model.PortfolioProject{}
	}

	for i := range projects {
		if projects[i].ProjectID == "" {
			projectID, err := newProjectID()
			if err != nil {
				return portfolioResponse(
					c, fiber.StatusInternalServerError,
					"gagal membuat ID project", nil,
				)
			}
			projects[i].ProjectID = projectID
		}
	}

	certificates := req.Certificates
	if certificates == nil {
		certificates = []model.PortfolioCertificate{}
	}

	portfolio := model.Portfolio{
		StudentID:    id,
		Headline:     strings.TrimSpace(req.Headline),
		Bio:          strings.TrimSpace(req.Bio),
		Projects:     projects,
		Certificates: certificates,
	}

	if err := s.portfolios.Replace(ctx, portfolio); err != nil {
		return portfolioError(c, err)
	}

	return portfolioResponse(
		c, fiber.StatusOK, "portfolio berhasil disimpan", portfolio,
	)
}

// POST /api/v1/students/:id/portfolio/projects
func (s *PortfolioService) AddProject(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Second)
	defer cancel()

	id, err := portfolioID(c)
	if err != nil {
		return portfolioResponse(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := s.ensureStudent(ctx, id); err != nil {
		return portfolioError(c, err)
	}

	var req model.CreatePortfolioProjectRequest
	if err := decodePortfolioJSON(c, &req); err != nil {
		return portfolioResponse(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := validateProject(
		req.Title, req.Technologies, req.Link, req.Year,
	); err != nil {
		return portfolioResponse(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	projectID, err := newProjectID()
	if err != nil {
		return portfolioResponse(
			c, fiber.StatusInternalServerError,
			"gagal membuat ID project", nil,
		)
	}

	project := model.PortfolioProject{
		ProjectID:    projectID,
		Title:        strings.TrimSpace(req.Title),
		Technologies: req.Technologies,
		Link:         req.Link,
		Year:         req.Year,
	}

	if err := s.portfolios.AddProject(ctx, id, project); err != nil {
		return portfolioError(c, err)
	}

	return portfolioResponse(
		c, fiber.StatusCreated, "project berhasil ditambahkan", project,
	)
}

// DELETE /api/v1/students/:id/portfolio/projects/:projectId
func (s *PortfolioService) DeleteProject(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Second)
	defer cancel()

	id, err := portfolioID(c)
	if err != nil {
		return portfolioResponse(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	projectID := strings.TrimSpace(c.Params("projectId"))
	if projectID == "" {
		return portfolioResponse(
			c, fiber.StatusBadRequest, "projectId wajib diisi", nil,
		)
	}

	if err := s.ensureStudent(ctx, id); err != nil {
		return portfolioError(c, err)
	}

	if err := s.portfolios.DeleteProject(ctx, id, projectID); err != nil {
		return portfolioError(c, err)
	}

	return portfolioResponse(
		c, fiber.StatusOK, "project berhasil dihapus", nil,
	)
}

// GET /api/v1/portfolios/stats/technologies
func (s *PortfolioService) TechnologyStats(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	stats, err := s.portfolios.TechnologyStats(ctx)
	if err != nil {
		return portfolioError(c, err)
	}

	return portfolioResponse(
		c, fiber.StatusOK, "statistik teknologi berhasil diambil", stats,
	)
}
