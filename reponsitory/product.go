package reponsitory

import (
	"context"
	"image-server/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ProductRepo interface {
	FindByID(ctx context.Context, id string) (model.Product, error)
	GetAll(ctx context.Context) ([]model.ProductResponse, error)
	Create(ctx context.Context, product model.Product) (model.Product, error)
	Update(ctx context.Context, product model.Product) (model.Product, error)
	Delete(ctx context.Context, id string) error
}

type ProductRepoI struct {
	DB *mongo.Database
}

func NewProductRepo(DB *mongo.Database) ProductRepo {
	return &ProductRepoI{DB: DB}
}

func (p *ProductRepoI) GetAll(ctx context.Context) ([]model.ProductResponse, error) {
	var products []model.ProductResponse
	var items []model.Product
	result, err := p.DB.Collection("products").Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	if err := result.All(context.Background(), &items); err != nil {
		return nil, err
	}
	for _, item := range items {
		products = append(products, model.ProductResponse{
			ID:               item.ID.Hex(),
			ProductName:      item.ProductName,
			Brand:            item.Brand,
			Quantity:         item.Quantity,
			Price:            item.Price,
			ProductImage_URL: item.ProductImage_URL,
			Description:      item.Description,
		})
	}
	return products, nil
}

func (p *ProductRepoI) FindByID(ctx context.Context, id string) (model.Product, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return model.Product{}, err
	}
	var product model.Product
	err = p.DB.Collection("products").FindOne(ctx, bson.M{"_id": objID}).Decode(&product)
	if err != nil {
		return model.Product{}, err
	}
	return product, nil
}

func (p *ProductRepoI) Create(ctx context.Context, product model.Product) (model.Product, error) {
	result, err := p.DB.Collection("products").InsertOne(ctx, product)
	if err != nil {
		return model.Product{}, err
	}

	product.ID = result.InsertedID.(primitive.ObjectID)
	return product, nil
}

func (p *ProductRepoI) Update(ctx context.Context, product model.Product) (model.Product, error) {
	result, err := p.DB.Collection("products").UpdateOne(ctx, bson.M{"_id": product.ID}, bson.M{
		"$set": bson.M{
			"productname":      product.ProductName,
			"brand":            product.Brand,
			"quantity":         product.Quantity,
			"price":            product.Price,
			"productimage_url": product.ProductImage_URL,
			"description":      product.Description,
		}})
	if err != nil {
		return model.Product{}, err
	}
	if result.MatchedCount == 0 {
		return model.Product{}, mongo.ErrNoDocuments
	}
	return model.Product{}, nil
}

func (p *ProductRepoI) Delete(ctx context.Context, id string) error {
	ID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	result, err := p.DB.Collection("products").DeleteOne(ctx, bson.M{"_id": ID})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return err
	}
	return nil
}
