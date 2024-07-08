package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Product struct {
	ID               primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	ProductName      string             `json:"productname" bson:"productname"`
	Brand            string             `json:"brand" bson:"brand"`
	Quantity         int                `json:"quantity" bson:"quantity"`
	Price            float64            `json:"price" bson:"price"`
	ProductImage_URL string             `json:"productimage_url" bson:"productimage_url"`
	Description      string             `json:"description" bson:"description"`
	Created_At       time.Time          `json:"created_at" bson:"created_at"`
	Updated_At       time.Time          `json:"updated_at" bson:"updated_at"`
}

type ProductResponse struct {
	ID               string  `json:"_id,omitempty" bson:"_id,omitempty"`
	ProductName      string  `json:"productname" bson:"productname"`
	Brand            string  `json:"brand" bson:"brand"`
	Quantity         int     `json:"quantity" bson:"quantity"`
	Price            float64 `json:"price" bson:"price"`
	ProductImage_URL string  `json:"productimage_url" bson:"productimage_url"`
	Description      string  `json:"description" bson:"description"`
}
