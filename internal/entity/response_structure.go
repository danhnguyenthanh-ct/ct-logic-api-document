package entity

import (
	"github.com/carousell/ct-go/pkg/container"
	mongodbutils "github.com/ct-logic-api-document/utils/mongodb"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ResponseStructure struct {
	mongodbutils.BaseEntity `bson:",inline"`
	ApiId                   primitive.ObjectID `json:"api_id" bson:"api_id"`
	BodySchema              map[string]any     `json:"body_schema" bson:"body_schema"`
}

func (r *ResponseStructure) BuildResponseBody() any {
	if len(r.BodySchema) == 0 {
		return nil
	}
	return container.Map{
		"200": container.Map{
			"description": "Success",
			"content": container.Map{
				"application/json": container.Map{
					"schema": r.BodySchema,
				},
			},
		},
	}
}
