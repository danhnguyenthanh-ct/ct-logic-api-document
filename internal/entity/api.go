package entity

import (
	"time"

	mongodbutils "github.com/ct-logic-api-document/utils/mongodb"
)

type Api struct {
	mongodbutils.BaseEntity `bson:",inline"`
	Title                   string     `json:"title" bson:"title"`
	Host                    string     `json:"host" bson:"host"`
	Path                    string     `json:"path" bson:"path"`
	Method                  string     `json:"method" bson:"method"`
	Description             string     `json:"description" bson:"description"`
	LatestBuildStructure    *time.Time `json:"latest_build_structure" bson:"latest_build_structure"`
}
