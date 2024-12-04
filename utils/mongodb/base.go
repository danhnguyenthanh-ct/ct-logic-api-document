package mongodb

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cast"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/carousell/ct-go/pkg/container"
	"github.com/carousell/ct-go/pkg/logger"
	logctx "github.com/carousell/ct-go/pkg/logger/log_context"

	"github.com/carousell/ct-core-uni-free-premium-service/config"
)

func ConnectDatabase(ctx context.Context, cfg *config.Config) (*mongo.Database, error) {
	connectionStr := cfg.Mongo.ConnectionString
	poolSize := cfg.Mongo.PoolSize

	opts := options.Client().ApplyURI(connectionStr).SetMaxPoolSize(poolSize)
	if cfg.Mongo.Debug {
		AddDebugOption(opts, logger.MustNamed("test"))
	}

	mongoClient, err := mongo.NewClient(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to initiate mongo: %w", err)
	}

	if err := mongoClient.Connect(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect mongo: %w", err)
	}

	if err := mongoClient.Ping(ctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("failed to ping mongoClient: %w", err)
	}

	return mongoClient.Database(cfg.Mongo.DBName), nil
}

func AddDebugOption(opts *options.ClientOptions, lg *logger.Logger) {
	cmd := sync.Map{}
	cmdMonitor := &event.CommandMonitor{
		Started: func(_ context.Context, evt *event.CommandStartedEvent) {
			cmd.Store(evt.RequestID, evt.Command.String())
		},
		Succeeded: func(_ context.Context, evt *event.CommandSucceededEvent) {
			r, ok := cmd.LoadAndDelete(evt.RequestID)
			if ok {
				rawCommand := cast.ToString(r)
				lg.Info(evt.DurationNanos%1000, " ms ", rawCommand)
			}
		},
		Failed: func(_ context.Context, evt *event.CommandFailedEvent) {
			r, ok := cmd.LoadAndDelete(evt.RequestID)
			if ok {
				rawCommand := cast.ToString(r)
				lg.Info(evt.DurationNanos%1000, " ms ", rawCommand)
			}
		},
	}
	opts.SetMonitor(cmdMonitor)
}

type BaseCollection[P any, T IEntity[P]] struct {
	collection     *mongo.Collection
	collectionName string
}

func NewBaseCollection[P any, T IEntity[P]](db *mongo.Database,
	collectionName string,
) *BaseCollection[P, T] {
	collection := db.Collection(collectionName)
	return &BaseCollection[P, T]{
		collection:     collection,
		collectionName: collectionName,
	}
}

func (col *BaseCollection[P, T]) Get(ctx context.Context, filter any, opts ...*options.FindOneOptions) (T, error) {
	res := new(T)
	err := col.collection.FindOne(ctx, filter, opts...).Decode(res)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		logctx.Errorf(ctx, "unable to get one from collection %s, error: %v", col.collection.Name(), err)
		return nil, fmt.Errorf("mongo get one: %w", err)
	}

	resMap, err := container.ToMapFromStruct(res)
	if err != nil {
		return nil, err
	}

	if deletedAt, ok := resMap["deleted_at"]; ok {
		if deletedAt != nil {
			return nil, nil
		}
	}

	return *res, nil
}

func (col *BaseCollection[P, T]) GetById(ctx context.Context, id any) (T, error) {
	ret, err := col.Get(ctx, primitive.M{
		"_id": id,
	})
	if err != nil {
		logctx.Errorf(ctx, "unable to get object from collection %s, error: %v", col.collection.Name(), err)
		return nil, err
	}

	if ret != nil && ret.GetDeletedAt() != nil {
		return nil, nil
	}

	return ret, nil
}

