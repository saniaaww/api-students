
package model

type Portfolio struct {
	StudentID   int                    `json:"student_id" bson:"student_id"`
	Headline    string                 `json:"headline" bson:"headline"`
	Bio         string                 `json:"bio" bson:"bio"`
	Projects     []PortfolioProject     `json:"projects" bson:"projects"`
	Certificates []PortfolioCertificate `json:"certificates" bson:"certificates"`
}

type PortfolioProject struct {
	ProjectID   string   `json:"project_id" bson:"project_id"`
	Title       string   `json:"title" bson:"title"`
	Technologies []string `json:"technologies" bson:"technologies"`
	Link        string   `json:"link" bson:"link"`
	Year        int      `json:"year" bson:"year"`
}

type PortfolioCertificate struct {
	Name   string `json:"name" bson:"name"`
	Issuer string `json:"issuer" bson:"issuer"`
	Date   string `json:"date" bson:"date"`
}

type SavePortfolioRequest struct {
	Headline    string                 `json:"headline"`
	Bio         string                 `json:"bio"`
	Projects     []PortfolioProject     `json:"projects"`
	Certificates []PortfolioCertificate `json:"certificates"`
}

type CreatePortfolioProjectRequest struct {
	Title        string   `json:"title"`
	Technologies []string `json:"technologies"`
	Link         string   `json:"link"`
	Year         int      `json:"year"`
}
