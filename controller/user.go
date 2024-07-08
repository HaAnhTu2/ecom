package controller

import (
	"bytes"
	"encoding/json"
	"image-server/model"
	"image-server/reponsitory"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserController struct {
	UserRepo reponsitory.UserRepo
	DB       *mongo.Database
}

func NewUserController(UserRepo reponsitory.UserRepo, db *mongo.Database) *UserController {
	return &UserController{UserRepo: UserRepo,
		DB: db}
}
func (u *UserController) Login(c *gin.Context) {
	var auth model.LoginRequest
	if err := c.ShouldBind(&auth); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	user, err := u.UserRepo.FindByEmail(c.Request.Context(), auth.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid credentials",
		})
		return
	}
	if auth.Email == user.Email && auth.Password == user.Password {
		token, err := u.UserRepo.SaveToken(&user)
		log.Print(token)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}
		cookie := http.Cookie{}
		cookie.Name = "Token"
		cookie.Value = token
		cookie.Expires = time.Now().Add(15 * time.Minute)
		http.SetCookie(c.Writer, &cookie)
		c.JSON(http.StatusOK, gin.H{"token": token})
		log.Print(token)
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid credentials",
		})
	}
}

func (u *UserController) Logout(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:   "Token",
		Value:  "",
		Path:   "/",
		MaxAge: -1, // Delete the Cookie
	})
	c.JSON(http.StatusOK, gin.H{
		"data": "Logout successful!",
	})
}

func (u *UserController) GetAllUser(c *gin.Context) {
	users, err := u.UserRepo.GetAll(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"users": users,
	})
}

func (u *UserController) CreateUser(c *gin.Context) {
	user := model.User{
		Name:     c.Request.FormValue("name"),
		Email:    c.Request.FormValue("email"),
		Password: c.Request.FormValue("password"),
	}
	file, header, err := c.Request.FormFile("image")
	log.Print(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image upload failed"})
		return
	}
	defer file.Close()
	//Create GridFS bucket
	bucket, err := gridfs.NewBucket(u.DB.Client().Database("test31"), options.GridFSBucket().SetName("photos"))
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
	user.UserImage_URL = strings.Trim(string(fileId), `"`)

	// Insert the user into the database
	users, err := u.UserRepo.Create(c.Request.Context(), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not insert user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"fileId":   user.UserImage_URL,
		"fileSize": fileSize,
		"user":     users,
	})
}

func (u *UserController) ServeImage(c *gin.Context) {
	imageId := strings.TrimPrefix(c.Request.URL.Path, "/image/")
	objID, err := primitive.ObjectIDFromHex(imageId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image ID"})
		return
	}

	bucket, _ := gridfs.NewBucket(u.DB.Client().Database("test31"), options.GridFSBucket().SetName("photos"))

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

func (u *UserController) UpdateUser(c *gin.Context) {
	userId := c.Param("id")
	user, err := u.UserRepo.FindByID(c.Request.Context(), userId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if name := c.PostForm("name"); name != "" {
		user.Name = name
	}
	if email := c.PostForm("email"); email != "" {
		user.Email = email
	}
	if password := c.PostForm("password"); password != "" {
		user.Password = password
	}

	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
	}
	defer file.Close()
	bucket, err := gridfs.NewBucket(u.DB.Client().Database("test31"), options.GridFSBucket().SetName("photos"))
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
	user.UserImage_URL = strings.Trim(string(fileId), `"`)

	updatedUser, err := u.UserRepo.Update(c.Request.Context(), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not update user",
			"err":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": updatedUser,
	})
}

func (u *UserController) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid argument id"})
		return
	}
	if err := u.UserRepo.Delete(c, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
}
