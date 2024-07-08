package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password,omitempty"`
}

type User struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name          string             `bson:"name" json:"name"`
	Email         string             `bson:"email,unique" json:"email"`
	Password      string             `bson:"password" json:"password"`
	UserImage_URL string             `bson:"userimage_url" json:"userimage_url"`
}

type UserResponse struct {
	Id        string `json:"_id,omitempty" bson:"_id,omitempty"`
	Name      string `json:"name,omitempty" bson:"name,omitempty"`
	Email     string `json:"email,omitempty" bson:"email,unique"`
	Password  string `json:"password,omitempty" bson:"password,omitempty"`
	Image_URL string `json:"userimage_url,omitempty" bson:"userimage_url,omitempty"`
}

type Token struct {
	ID               primitive.ObjectID `bson:"_id"`
	AccessToken      string             `bson:"access_token"`
	RefreshToken     string             `bson:"refresh_token"`
	TokenType        string             `bson:"token_type"`
	UserID           primitive.ObjectID `bson:"user_id"`
	Expired_At       time.Time          `bson:"expired_at"`
	RefreshExpiredAt time.Time          `bson:"refresh_expired_at"`
	Created_At       time.Time          `bson:"created_at"`
}
