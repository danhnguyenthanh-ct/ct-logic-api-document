package entity

import (
	"time"

	mongodbutils "github.com/ct-logic-api-document/utils/mongodb"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SampleResponse struct {
	mongodbutils.BaseEntity `bson:",inline"`
	ApiId                   primitive.ObjectID `json:"api_id" bson:"api_id"`
	HttpStatusCode          int                `json:"status_code" bson:"http_status_code"`
	Body                    string             `json:"body" bson:"body"`
}

type GetSampleResponseByApiIdRequest struct {
	ApiId  primitive.ObjectID `json:"api_id"`
	Limit  int64              `json:"limit"`
	Offset int64              `json:"offset"`
	From   *time.Time         `json:"from"`
	To     *time.Time         `json:"to"`
}
