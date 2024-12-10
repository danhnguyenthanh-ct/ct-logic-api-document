package entity

import (
	mongodbutils "github.com/ct-logic-api-document/utils/mongodb"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ResponseStructure struct {
	mongodbutils.BaseEntity `bson:",inline"`
	ApiId                   primitive.ObjectID `json:"api_id" bson:"api_id"`
	BodySchema              map[string]any     `json:"body_schema" bson:"body_schema"`
}
