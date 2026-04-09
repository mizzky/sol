package routes

import (
	"database/sql"
	"sol_coffeesys/backend/auth"
	"sol_coffeesys/backend/db"
	"sol_coffeesys/backend/handler"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, conn *sql.DB, queries *db.Queries) {
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

		api.PATCH("/users/:id/role", auth.AdminOnly(queries), handler.SetUserRoleHandler(queries))

		api.GET("/cart", auth.RequireAuth(queries), handler.GetCartHandler(queries))
		api.POST("/cart/items", auth.RequireAuth(queries), handler.AddToCartHandler(queries))
		api.PUT("/cart/items/:id", auth.RequireAuth(queries), handler.UpdateCartItemHandler(queries))
		api.DELETE("/cart/items/:id", auth.RequireAuth(queries), handler.RemoveCartItemHandler(queries))
		api.DELETE("/cart", auth.RequireAuth(queries), handler.ClearCartHandler(queries))

		api.GET("/me", handler.MeHandler(queries))

		api.GET("/orders", auth.RequireAuth(queries), handler.GetOrdersHandler(queries))
		api.POST("/orders", auth.RequireAuth(queries), handler.CreateOrderHandler(conn, queries))
		api.POST("/orders/:id/cancel", auth.RequireAuth(queries), handler.CancelOrderHandler(conn, queries))

		api.POST("/refresh", handler.RefreshTokenHandler(queries, tokenGenerator))
		api.POST("/logout", handler.LogoutHandler(queries))
	}
}
