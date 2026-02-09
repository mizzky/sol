package routes

import (
	"net/http"
	"sol_coffeesys/backend/auth"
	"sol_coffeesys/backend/db"
	"sol_coffeesys/backend/handler"
	"sol_coffeesys/backend/pkg/respond"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, queries *db.Queries) {
	api := r.Group("/api")
	tokenGenerator := auth.DefaultTokenGenerator{}
	{
		api.POST("/register", handler.RegisterUserHandler(queries))
		api.POST("/login", handler.LoginUserHandler(queries, tokenGenerator))

		api.POST("/cagtegories", auth.AdminOnly(queries), handler.CreateCategoryHandler(queries))
		api.PUT("/categories/:id", auth.AdminOnly(queries), handler.UpdateCategoryHandler(queries))
		api.GET("/categories", handler.GetCategoriesHandler(queries))
		api.DELETE("/categories/:id", auth.AdminOnly(queries), handler.DeleteCategoryHandler(queries))

		api.GET("/products", func(c *gin.Context) {
			products, err := queries.ListProducts(c.Request.Context())
			if err != nil {
				respond.RespondError(c, http.StatusInternalServerError, "予期せぬエラーが発生しました")
				return
			}
			c.JSON(http.StatusOK, products)
		})

		api.POST("/products", func(c *gin.Context) {
			var input struct {
				Name  string `json:"name"`
				Price int32  `json:"price"`
			}

			if err := c.ShouldBindJSON(&input); err != nil {
				respond.RespondError(c, http.StatusBadRequest, "リクエスト形式が正しくありません")
				return
			}
			product, err := queries.CreateProduct(c.Request.Context(), db.CreateProductParams{
				Name:        input.Name,
				Price:       input.Price,
				IsAvailable: true,
			})
			if err != nil {
				respond.RespondError(c, http.StatusInternalServerError, "予期せぬエラーが発生しました")
				return
			}
			c.JSON(http.StatusCreated, product)
		})

	}
}
