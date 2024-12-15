package entity

import (
	"time"

	"github.com/carousell/ct-go/pkg/container"
	mongodbutils "github.com/ct-logic-api-document/utils/mongodb"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SampleRequest struct {
	mongodbutils.BaseEntity `bson:",inline"`
	ApiId                   primitive.ObjectID `json:"api_id" bson:"api_id"`
	Parameters              []*Parameter       `json:"parameters" bson:"parameters"`
	Body                    string             `json:"body" bson:"body"`
}

type Parameter struct {
	Name     string `json:"name" bson:"name"`
	Type     string `json:"type" bson:"type"`
	Value    string `json:"value" bson:"value"`
	In       string `json:"in" bson:"in"`
	Required bool   `json:"required" bson:"required"`
}

func (r *Parameter) buildParameter() any {
	return container.Map{
		"name":     r.Name,
		"in":       r.In,
		"required": r.Required,
		"schema": container.Map{
			"type": r.Type,
		},
	}
}

type GetSampleRequestByApiIdRequest struct {
	ApiId  primitive.ObjectID `json:"api_id"`
	Limit  int64              `json:"limit"`
	Offset int64              `json:"offset"`
	From   *time.Time         `json:"from"`
	To     *time.Time         `json:"to"`
}
