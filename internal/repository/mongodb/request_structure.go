package mongodb

import (
	"context"
	"errors"

	"github.com/carousell/ct-go/pkg/container"
	"github.com/ct-logic-api-document/internal/constants"
	"github.com/ct-logic-api-document/internal/entity"
	mongodbutils "github.com/ct-logic-api-document/utils/mongodb"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type IRequestStructureCollection interface {
	CreateRequestStructure(ctx context.Context, requestStructure *entity.RequestStructure) error
	GetRequestStructureByApiId(ctx context.Context, apiId primitive.ObjectID) (*entity.RequestStructure, error)
	UpdateRequestStructure(ctx context.Context, id primitive.ObjectID, parameters []*entity.Parameter, bodySchema map[string]any) error
}

type RequestStructureCollection struct {
	mongodbutils.BaseCollection[entity.RequestStructure, *entity.RequestStructure]
}

var _ IRequestStructureCollection = (*RequestStructureCollection)(nil)

func NewRequestStructureCollection(db *mongo.Database) *RequestStructureCollection {
	baseCollection := mongodbutils.NewBaseCollection[entity.RequestStructure](db, constants.RequestStructuresCollection)
	return &RequestStructureCollection{
		BaseCollection: *baseCollection,
	}
}

func (r *RequestStructureCollection) CreateRequestStructure(
	ctx context.Context,
	requestStructure *entity.RequestStructure,
) error {
	return r.Insert(ctx, requestStructure)
}

func (r *RequestStructureCollection) GetRequestStructureByApiId(
	ctx context.Context,
	apiId primitive.ObjectID,
) (*entity.RequestStructure, error) {
	filter := map[string]any{
		"api_id": apiId,
	}
	return r.Get(ctx, filter)
}

func (r *RequestStructureCollection) UpdateRequestStructure(
	ctx context.Context,
	id primitive.ObjectID,
	parameters []*entity.Parameter,
	bodySchema map[string]any,
) error {
	filter := container.Map{
		"_id": id,
	}
	update := container.Map{
		"parameters":  parameters,
		"body_schema": bodySchema,
	}
	updatedResult, err := r.UpdatePartial(ctx, filter, update)
	if err != nil {
		return err
	}
	if updatedResult.MatchedCount == 0 {
		return errors.New("no documents found")
	}
	return nil
}
