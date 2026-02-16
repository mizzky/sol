package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"sol_coffeesys/backend/db"
	"sol_coffeesys/backend/pkg/respond"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

type ProductResponse struct {
	ID            int64   `json:"id"`
	Name          string  `json:"name"`
	Price         int32   `json:"price"`
	IsAvailable   bool    `json:"is_available"`
	CategoryID    int64   `json:"category_id"`
	Sku           string  `json:"sku"`
	Description   *string `json:"description"`
	ImageUrl      *string `json:"image_url,omitempty"`
	StockQuantity int32   `json:"stock_quantity"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

type CreateProductHandlerRequest struct {
	Name          string  `json:"name"`
	Price         int32   `json:"price"`
	IsAvailable   bool    `json:"is_available"`
	CategoryID    int64   `json:"category_id"`
	Sku           string  `json:"sku"`
	Description   *string `json:"description"`
	ImageUrl      *string `json:"image_url,omitempty"`
	StockQuantity int32   `json:"stock_quantity"`
}

type UpdateProductHandlerRequest = CreateProductHandlerRequest

// ＋＋商品一覧取得機能＋＋
func ListProductsHandler(q db.Querier) gin.HandlerFunc {
	return func(c *gin.Context) {
		products, err := q.ListProducts(c.Request.Context())
		if err != nil {
			respond.RespondError(c, http.StatusInternalServerError, "予期せぬエラーが発生しました")
			return
		}
		resp := make([]ProductResponse, 0, len(products))
		for _, p := range products {
			var desc *string
			if p.Description.Valid {
				desc = &p.Description.String
			}
			var img *string
			if p.ImageUrl.Valid {
				img = &p.ImageUrl.String
			}
			resp = append(resp, ProductResponse{
				ID:            p.ID,
				Name:          p.Name,
				Price:         p.Price,
				IsAvailable:   p.IsAvailable,
				CategoryID:    p.CategoryID,
				Sku:           p.Sku,
				Description:   desc,
				ImageUrl:      img,
				StockQuantity: p.StockQuantity,
				CreatedAt:     p.CreatedAt.Format(time.RFC3339),
				UpdatedAt:     p.UpdatedAt.Format(time.RFC3339),
			})
		}
		c.JSON(http.StatusOK, gin.H{"products": resp})
	}
}

// ＋＋商品取得機能＋＋
func GetProductHandler(q db.Querier) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			respond.RespondError(c, http.StatusBadRequest, "IDが正しくありません")
			return
		}

		product, err := q.GetProduct(c.Request.Context(), int64(id))
		if err != nil {
			if err == sql.ErrNoRows {
				respond.RespondError(c, http.StatusNotFound, "商品が見つかりません")
			} else {
				respond.RespondError(c, http.StatusInternalServerError, "予期せぬエラーが発生しました")
			}
			return
		}
		var desc *string
		if product.Description.Valid {
			desc = &product.Description.String
		}
		var img *string
		if product.ImageUrl.Valid {
			img = &product.ImageUrl.String
		}
		c.JSON(http.StatusOK, ProductResponse{
			ID:            product.ID,
			Name:          product.Name,
			Price:         product.Price,
			IsAvailable:   product.IsAvailable,
			CategoryID:    product.CategoryID,
			Sku:           product.Sku,
			Description:   desc,
			ImageUrl:      img,
			StockQuantity: product.StockQuantity,
			CreatedAt:     product.CreatedAt.Format(time.RFC3339),
			UpdatedAt:     product.UpdatedAt.Format(time.RFC3339),
		})
	}
}

// ＋＋商品登録機能＋＋
func CreateProductHandler(q db.Querier) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateProductHandlerRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			respond.RespondError(c, http.StatusBadRequest, "リクエスト形式が正しくありません")
			return
		}

		if req.Name == "" {
			respond.RespondError(c, http.StatusBadRequest, "nameは必須です")
			return
		}
		if req.Price <= 0 {
			respond.RespondError(c, http.StatusBadRequest, "priceは正の整数である必要があります")
			return
		}
		if req.Sku == "" {
			respond.RespondError(c, http.StatusBadRequest, "skuは必須です")
			return
		}

		if len(req.Name) > 255 {
			respond.RespondError(c, http.StatusBadRequest, "nameは255文字以内である必要があります")
			return
		}

		if _, err := q.GetCategory(c.Request.Context(), req.CategoryID); err != nil {
			if err == sql.ErrNoRows {
				respond.RespondError(c, http.StatusNotFound, "カテゴリが見つかりません")
			} else {
				respond.RespondError(c, http.StatusInternalServerError, "予期せぬエラーが発生しました")
			}
			return
		}

		var description sql.NullString
		if req.Description != nil {
			description = sql.NullString{String: *req.Description, Valid: true}
		}
		var imageUrl sql.NullString
		if req.ImageUrl != nil {
			imageUrl = sql.NullString{String: *req.ImageUrl, Valid: true}
		}

		product, err := q.CreateProduct(c.Request.Context(), db.CreateProductParams{
			Name:          req.Name,
			Price:         req.Price,
			IsAvailable:   req.IsAvailable,
			CategoryID:    req.CategoryID,
			Sku:           req.Sku,
			Description:   description,
			ImageUrl:      imageUrl,
			StockQuantity: req.StockQuantity,
		})
		if err != nil {
			var pqErr *pq.Error
			if errors.As(err, &pqErr) && pqErr.Code == "23505" {
				respond.RespondError(c, http.StatusConflict, "SKUが既に存在します")
				return
			}
			respond.RespondError(c, http.StatusInternalServerError, "予期せぬエラーが発生しました")
			return
		}

		var respDesc *string
		if product.Description.Valid {
			respDesc = &product.Description.String
		}
		var respImg *string
		if product.ImageUrl.Valid {
			respImg = &product.ImageUrl.String
		}

		c.JSON(http.StatusCreated, ProductResponse{
			ID:            product.ID,
			Name:          product.Name,
			Price:         product.Price,
			IsAvailable:   product.IsAvailable,
			CategoryID:    product.CategoryID,
			Sku:           product.Sku,
			Description:   respDesc,
			ImageUrl:      respImg,
			StockQuantity: product.StockQuantity,
			CreatedAt:     product.CreatedAt.Format(time.RFC3339),
			UpdatedAt:     product.UpdatedAt.Format(time.RFC3339),
		})
	}
}

// ＋＋商品更新機能＋＋
func UpdateProductHandler(q db.Querier) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			respond.RespondError(c, http.StatusBadRequest, "IDが正しくありません")
			return
		}
		var req UpdateProductHandlerRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			respond.RespondError(c, http.StatusBadRequest, "リクエスト形式が正しくありません")
			return
		}

		if req.Name == "" {
			respond.RespondError(c, http.StatusBadRequest, "nameは必須です")
			return
		}
		if req.Price <= 0 {
			respond.RespondError(c, http.StatusBadRequest, "priceは正の整数である必要があります")
			return
		}
		if req.Sku == "" {
			respond.RespondError(c, http.StatusBadRequest, "skuは必須です")
			return
		}

		if len(req.Name) > 255 {
			respond.RespondError(c, http.StatusBadRequest, "nameは255文字以内である必要があります")
			return
		}

		// category 存在確認
		if _, err := q.GetCategory(c.Request.Context(), req.CategoryID); err != nil {
			if err == sql.ErrNoRows {
				respond.RespondError(c, http.StatusNotFound, "カテゴリが見つかりません")
			} else {
				respond.RespondError(c, http.StatusInternalServerError, "予期せぬエラーが発生しました")
			}
			return
		}

		var description sql.NullString
		if req.Description != nil {
			description = sql.NullString{String: *req.Description, Valid: true}
		}
		var imageUrl sql.NullString
		if req.ImageUrl != nil {
			imageUrl = sql.NullString{String: *req.ImageUrl, Valid: true}
		}

		product, err := q.UpdateProduct(c.Request.Context(), db.UpdateProductParams{
			Name:          req.Name,
			Price:         req.Price,
			IsAvailable:   req.IsAvailable,
			CategoryID:    req.CategoryID,
			Sku:           req.Sku,
			Description:   description,
			ImageUrl:      imageUrl,
			StockQuantity: req.StockQuantity,
			ID:            int64(id),
		})
		if err != nil {
			var pqErr *pq.Error
			if errors.As(err, &pqErr) && pqErr.Code == "23505" {
				respond.RespondError(c, http.StatusConflict, "SKUが既に存在します")
				return
			}

			if err == sql.ErrNoRows {
				respond.RespondError(c, http.StatusNotFound, "商品が見つかりません")
			} else {
				respond.RespondError(c, http.StatusInternalServerError, "予期せぬエラーが発生しました")
			}
			return
		}

		var respDesc *string
		if product.Description.Valid {
			respDesc = &product.Description.String
		}
		var respImg *string
		if product.ImageUrl.Valid {
			respImg = &product.ImageUrl.String
		}

		c.JSON(http.StatusOK, ProductResponse{
			ID:            product.ID,
			Name:          product.Name,
			Price:         product.Price,
			IsAvailable:   product.IsAvailable,
			CategoryID:    product.CategoryID,
			Sku:           product.Sku,
			Description:   respDesc,
			ImageUrl:      respImg,
			StockQuantity: product.StockQuantity,
			CreatedAt:     product.CreatedAt.Format(time.RFC3339),
			UpdatedAt:     product.UpdatedAt.Format(time.RFC3339),
		})
	}
}

// ＋＋商品削除機能＋＋
func DeleteProductHandler(q db.Querier) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			respond.RespondError(c, http.StatusBadRequest, "IDが正しくありません")
			return
		}
		err = q.DeleteProduct(c.Request.Context(), int64(id))
		if err != nil {
			if err == sql.ErrNoRows {
				respond.RespondError(c, http.StatusNotFound, "商品が見つかりません")
			} else {
				respond.RespondError(c, http.StatusInternalServerError, "予期せぬエラーが発生しました")
			}
			return
		}
		c.JSON(http.StatusNoContent, nil)
	}
}
