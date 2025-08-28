package services

import (
	"context"
	"time"

	"taraftar-analizi/config"
	"taraftar-analizi/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type TeamService struct {
	collection *mongo.Collection
}

func NewTeamService() *TeamService {
	return &TeamService{
		collection: config.GetCollection("teams"),
	}
}

func (ts *TeamService) GetAllTeams() ([]models.Team, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := ts.collection.Find(ctx, bson.M{"is_active": true})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var teams []models.Team
	if err = cursor.All(ctx, &teams); err != nil {
		return nil, err
	}

	return teams, nil
}