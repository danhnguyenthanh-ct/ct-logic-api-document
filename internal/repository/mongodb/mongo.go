package mongodb

import (
	"context"
	"encoding/json"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/carousell/ct-go/pkg/logger"

	"github.com/ct-logic-api-document/config"
	mongodbutils "github.com/ct-logic-api-document/utils/mongodb"
)

type MongoStorage interface {
	IApiCollection
	ISampleRequestCollection
	ISampleResponseCollection
	IRequestStructureCollection
	IResponseStructureCollection
	ITypeCollection
}

type mongoStorage struct {
	log       *logger.Logger
	mgo       *mongo.Database
	conf      *config.Config
	mgoClient *mongo.Client

	ApiCollection
	SampleRequestCollection
	SampleResponseCollection
	RequestStructureCollection
	ResponseStructureCollection
	TypeCollection
}

var _ MongoStorage = &mongoStorage{}

func NewMongoStorage(
	conf *config.Config,
) MongoStorage {
	log := logger.MustNamed("mongodb")
	return newMongoStorage(log, conf)
}

func newMongoStorage(log *logger.Logger, conf *config.Config) *mongoStorage { //nolint:revive
	ctx := context.Background()
	connectionStr := conf.Mongo.ConnectionString
	poolSize := conf.Mongo.PoolSize
	opts := options.Client().ApplyURI(connectionStr).SetMaxPoolSize(poolSize)
	if conf.Mongo.Debug {
		mongodbutils.AddDebugOption(opts, log)
	}
	mongoClient, err := mongo.Connect(ctx, opts)
	if err != nil {
		log.Fatalf("failed to NewClient mongodb: %v", err) //nolint:revive
	}
	mongoDB := mongoClient.Database(conf.Mongo.DBName)
	log.Info("Connecting to mongodb")
	err = mongoClient.Ping(ctx, readpref.Primary())
	if err != nil {
		panic(err)
	}
	log.Info("Connected to mongodb")
	return &mongoStorage{
		log:       log,
		mgo:       mongoDB,
		conf:      conf,
		mgoClient: mongoClient,

		ApiCollection:               *NewApiCollection(mongoDB),
		SampleRequestCollection:     *NewSampleRequestCollection(mongoDB),
		SampleResponseCollection:    *NewSampleResponseCollection(mongoDB),
		RequestStructureCollection:  *NewRequestStructureCollection(mongoDB),
		ResponseStructureCollection: *NewResponseStructureCollection(mongoDB),
		TypeCollection:              *NewTypeCollection(mongoDB),
	}
}

func (m *mongoStorage) StopMongoDB() {
	ctx := context.Background()
	err := m.mgoClient.Disconnect(ctx)
	if err != nil {
		panic(err)
	}
}

func MongoGetOneWithOption[R any](ctx context.Context, m *mongoStorage, funcName string,
	collName string, filterObj any, opts *options.FindOneOptions,
) (*R, error) {
	res := new(R)
	err := m.mgo.Collection(collName).FindOne(ctx, filterObj, opts).Decode(res)
	if err != nil {
		m.log.Errorf("%s error, req: %+v, err: %s", funcName, filterObj, err)
		return nil, err
	}

	return res, nil
}

func MongoGetAll[S any, R any](ctx context.Context, m *mongoStorage, funcName string, collName string,
	filterObj primitive.M, sortObj primitive.D, resultTransformFunc func([]byte) (*R, error),
) (*R, error) {
	findAllOptions := options.Find()
	findAllOptions.SetSort(sortObj)

	cursor, err := m.mgo.Collection(collName).Find(ctx, filterObj)
	if err != nil {
		return nil, err
	}

	respArrData := new([]S)

	if err = cursor.All(ctx, respArrData); err != nil {
		m.log.Errorf("%s cursor all error: %v", funcName, err)
		return nil, err
	}

	m.log.Infof("Done %s data size:%d, data:%+v", funcName, len(*respArrData), respArrData)
	marshaledData, err := json.Marshal(respArrData)
	if err != nil {
		return nil, err
	}

	respData, err := resultTransformFunc(marshaledData)
	if respData == nil {
		return new(R), nil
	}
	return respData, err
}

