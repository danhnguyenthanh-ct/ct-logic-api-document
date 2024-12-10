package mongodb

import (
	"context"

	"github.com/carousell/ct-go/pkg/container"
	"github.com/ct-logic-api-document/internal/constants"
	"github.com/ct-logic-api-document/internal/entity"
	mongodbutils "github.com/ct-logic-api-document/utils/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type ISampleResponseCollection interface {
	CreateSampleResponse(ctx context.Context, sampleResponse *entity.SampleResponse) error
	GetSampleResponseByApiId(ctx context.Context, req *entity.GetSampleResponseByApiIdRequest) ([]*entity.SampleResponse, error)
}

type SampleResponseCollection struct {
	mongodbutils.BaseCollection[entity.SampleResponse, *entity.SampleResponse]
}

var _ ISampleResponseCollection = (*SampleResponseCollection)(nil)

func NewSampleResponseCollection(db *mongo.Database) *SampleResponseCollection {
	baseCollection := mongodbutils.NewBaseCollection[entity.SampleResponse](db, constants.SampleResponsesCollection)
	return &SampleResponseCollection{
		BaseCollection: *baseCollection,
	}
}

func (s *SampleResponseCollection) CreateSampleResponse(ctx context.Context, sampleResponse *entity.SampleResponse) error {
	return s.Insert(ctx, sampleResponse)
}

func (s *SampleResponseCollection) GetSampleResponseByApiId(ctx context.Context, req *entity.GetSampleResponseByApiIdRequest) ([]*entity.SampleResponse, error) {
	filter := container.Map{
		"api_id": req.ApiId,
	}
	if req.From != nil {
		filter["created_at"] = bson.M{"$gte": req.From}
	}
	if req.To != nil {
		filter["created_at"] = bson.M{"$lte": req.To}
	}
	sort := bson.D{{Key: "created_at", Value: -1}}
	return s.GetByBatch(ctx, filter, sort, req.Limit, req.Offset)
}
