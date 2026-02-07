package handler

import (
	"database/sql"
	"net/http"
	"sol_coffeesys/backend/db"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ＋＋カテゴリー登録機能＋＋
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

// ＋＋カテゴリー更新機能＋＋
type UpdateCategoryHandlerRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

func UpdateCategoryHandler(queries db.Querier) gin.HandlerFunc {

	return func(c *gin.Context) {

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			RespondError(c, http.StatusBadRequest, "IDが正しくありません")
			return
		}

		var req UpdateCategoryHandlerRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			RespondError(c, http.StatusBadRequest, "リクエスト形式が正しくありません")
			return
		}

		if req.Name == nil || *req.Name == "" {
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

		category, err := queries.UpdateCategory(c.Request.Context(), db.UpdateCategoryParams{
			ID:          int64(id),
			Name:        *req.Name,
			Description: description,
		})
		if err != nil {
			if err == sql.ErrNoRows {
				RespondError(c, http.StatusNotFound, "カテゴリが見つかりません")
			} else {
				RespondError(c, http.StatusInternalServerError, "予期せぬエラーが発生しました")
			}
			return
		}

		var responseDescription *string
		if category.Description.Valid {
			responseDescription = &category.Description.String
		}
		c.JSON(http.StatusOK, CategoryResponse{
			ID:          category.ID,
			Name:        category.Name,
			Description: responseDescription,
		})

	}
}

// ＋＋カテゴリー一覧取得機能＋＋
func GetCategoriesHandler(queries db.Querier) gin.HandlerFunc {
	return func(c *gin.Context) {
		categories, err := queries.ListCategories(c.Request.Context())
		if err != nil {
			RespondError(c, http.StatusInternalServerError, "予期せぬエラーが発生しました")
			return
		}
		resp := make([]CategoryResponse, 0, len(categories))
		for _, cat := range categories {
			var desc *string
			if cat.Description.Valid {
				desc = &cat.Description.String
			}
			resp = append(resp, CategoryResponse{
				ID:          cat.ID,
				Name:        cat.Name,
				Description: desc,
			})
		}
		c.JSON(http.StatusOK, gin.H{"categories": resp})
	}
}

// ＋＋カテゴリー削除機能＋＋
func DeleteCategoryHandler(queries db.Querier) gin.HandlerFunc {
	return nil
}