func MongoInsertOne[S any, P any, R any](ctx context.Context, funcName string, m *mongoStorage,
	collName string, payload *P, resultTransformFunc func(string) (*R, error),
) (*R, error) {
	marshaledJSONStr, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	convertedPayload := new(S)
	err = json.Unmarshal(marshaledJSONStr, convertedPayload)
	if err != nil {
		return nil, err
	}

	insertResult, err := m.mgo.Collection(collName).InsertOne(ctx, convertedPayload)
	if err != nil {
		m.log.Errorf("%s error: %v", funcName, err)
		return nil, err
	}

	insertedId, err := insertResult.InsertedID.(primitive.ObjectID).MarshalText() //nolint:errcheck
	if err != nil {
		m.log.Errorf("%s error: %v", funcName, err)
		return nil, err
	}

	return resultTransformFunc(string(insertedId))
}

func MongoGetOne[S any, R any](ctx context.Context, m *mongoStorage, funcName string, collName string,
	filterObj primitive.M, sortObj primitive.D, resultTransformFunc func([]byte) (*R, error),
) (*R, error) {
	findOneOptions := options.FindOne()
	findOneOptions.SetSort(sortObj)

	result := m.mgo.Collection(collName).FindOne(ctx, filterObj)

	err := result.Err()
	if errors.Is(err, mongo.ErrNoDocuments) {
		return new(R), nil
	}

	if err != nil {
		m.log.Errorf("unable to find one: %v", err)
		return nil, err
	}

	respArrData := new(S)
	if err := result.Decode(respArrData); err != nil {
		m.log.Errorf("%s decode error: %v", funcName, err)
		return nil, err
	}

	marshaledData, err := json.Marshal(respArrData)
	if err != nil {
		m.log.Errorf("%s error: %v", funcName, err)
		return nil, err
	}
	return resultTransformFunc(marshaledData)
}

func MongoUpdateOne[S any, P any, R any](ctx context.Context, funcName string, m *mongoStorage,
	collName string, filterObj bson.D, payload *P, resultTransformFunc func(string) (*R, error),
) (*R, error) {
	marshaledJSONStr, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	convertedPayload := new(S)
	err = json.Unmarshal(marshaledJSONStr, convertedPayload)
	if err != nil {
		return nil, err
	}

	isUpsert := true
	updateResult, err := m.mgo.Collection(collName).UpdateOne(ctx, filterObj, convertedPayload, &options.UpdateOptions{
		Upsert: &isUpsert,
	})
	if err != nil {
		m.log.Errorf("%s error: %v", funcName, err)
		return nil, err
	}

	upsertedId, err := updateResult.UpsertedID.(primitive.ObjectID).MarshalText() //nolint:errcheck
	if err != nil {
		m.log.Errorf("%s error: %v", funcName, err)
		return nil, err
	}

	return resultTransformFunc(string(upsertedId))
}

func MongoDeleteOne[S any](ctx context.Context, m *mongoStorage, funcName string,
	collName string, filterObj bson.M, option *options.DeleteOptions,
) (*S, error) {
	_, err := m.mgo.Collection(collName).DeleteOne(ctx, filterObj, option)
	if err != nil {
		m.log.Errorf("%s, unable to delete one: %v", funcName, err)
		return nil, err
	}

	return new(S), nil
}

func MongoFind[S any](ctx context.Context, m *mongoStorage, funcName string,
	collName string, filterObj any, sortObj primitive.D,
) ([]*S, error) {
	findAllOptions := options.Find()
	findAllOptions.SetSort(sortObj)

	cursor, err := m.mgo.Collection(collName).Find(ctx, filterObj)
	if err != nil {
		m.log.Errorf("%s Find mongo error: %v", funcName, err)
		return nil, err
	}

	respArrData := []*S{}
	if err = cursor.All(ctx, &respArrData); err != nil {
		m.log.Errorf("%s mongodb cursor all error: %v", funcName, err)
		return nil, err
	}

	m.log.Infof("Done %s mongodb, data size:%d, data:%+v", funcName, len(respArrData), respArrData)
	return respArrData, nil
}
