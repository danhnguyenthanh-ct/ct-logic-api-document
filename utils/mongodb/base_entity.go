package mongodb

import (
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IEntity[T any] interface {
	SetObjectID(rawID any) error // to convert ID generated to ID in entity
	GetObjectID() any            // get back the ObjectID stored into MongoDB

	SetUpdatedAt(time.Time)
	SetCreatedAt(time.Time)
	SetDeletedAt(time.Time)
	GetDeletedAt() *time.Time
	*T
}

type BaseEntity struct {
	Id        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	CreatedAt *time.Time         `json:"created_at,omitempty" bson:"created_at,omitempty"`
	UpdatedAt *time.Time         `json:"updated_at,omitempty" bson:"updated_at,omitempty"`
	DeletedAt *time.Time         `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
}

func (e *BaseEntity) SetObjectID(id any) error {
	v, ok := id.(primitive.ObjectID)
	if !ok {
		return errors.New("invalid objectID type")
	}
	e.Id = v
	return nil
}

func (e *BaseEntity) GetObjectID() any {
	return e.Id
}

func (e *BaseEntity) GetIdStr() string {
	return e.Id.Hex()
}

func (e *BaseEntity) SetUpdatedAt(t time.Time) {
	e.UpdatedAt = &t
}

func (e *BaseEntity) SetCreatedAt(t time.Time) {
	if e.CreatedAt != nil {
		return
	}
	e.CreatedAt = &t
	e.UpdatedAt = &t
}

func (e *BaseEntity) SetDeletedAt(t time.Time) {
	e.DeletedAt = &t
}

func (e *BaseEntity) GetDeletedAt() *time.Time {
	return e.DeletedAt
}
