package mongodb

import (
	"context"
	"errors"
	"time"

	"github.com/carousell/ct-go/pkg/container"
	logctx "github.com/carousell/ct-go/pkg/logger/log_context"
	"github.com/ct-logic-api-document/internal/constants"
	"github.com/ct-logic-api-document/internal/entity"
	mongodbutils "github.com/ct-logic-api-document/utils/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type IApiCollection interface {
	CreateApi(ctx context.Context, api *entity.Api) error
	GetApiByPath(ctx context.Context, path string) (*entity.Api, error)
	GetApis(ctx context.Context, limit, offset int64) ([]*entity.Api, error)
	CountApis(ctx context.Context) (int64, error)
	UpdateLatestBuildStructure(ctx context.Context, id primitive.ObjectID, latestBuildStructure *time.Time) error
	GetApiByIdInStr(ctx context.Context, id string) (*entity.Api, error)
}

type ApiCollection struct {
	mongodbutils.BaseCollection[entity.Api, *entity.Api]
}

var _ IApiCollection = (*ApiCollection)(nil)

func NewApiCollection(db *mongo.Database) *ApiCollection {
	baseCollection := mongodbutils.NewBaseCollection[entity.Api](db, constants.ApisCollection)
	return &ApiCollection{
		BaseCollection: *baseCollection,
	}
}

func (a *ApiCollection) CreateApi(ctx context.Context, api *entity.Api) error {
	return a.Insert(ctx, api)
}

func (a *ApiCollection) GetApiByPath(ctx context.Context, path string) (*entity.Api, error) {
	filter := container.Map{
		"path": path,
	}
	return a.Get(ctx, filter)
}

func (a *ApiCollection) GetApis(ctx context.Context, limit, offset int64) ([]*entity.Api, error) {
	filter := container.Map{}
	sort := bson.D{{Key: "created_at", Value: -1}}
	apis, err := a.GetByBatch(ctx, filter, sort, limit, offset)
	if err != nil {
		return nil, err
	}
	return apis, nil
}

func (a *ApiCollection) CountApis(ctx context.Context) (int64, error) {
	filter := container.Map{}
	return a.CountByFilter(ctx, filter)
}

func (a *ApiCollection) UpdateLatestBuildStructure(ctx context.Context, id primitive.ObjectID, latestBuildStructure *time.Time) error {
	filter := container.Map{
		"_id": id,
	}
	update := container.Map{
		"latest_build_structure": latestBuildStructure,
	}
	updateResult, err := a.UpdatePartial(ctx, filter, update)
	if err != nil {
		return err
	}
	if updateResult.MatchedCount == 0 {
		logctx.Infow(ctx, "no documents found", "id", id)
		return errors.New("no documents found")
	}
	if updateResult.ModifiedCount == 0 {
		logctx.Infow(ctx, "no documents updated", "id", id)
		return errors.New("no documents updated")
	}
	return nil
}

func (a *ApiCollection) GetApiByIdInStr(ctx context.Context, id string) (*entity.Api, error) {
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	filter := container.Map{
		"_id": objId,
	}
	return a.Get(ctx, filter)
}
