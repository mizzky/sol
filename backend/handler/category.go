package handler

import (
	"database/sql"
	"net/http"
	"sol_coffeesys/backend/db"
	"sol_coffeesys/backend/pkg/apperror"
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
			_ = c.Error(apperror.NewValidationError("request", nil, "", ""))
			return
		}

		if req.Name == "" {
			_ = c.Error(apperror.NewValidationError("category", nil, "", ""))
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
			_ = c.Error(apperror.NewInternalError("CreateCategory", err, apperror.InternalServerMessageCommon))
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

		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			_ = c.Error(apperror.NewValidationError("id", id, "", ""))
			return
		}

		var req UpdateCategoryHandlerRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			_ = c.Error(apperror.NewValidationError("request", nil, "bind", apperror.ValidationMessageRequest))
			return
		}

		if req.Name == nil || *req.Name == "" {
			_ = c.Error(apperror.NewValidationError("category", nil, "", ""))
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
				_ = c.Error(apperror.NewNotFoundError("category", id, ""))
			} else {
				_ = c.Error(apperror.NewInternalError("UpdateCategory", err, apperror.InternalServerMessageCommon))
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
			_ = c.Error(apperror.NewInternalError("ListCategories", err, apperror.InternalServerMessageCommon))
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
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			_ = c.Error(apperror.NewValidationError("id", id, "", ""))
			return
		}

		respErr := queries.DeleteCategory(c.Request.Context(), int64(id))
		if respErr != nil {
			if respErr == sql.ErrNoRows {
				_ = c.Error(apperror.NewNotFoundError("category", id, ""))
			} else {
				_ = c.Error(apperror.NewInternalError("DeleteCategory", respErr, apperror.InternalServerMessageCommon))
			}
			return
		}

		c.JSON(http.StatusNoContent, nil)
	}
}