func (col *BaseCollection[P, T]) GetByBatch(ctx context.Context, filter any,
	sort primitive.D, limit int64, offset int64,
) ([]T, error) {
	opts := make([]*options.FindOptions, 0)
	if sort != nil {
		opts = append(opts, options.Find().SetSort(sort))
	}
	if limit > 0 {
		opts = append(opts, options.Find().SetLimit(limit))
	}
	if offset > 0 {
		opts = append(opts, options.Find().SetSkip(offset))
	}

	mapFilter := structToMap(filter, true)
	mapFilter["deleted_at"] = bson.M{"$exists": false}
	cursor, err := col.collection.Find(ctx, mapFilter, opts...)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}

		logctx.Errorf(ctx, "unable to get all from collection %s, err: %v", col.collection.Name(), err)
		return nil, fmt.Errorf("mongo get all: %w", err)
	}
	ret := make([]T, cursor.RemainingBatchLength())
	if err := cursor.All(ctx, &ret); err != nil {
		logctx.Errorf(ctx, "unable to parse result from %s, err: %v", col.collection.Name(), err)
		return nil, fmt.Errorf("mongo get_all parse result: %w", err)
	}

	return ret, nil
}

func (col *BaseCollection[P, T]) Insert(ctx context.Context, item T) error {
	item.SetCreatedAt(time.Now())
	ret, err := col.collection.InsertOne(ctx, item)
	if err != nil {
		logctx.Errorf(ctx, "mongo insert one, collection: %s, err: %v", col.collection.Name(), err)
		return err
	}
	return item.SetObjectID(ret.InsertedID)
}

func (col *BaseCollection[P, T]) InsertMany(ctx context.Context, items []T) error {
	interfaceItems := make([]any, len(items))
	for i, v := range items {
		v.SetCreatedAt(time.Now())
		interfaceItems[i] = v
	}
	rets, err := col.collection.InsertMany(ctx, interfaceItems)
	if err != nil {
		logctx.Errorf(ctx, "mongo insert one, collection: %s, err: %v", col.collection.Name(), err)
		return err
	}
	for i := range items {
		err := items[i].SetObjectID(rets.InsertedIDs[i])
		if err != nil {
			return err
		}
	}
	return nil
}

// NOTED: nested struct won't be converted to Object
func (col *BaseCollection[P, T]) Update(ctx context.Context, filter any, item any) (int64, error) {
	params := structToMap(item, false)
	params = params.Except([]string{"_id", "id", "created_at"})
	params = params.Merge(container.Map{
		"updated_at": time.Now(),
	})
	result, err := col.collection.UpdateOne(ctx, filter, primitive.M{
		"$set": primitive.M(params),
	})
	if err != nil {
		logctx.Errorf(ctx, "mongo update, collection: %s, err: %v", col.collection.Name(), err)
		return 0, fmt.Errorf("mongo update, %w", err)
	}
	return result.ModifiedCount, nil
}

func (col *BaseCollection[P, T]) UpdateMany(ctx context.Context, filter any, item any) (int64, error) {
	params := structToMap(item, false)
	params = params.Except([]string{"_id", "id", "created_at"})
	params = params.Merge(container.Map{
		"updated_at": time.Now(),
	})
	result, err := col.collection.UpdateMany(ctx, filter, primitive.M{
		"$set": primitive.M(params),
	})
	if err != nil {
		logctx.Errorf(ctx, "mongo update, collection: %s, err: %v", col.collection.Name(), err)
		return 0, fmt.Errorf("mongo update, %w", err)
	}
	return result.ModifiedCount, nil
}

func (col *BaseCollection[P, T]) UpdatePartial(ctx context.Context,
	filter any, params container.Map,
) (int64, error) {
	params = params.Except([]string{"_id", "id", "created_at"})
	params = params.Merge(container.Map{
		"updated_at": time.Now(),
	})
	result, err := col.collection.UpdateOne(ctx, filter, primitive.M{
		"$set": primitive.M(params),
	})
	if err != nil {
		logctx.Errorf(ctx, "mongo update, collection: %s, err: %v", col.collection.Name(), err)
		return 0, fmt.Errorf("mongo update, %w", err)
	}
	return result.ModifiedCount, nil
}

