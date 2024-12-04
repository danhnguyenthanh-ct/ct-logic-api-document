package mongodb

import (
	"context"

	"github.com/carousell/ct-go/pkg/container"
	"github.com/ct-logic-api-document/internal/constants"
	"github.com/ct-logic-api-document/internal/entity"
	mongodbutils "github.com/ct-logic-api-document/utils/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
)

type IApiCollection interface {
	CreateApi(ctx context.Context, api *entity.Api) error
	GetApiByPath(ctx context.Context, path string) (*entity.Api, error)
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
