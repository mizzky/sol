package routes

import (
	"sol_coffeesys/backend/auth"
	"sol_coffeesys/backend/db"
	"sol_coffeesys/backend/handler"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, queries *db.Queries) {
	api := r.Group("/api")
	tokenGenerator := auth.DefaultTokenGenerator{}
	{
		api.POST("/register", handler.RegisterUserHandler(queries))
		api.POST("/login", handler.LoginUserHandler(queries, tokenGenerator))

		api.POST("/categories", auth.AdminOnly(queries), handler.CreateCategoryHandler(queries))
		api.PUT("/categories/:id", auth.AdminOnly(queries), handler.UpdateCategoryHandler(queries))
		api.DELETE("/categories/:id", auth.AdminOnly(queries), handler.DeleteCategoryHandler(queries))
		api.GET("/categories", handler.GetCategoriesHandler(queries))

		api.GET("/products", handler.ListProductsHandler(queries))
		api.GET("/products/:id", handler.GetProductHandler(queries))
		api.POST("/products", auth.AdminOnly(queries), handler.CreateProductHandler(queries))
		api.PUT("/products/:id", auth.AdminOnly(queries), handler.UpdateProductHandler(queries))
		api.DELETE("/products/:id", auth.AdminOnly(queries), handler.DeleteProductHandler(queries))
	}
}
