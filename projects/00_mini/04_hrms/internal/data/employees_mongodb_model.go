package data

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MongoDbEmployee struct {
	ID     primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	Name   string               `json:"name"`
	Salary primitive.Decimal128 `json:"salary"`
	Age    int                  `json:"age"`
}
