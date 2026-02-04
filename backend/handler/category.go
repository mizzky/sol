package handler

import (
	"database/sql"
	"net/http"
	"sol_coffeesys/backend/db"

	"github.com/gin-gonic/gin"
)

type CreateCategoryHandlerRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type CategoryResponse struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

func CreateCategoryHandler(queries db.Querier) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateCategoryHandlerRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			RespondError(c, http.StatusBadRequest, "リクエスト形式が正しくありません")
			return
		}

		if req.Name == "" {
			RespondError(c, http.StatusBadRequest, "カテゴリ名は必須です")
			return
		}

		var description sql.NullString
		if req.Description != nil {
			description = sql.NullString{
				String: *req.Description,
				Valid:  true,
			}
		}

		category, err := queries.CreateCategory(c.Request.Context(), db.CreateCategoryParams{
			Name:        req.Name,
			Description: description,
		})
		if err != nil {
			RespondError(c, http.StatusInternalServerError, "予期せぬエラーが発生しました")
			return
		}

		var responseDescription *string
		if category.Description.Valid {
			responseDescription = &category.Description.String
		}
		c.JSON(http.StatusCreated, CategoryResponse{
			ID:          category.ID,
			Name:        category.Name,
			Description: responseDescription,
		})
	}
}

func UpdateCategory(queries db.Querier) gin.HandlerFunc { return nil }