// NOTED: nested struct won't be converted to Object
func (col *BaseCollection[P, T]) Upsert(ctx context.Context,
	filter any, item T) (modifiedCount int64,
	upsertedCount int64, err error,
) {
	params := structToMap(item, false)
	params = params.Except([]string{"_id", "id", "created_at"})
	params = params.Merge(container.Map{
		"updated_at": time.Now(),
	})
	isUpsert := true
	updated, err := col.collection.UpdateOne(ctx, filter, primitive.M{
		"$set":         primitive.M(params),
		"$setOnInsert": primitive.M{"created_at": time.Now()},
	}, &options.UpdateOptions{
		Upsert: &isUpsert,
	})
	if err != nil {
		logctx.Errorf(ctx, "mongo upsert, collection: %s, err: %v", col.collection.Name(), err)
		return 0, 0, err
	}

	if updated.UpsertedID == nil {
		return updated.MatchedCount, 0, nil
	}

	// if upserted, update ObjectID
	return updated.ModifiedCount, updated.UpsertedCount, item.SetObjectID(updated.UpsertedID)
}

// update item with ObjectID = id, with all fields from item
func (col *BaseCollection[P, T]) UpdateById(ctx context.Context,
	id any, item T,
) (int64, error) {
	return col.Update(ctx, primitive.M{
		"_id": id,
	}, item)
}

// update item with ObjectID = id, with all fields from item
func (col *BaseCollection[P, T]) UpsertById(ctx context.Context,
	id any, item T,
) (modifiedCount, upsertedCount int64, err error) {
	return col.Upsert(ctx, primitive.M{
		"_id": id,
	}, item)
}

func (col *BaseCollection[P, T]) Delete(ctx context.Context,
	filter bson.M, option *options.DeleteOptions,
) error {
	_, err := col.collection.DeleteOne(ctx, filter, option)
	if err != nil {
		logctx.Errorf(ctx, "mongo delete, collection: %s, err: %v", col.collection.Name(), err)
		return fmt.Errorf("mongo delete: %w", err)
	}

	return nil
}

func (col *BaseCollection[P, T]) SoftDelete(ctx context.Context,
	filter bson.M,
) (int64, error) {
	return col.UpdateMany(ctx, filter, container.Map{
		"deleted_at": time.Now(),
	})
}

func (col *BaseCollection[P, T]) DeleteById(ctx context.Context, id any) error {
	return col.Delete(ctx, primitive.M{"_id": id}, nil)
}

func (col *BaseCollection[P, T]) SortDeleteById(ctx context.Context, id any) (int64, error) {
	return col.Update(ctx, primitive.M{
		"_id": id,
	}, container.Map{
		"deleted_at": time.Now(),
	})
}

func (col *BaseCollection[P, T]) CountByFilter(ctx context.Context, filter any) (int64, error) {
	return col.collection.CountDocuments(ctx, filter)
}

func structToMap(item any, isIgnoreEmpty bool) container.Map {
	if fmt.Sprintf("%T", item) == fmt.Sprintf("%T", container.Map{}) {
		itemInMap, ok := item.(container.Map)
		if !ok {
			return container.Map{}
		}
		return itemInMap
	}
	res := container.Map{}
	if item == nil {
		return res
	}
	v := reflect.TypeOf(item)
	reflectValue := reflect.ValueOf(item)
	reflectValue = reflect.Indirect(reflectValue)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	for i := 0; i < v.NumField(); i++ {
		tagStr := v.Field(i).Tag.Get("bson")
		tags := strings.Split(tagStr, ",")
		isOmitEmpty := len(tags) > 1 && tags[1] == "omitempty" && isIgnoreEmpty
		val := reflectValue.Field(i).Interface()
		if val == nil {
			continue
		}
		isDefaultValue := reflect.DeepEqual(
			reflectValue.Field(i).Interface(),
			reflect.Zero(reflect.TypeOf(val)).Interface(),
		)
		if isOmitEmpty && isDefaultValue {
			continue
		}
		if tags[0] != "" && tags[0] != "-" {
			field := reflectValue.Field(i).Interface()
			res[tags[0]] = field
		}
	}
	return res
}
