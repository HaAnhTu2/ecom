package route

import (
	"image-server/controller"
	"image-server/db"
	"image-server/middleware"
	"image-server/reponsitory"
	"os"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func Route(r *gin.Engine, DB *mongo.Database) {
	//All routes will be added here
	client := db.ConnectDB()
	ProductRepo := reponsitory.NewProductRepo(client.Database(os.Getenv("DB_NAME")))
	productController := controller.NewProductController(ProductRepo, DB)
	UserRepo := reponsitory.NewUserRepo(client.Database(os.Getenv("DB_NAME")))
	userController := controller.NewUserController(UserRepo, DB)
	authMiddleware := middleware.AuthMiddleware
	// r.Use(sessions.Sessions("session", cookie.NewStore([]byte(os.Getenv("SECRET_KEY")))))
	r.POST("api/login", userController.Login)
	r.DELETE("api/logout", userController.Logout)
	auth := r.Group("/")
	auth.Use(authMiddleware)
	{
		auth.POST("/api/user/create", userController.CreateUser)
		auth.PUT("/api/user/update/:id", userController.UpdateUser)
		auth.DELETE("/api/user/delete/:id", userController.DeleteUser)

		auth.POST("/api/product/create", productController.CreateProduct)
		auth.PUT("/api/product/update/:id", productController.UpdateProduct)
		auth.DELETE("/api/product/delete/:id", productController.DeleteProduct)
	}
	// r.POST("/api/user/create", userController.CreateUser)
	r.GET("/api/user/get", userController.GetAllUser)
	r.GET("image/:imageId", userController.ServeImage)
	r.GET("/api/product/get", productController.GetAllProduct)
	r.GET("image2/:imageId", productController.ServeImageProduct)
}
