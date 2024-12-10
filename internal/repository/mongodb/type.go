package mongodb

import (
	"context"

	"github.com/ct-logic-api-document/internal/constants"
	"github.com/ct-logic-api-document/internal/entity"
	mongodbutils "github.com/ct-logic-api-document/utils/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
)

type ITypeCollection interface{}

type TypeCollection struct {
	mongodbutils.BaseCollection[entity.Type, *entity.Type]
}

var _ ITypeCollection = (*TypeCollection)(nil)

func NewTypeCollection(db *mongo.Database) *TypeCollection {
	baseCollection := mongodbutils.NewBaseCollection[entity.Type](db, constants.TypesCollection)
	return &TypeCollection{
		BaseCollection: *baseCollection,
	}
}

func (c *TypeCollection) CreateType(ctx context.Context, t *entity.Type) error {
	return c.Insert(ctx, t)
}
