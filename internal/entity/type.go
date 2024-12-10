package entity

import mongodbutils "github.com/ct-logic-api-document/utils/mongodb"

type Type struct {
	mongodbutils.BaseEntity `bson:",inline"`
	Name                    string         `json:"name" bson:"name"`
	Properties              map[string]any `json:"properties" bson:"properties"`
}
