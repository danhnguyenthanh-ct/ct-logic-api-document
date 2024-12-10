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

type ISampleRequestCollection interface {
	CreateSampleRequest(ctx context.Context, sampleRequest *entity.SampleRequest) error
	GetSampleRequestByApiId(ctx context.Context, req *entity.GetSampleRequestByApiIdRequest) ([]*entity.SampleRequest, error)
}

type SampleRequestCollection struct {
	mongodbutils.BaseCollection[entity.SampleRequest, *entity.SampleRequest]
}

var _ ISampleRequestCollection = (*SampleRequestCollection)(nil)

func NewSampleRequestCollection(db *mongo.Database) *SampleRequestCollection {
	baseCollection := mongodbutils.NewBaseCollection[entity.SampleRequest](db, constants.SampleRequestsCollection)
	return &SampleRequestCollection{
		BaseCollection: *baseCollection,
	}
}

func (s *SampleRequestCollection) CreateSampleRequest(ctx context.Context, sampleRequest *entity.SampleRequest) error {
	return s.Insert(ctx, sampleRequest)
}

func (s *SampleRequestCollection) GetSampleRequestByApiId(
	ctx context.Context,
	req *entity.GetSampleRequestByApiIdRequest,
) ([]*entity.SampleRequest, error) {
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
