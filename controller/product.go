package controller

import (
	"bytes"
	"encoding/json"
	"image-server/model"
	"image-server/reponsitory"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ProductController struct {
	ProductRepo reponsitory.ProductRepo
	DB          *mongo.Database
}

func NewProductController(ProductRepo reponsitory.ProductRepo, db *mongo.Database) *ProductController {
	return &ProductController{ProductRepo: ProductRepo, DB: db}
}

func (p *ProductController) GetAllProduct(c *gin.Context) {
	products, err := p.ProductRepo.GetAll(c)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"products": products,
	})
}

func (p *ProductController) CreateProduct(c *gin.Context) {
	product := model.Product{
		ProductName: c.Request.FormValue("productname"),
		Brand:       c.Request.FormValue("brand"),
		Description: c.Request.FormValue("description"),
	}
	quantity, err := strconv.Atoi(c.Request.FormValue("quantity"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quantity"})
		return
	}
	product.Quantity = quantity

	price, err := strconv.ParseFloat(c.Request.FormValue("price"), 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid price"})
		return
	}
	product.Price = price
	file, header, err := c.Request.FormFile("image2")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image upload failed"})
		log.Print(err)
		return
	}
	product.Created_At = time.Now()
	product.Updated_At = time.Now()
	defer file.Close()
	//Create GridFS bucket

	bucket, err := gridfs.NewBucket(p.DB.Client().Database(os.Getenv("DB_NAME")), options.GridFSBucket().SetName("products"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not create GridFS bucket"})
		return
	}
	//Read image
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not read image"})
		return
	}
	//Open upload stream
	filename := time.Now().Format(time.RFC3339) + "_" + header.Filename
	uploadStream, err := bucket.OpenUploadStream(filename)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not open upload stream"})
		return
	}
	defer uploadStream.Close()
	//Write to upload stream
	fileSize, err := uploadStream.Write(buf.Bytes())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not write to upload stream"})
		return
	}

	// Save the file ID to the user model
	fileId, err := json.Marshal(uploadStream.FileID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not marshal file ID"})
		return
	}
	product.ProductImage_URL = strings.Trim(string(fileId), `"`)
	// Insert the user into the database
	products, err := p.ProductRepo.Create(c.Request.Context(), product)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not insert user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"fileId":   product.ProductImage_URL,
		"fileSize": fileSize,
		"product":  products,
	})
}

func (p *ProductController) ServeImageProduct(c *gin.Context) {
	imageId := strings.TrimPrefix(c.Request.URL.Path, "/image2/")
	objID, err := primitive.ObjectIDFromHex(imageId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image ID"})
		return
	}

	bucket, _ := gridfs.NewBucket(p.DB.Client().Database(os.Getenv("DB_NAME")), options.GridFSBucket().SetName("products"))

	var buf bytes.Buffer
	_, err = bucket.DownloadToStream(objID, &buf)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not download image"})
		return
	}

	contentType := http.DetectContentType(buf.Bytes())
	c.Writer.Header().Add("Content-Type", contentType)
	c.Writer.Header().Add("Content-Length", strconv.Itoa(len(buf.Bytes())))
	c.Writer.Write(buf.Bytes())
}

func (p *ProductController) UpdateProduct(c *gin.Context) {
	productid := c.Param("id")
	product, err := p.ProductRepo.FindByID(c.Request.Context(), productid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}
	if productname := c.PostForm("productname"); productname != "" {
		product.ProductName = productname
	}
	if brand := c.PostForm("brand"); brand != "" {
		product.Brand = brand
	}
	if quantityStr := c.PostForm("quantity"); quantityStr != "" {
		quantity, err := strconv.Atoi(quantityStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid quantity",
			})
			return
		}
		product.Quantity = quantity
	}
	if prices := c.PostForm("price"); prices != "" {
		price, err := strconv.ParseFloat(prices, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid price",
			})
			return
		}
		product.Price = price
	}
	if description := c.PostForm("description"); description != "" {
		product.Description = description
	}
	file, header, err := c.Request.FormFile("image2")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
	}
	defer file.Close()
	bucket, err := gridfs.NewBucket(p.DB.Client().Database(os.Getenv("DB_NAME")), options.GridFSBucket().SetName("products"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not create GridFS bucket"})
		return
	}

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not read image"})
		return
	}

	filename := time.Now().Format(time.RFC3339) + "_" + header.Filename
	uploadStream, err := bucket.OpenUploadStream(filename)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not open upload stream"})
		return
	}
	defer uploadStream.Close()

	if _, err := uploadStream.Write(buf.Bytes()); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not write to upload stream"})
		return
	}

	fileId, err := json.Marshal(uploadStream.FileID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not marshal file ID"})
		return
	}
	log.Print(fileId)
	product.ProductImage_URL = strings.Trim(string(fileId), `"`)

	updatedProduct, err := p.ProductRepo.Update(c.Request.Context(), product)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not update product",
			"err":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"product": updatedProduct,
	})
}

func (p *ProductController) DeleteProduct(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid argument id",
		})
		return
	}
	if err := p.ProductRepo.Delete(c, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}
}
