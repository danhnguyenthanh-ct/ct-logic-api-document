package usecase

import (
	"testing"

	"github.com/ct-logic-api-document/config"
	"github.com/ct-logic-api-document/internal/repository/mongodb"
)

var (
	mongoStorage mongodb.MongoStorage
	conf         *config.Config
)

func TestMain(m *testing.M) {
	testMainWrapper(m)
}

func testMainWrapper(m *testing.M) int {
	conf = config.MustLoad()
	mongoStorage = mongodb.NewMongoStorage(conf)
	return m.Run()
}
