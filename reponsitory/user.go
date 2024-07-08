package reponsitory

import (
	"context"
	"errors"
	"image-server/model"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepo interface {
	FindByID(ctx context.Context, id string) (model.User, error)
	GetByID(ctx context.Context, ID primitive.ObjectID) (model.User, error)
	FindByEmail(ctx context.Context, email string) (model.User, error)
	GetAll(ctx context.Context) ([]model.UserResponse, error)
	Create(ctx context.Context, user model.User) (model.User, error)
	Update(ctx context.Context, user model.User) (model.User, error)
	Delete(ctx context.Context, id string) error
	SaveToken(user *model.User) (string, error)
}
type UserRepoI struct {
	db *mongo.Database
}

func NewUserRepo(db *mongo.Database) UserRepo {
	return &UserRepoI{db: db}
}
func (u *UserRepoI) GetByID(ctx context.Context, ID primitive.ObjectID) (model.User, error) {
	var user model.User
	err := u.db.Collection("users").FindOne(ctx, bson.M{"_id": ID}).Decode(&user)
	if err != nil {
		return model.User{}, err
	}
	return user, nil
}
func (u *UserRepoI) FindByID(ctx context.Context, id string) (model.User, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return model.User{}, errors.New("invalid user ID")
	}

	var user model.User
	err = u.db.Collection("users").FindOne(ctx, bson.M{"_id": objID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return model.User{}, errors.New("user not found")
		}
		return model.User{}, err
	}
	return user, nil
}
func (u *UserRepoI) FindByEmail(ctx context.Context, email string) (model.User, error) {
	var user model.User
	err := u.db.Collection("users").FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return model.User{}, errors.New("user not found")
		}
		return model.User{}, err
	}
	return user, nil
}
func (u *UserRepoI) GetAll(ctx context.Context) ([]model.UserResponse, error) {
	var users []model.UserResponse
	var items []model.User
	r, err := u.db.Collection("users").Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	if er := r.All(context.TODO(), &items); er != nil {
		return nil, er
	}

	for _, item := range items {
		users = append(users, model.UserResponse{
			Id:        item.ID.Hex(),
			Name:      item.Name,
			Email:     item.Email,
			Image_URL: item.UserImage_URL,
		})
	}
	return users, nil
}
func (u *UserRepoI) Create(ctx context.Context, user model.User) (model.User, error) {
	result, err := u.db.Collection("users").InsertOne(ctx, user)
	if err != nil {
		return model.User{}, err
	}
	user.ID = result.InsertedID.(primitive.ObjectID)
	return user, nil
}
func (u *UserRepoI) Update(ctx context.Context, user model.User) (model.User, error) {
	result, err := u.db.Collection("users").UpdateOne(ctx, bson.M{"_id": user.ID}, bson.M{
		"$set": bson.M{
			"name":          user.Name,
			"email":         user.Email,
			"password":      user.Password,
			"userimage_url": user.UserImage_URL,
		}})
	if err != nil {
		return model.User{}, err
	}
	if result.MatchedCount == 0 {
		return model.User{}, mongo.ErrNoDocuments
	}
	return model.User{}, nil
}
func (u *UserRepoI) Delete(ctx context.Context, id string) error {
	ID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	result, err := u.db.Collection("users").DeleteOne(ctx, bson.M{"_id": ID})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return err
	}
	return nil
}
func (u *UserRepoI) SaveToken(user *model.User) (string, error) {
	secret := os.Getenv("SECRET_KEY")
	expired_At := time.Now().Add(15 * time.Minute)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.MapClaims{
		"sub": user.Email,
		"exp": expired_At.Unix(),
	})

	return token.SignedString([]byte(secret))
}
