package handler

import (
	"database/sql"
	"net/http"
	"sol_coffeesys/backend/db"

	"github.com/gin-gonic/gin"
)

type CreateCategoryRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
}

func CreateCategory(queries db.Querier) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateCategoryRequest

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

		c.JSON(http.StatusCreated, category)
	}
}
