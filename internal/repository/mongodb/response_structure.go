package mongodb

import (
	"context"
	"errors"

	logctx "github.com/carousell/ct-go/pkg/logger/log_context"
	"github.com/ct-logic-api-document/internal/constants"
	"github.com/ct-logic-api-document/internal/entity"
	mongodbutils "github.com/ct-logic-api-document/utils/mongodb"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type IResponseStructureCollection interface {
	CreateResponseStructure(ctx context.Context, responseStructure *entity.ResponseStructure) error
	GetResponseStructureByApiId(ctx context.Context, apiId primitive.ObjectID) (*entity.ResponseStructure, error)
	UpdateResponseStructure(ctx context.Context, id primitive.ObjectID, bodySchema map[string]any) error
}

type ResponseStructureCollection struct {
	mongodbutils.BaseCollection[entity.ResponseStructure, *entity.ResponseStructure]
}

var _ IResponseStructureCollection = (*ResponseStructureCollection)(nil)

func NewResponseStructureCollection(db *mongo.Database) *ResponseStructureCollection {
	baseCollection := mongodbutils.NewBaseCollection[entity.ResponseStructure](db, constants.ResponseStructuresCollection)
	return &ResponseStructureCollection{
		BaseCollection: *baseCollection,
	}
}

func (r *ResponseStructureCollection) CreateResponseStructure(
	ctx context.Context,
	responseStructure *entity.ResponseStructure,
) error {
	return r.Insert(ctx, responseStructure)
}

func (r *ResponseStructureCollection) GetResponseStructureByApiId(
	ctx context.Context,
	apiId primitive.ObjectID,
) (*entity.ResponseStructure, error) {
	filter := map[string]any{
		"api_id": apiId,
	}
	return r.Get(ctx, filter)
}

func (r *ResponseStructureCollection) UpdateResponseStructure(ctx context.Context, id primitive.ObjectID, bodySchema map[string]any) error {
	filter := map[string]any{
		"_id": id,
	}
	update := map[string]any{
		"body_schema": bodySchema,
	}
	updatedResult, err := r.UpdatePartial(ctx, filter, update)
	if err != nil {
		return err
	}
	if updatedResult.MatchedCount == 0 {
		logctx.Errorw(ctx, "no documents found", "id", id)
		return errors.New("no documents found")
	}
	return nil
}
