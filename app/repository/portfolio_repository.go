package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"api-students/app/model"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	ErrPortfolioNotFound   = errors.New("portfolio tidak ditemukan")
	ErrPortfolioExists     = errors.New("portfolio sudah ada")
	ErrProjectLimit        = errors.New("maksimal 20 project")
	ErrProjectNotFound     = errors.New("project tidak ditemukan")
	ErrProjectLimitReached = errors.New("jumlah project sudah mencapai batas maksimal")
)

const maxPortfolioProjects = 20

type PortfolioRepository interface {
	FindByStudentID(ctx context.Context, studentID int) (model.Portfolio, error)
	Create(ctx context.Context, portfolio model.Portfolio) error
	Replace(ctx context.Context, portfolio model.Portfolio) error
	AddProject(ctx context.Context, studentID int, project model.PortfolioProject) error
	DeleteProject(ctx context.Context, studentID int, projectID string) error
	TechnologyStats(ctx context.Context) ([]bson.M, error)
}

type portfolioRepository struct {
	collection *mongo.Collection
}

func NewPortfolioRepository(db *mongo.Database) PortfolioRepository {
	return &portfolioRepository{
		collection: db.Collection("student_portfolios"),
	}
}

func (r *portfolioRepository) FindByStudentID(
	ctx context.Context,
	studentID int,
) (model.Portfolio, error) {
	var portfolio model.Portfolio

	err := r.collection.FindOne(
		ctx,
		bson.M{"student_id": studentID},
	).Decode(&portfolio)

	if errors.Is(err, mongo.ErrNoDocuments) {
		return model.Portfolio{}, ErrPortfolioNotFound
	}
	if err != nil {
		return model.Portfolio{}, fmt.Errorf("gagal membaca portfolio: %w", err)
	}

	return portfolio, nil
}

func (r *portfolioRepository) Create(
	ctx context.Context,
	portfolio model.Portfolio,
) error {
	_, err := r.collection.InsertOne(ctx, portfolio)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrPortfolioExists
		}
		return fmt.Errorf("gagal membuat portfolio: %w", err)
	}

	return nil
}


func (r *portfolioRepository) Replace(
	ctx context.Context,
	portfolio model.Portfolio,
) error {
	_, err := r.collection.ReplaceOne(
		ctx,
		bson.M{"student_id": portfolio.StudentID},
		portfolio,
		options.Replace().SetUpsert(true),
	)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrPortfolioExists
		}
		return fmt.Errorf("gagal menyimpan portfolio: %w", err)
	}

	return nil
}


func (r *portfolioRepository) AddProject(
	ctx context.Context,
	studentID int,
	project model.PortfolioProject,
) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{
		"student_id": studentID,
		"$expr": bson.M{
			"$lt": []interface{}{
				bson.M{"$size": bson.M{
					"$ifNull": []interface{}{"$projects", []interface{}{}},
				}},
				maxPortfolioProjects,
			},
		},
	}

	result, err := r.collection.UpdateOne(
		ctx,
		filter,
		bson.M{"$push": bson.M{"projects": project}},
	)
	if err != nil {
		return fmt.Errorf("gagal menambahkan project: %w", err)
	}

	if result.MatchedCount == 0 {
		portfolio, findErr := r.FindByStudentID(ctx, studentID)
		if findErr != nil {
			return findErr
		}
		if len(portfolio.Projects) >= maxPortfolioProjects {
			return ErrProjectLimitReached
		}
		return ErrPortfolioNotFound
	}

	return nil
}

func (r *portfolioRepository) DeleteProject(
	ctx context.Context,
	studentID int,
	projectID string,
) error {
	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"student_id": studentID},
		bson.M{"$pull": bson.M{
			"projects": bson.M{"project_id": projectID},
		}},
	)
	if err != nil {
		return fmt.Errorf("gagal menghapus project: %w", err)
	}

	if result.MatchedCount == 0 {
		return ErrPortfolioNotFound
	}
	if result.ModifiedCount == 0 {
		return ErrProjectNotFound
	}

	return nil
}

func (r *portfolioRepository) TechnologyStats(
	ctx context.Context,
) ([]bson.M, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$unwind", Value: "$projects"}},
		{{Key: "$unwind", Value: "$projects.technologies"}},
		{{Key: "$group", Value: bson.M{
			"_id":   "$projects.technologies",
			"total": bson.M{"$sum": 1},
		}}},
		{{Key: "$sort", Value: bson.D{
			{Key: "total", Value: -1},
			{Key: "_id", Value: 1},
		}}},
		{{Key: "$limit", Value: 10}},
		{{Key: "$project", Value: bson.M{
			"_id":        0,
			"technology": "$_id",
			"total":      1,
		}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("gagal menghitung statistik teknologi: %w", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("gagal membaca statistik teknologi: %w", err)
	}
	if results == nil {
		results = []bson.M{}
	}

	return results, nil
}
