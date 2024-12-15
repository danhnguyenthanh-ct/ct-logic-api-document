package entity

import (
	"github.com/carousell/ct-go/pkg/container"
	mongodbutils "github.com/ct-logic-api-document/utils/mongodb"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RequestStructure struct {
	mongodbutils.BaseEntity `bson:",inline"`
	ApiId                   primitive.ObjectID `json:"api_id" bson:"api_id"`
	Parameters              []*Parameter       `json:"parameters" bson:"parameters"`
	BodySchema              map[string]any     `json:"body_schema" bson:"body_schema"`
}

func (r *RequestStructure) BuildRequestBody() any {
	if len(r.BodySchema) == 0 {
		return nil
	}
	return container.Map{
		"content": container.Map{
			"application/json": container.Map{
				"schema": r.BodySchema,
			},
		},
	}
}

func (r *RequestStructure) BuildParameters() any {
	if r.Parameters == nil {
		return nil
	}
	parameters := make([]any, 0)
	for _, parameter := range r.Parameters {
		parameters = append(parameters, parameter.buildParameter())
	}
	return parameters
}
