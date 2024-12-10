package entity

import (
	mongodbutils "github.com/ct-logic-api-document/utils/mongodb"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RequestStructure struct {
	mongodbutils.BaseEntity `bson:",inline"`
	ApiId                   primitive.ObjectID `json:"api_id" bson:"api_id"`
	Parameters              []*Parameter       `json:"parameters" bson:"parameters"`
	BodySchema              map[string]any     `json:"body_schema" bson:"body_schema"`
}
