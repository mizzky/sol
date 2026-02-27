package handler

import (
	"database/sql"
	"net/http"
	"sol_coffeesys/backend/db"
	"sol_coffeesys/backend/pkg/respond"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func GetCartHandler(q db.Querier) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, exists := c.Get("userID")
		if !exists {
			respond.RespondError(c, http.StatusUnauthorized, "認証が必要です")
			return
		}

		var userID int64
		switch v := raw.(type) {
		case int64:
			userID = v
		case int:
			userID = int64(v)
		case float64:
			userID = int64(v)
		default:
			respond.RespondError(c, http.StatusUnauthorized, "認証が必要です")
			return
		}

		rows, err := q.ListCartItemsByUser(c.Request.Context(), userID)
		if err != nil {
			respond.RespondError(c, http.StatusInternalServerError, "予期せぬエラーが発生しました")
			return
		}

		items := make([]gin.H, 0, len(rows))
		for _, it := range rows {
			items = append(items, gin.H{
				"id":            it.ID,
				"cart_id":       it.CartID,
				"product_id":    it.ProductID,
				"quantity":      it.Quantity,
				"price":         it.Price,
				"created_at":    it.CreatedAt.Format(time.RFC3339),
				"updated_at":    it.UpdatedAt.Format(time.RFC3339),
				"product_name":  it.ProductName,
				"product_price": it.ProductPrice,
				"product_stock": it.ProductStock,
			})
		}
		c.JSON(http.StatusOK, gin.H{"items": items})
	}
}

type addToCartRequest struct {
	ProductID int64 `json:"product_id"`
	Quantity  int32 `json:"quantity"`
}

func AddToCartHandler(q db.Querier) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, ok := c.Get("userID")
		if !ok {
			respond.RespondError(c, http.StatusUnauthorized, "認証が必要です")
			return
		}

		var req addToCartRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			respond.RespondError(c, http.StatusBadRequest, "リクエスト形式が正しくありません")
			return
		}
		if req.Quantity <= 0 {
			respond.RespondError(c, http.StatusBadRequest, "quantityは1以上である必要があります")
			return
		}

		var userID int64
		switch v := uid.(type) {
		case int64:
			userID = v
		case int:
			userID = int64(v)
		case float64:
			userID = int64(v)
		default:
			respond.RespondError(c, http.StatusUnauthorized, "認証が必要です")
			return
		}

		product, err := q.GetProduct(c.Request.Context(), req.ProductID)
		if err != nil {
			if err == sql.ErrNoRows {
				respond.RespondError(c, http.StatusNotFound, "商品が見つかりません")
			} else {
				respond.RespondError(c, http.StatusInternalServerError, "予期せぬエラーが発生しました")
			}
			return
		}

		cart, err := q.GetOrCreateCartForUser(c.Request.Context(), userID)
		if err != nil {
			respond.RespondError(c, http.StatusInternalServerError, "予期せぬエラーが発生しました")
			return
		}

		item, err := q.AddCartItem(c.Request.Context(), db.AddCartItemParams{
			CartID:    cart.ID,
			ProductID: req.ProductID,
			Quantity:  req.Quantity,
			Price:     int64(product.Price),
		})
		if err != nil {
			respond.RespondError(c, http.StatusInternalServerError, "予期せぬエラーが発生しました")
			return
		}

		c.JSON(http.StatusCreated, gin.H{"item": item})
	}
}

type updateCartItemRequest struct {
	Quantity int32 `json:"quantity"`
}

func UpdateCartItemHandler(q db.Querier) gin.HandlerFunc {
	return func(c *gin.Context) {
		// URL pramの解析
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			respond.RespondError(c, http.StatusBadRequest, "idが正しくありません")
			return
		}

		uid, ok := c.Get("userID")
		if !ok {
			respond.RespondError(c, http.StatusUnauthorized, "認証が必要です")
			return
		}
		var req updateCartItemRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			respond.RespondError(c, http.StatusBadRequest, "リクエスト形式が正しくありません")
			return
		}

		if req.Quantity <= 0 {
			respond.RespondError(c, http.StatusBadRequest, "quantitiyは１以上である必要があります")
			return
		}

		var userID int64
		switch v := uid.(type) {
		case int64:
			userID = v
		case int:
			userID = int64(v)
		case float64:
			userID = int64(v)
		default:
			respond.RespondError(c, http.StatusUnauthorized, "認証が必要です")
			return
		}

		item, err := q.UpdateCartItemQtyByUser(c.Request.Context(), db.UpdateCartItemQtyByUserParams{
			ID:       id,
			Quantity: req.Quantity,
			UserID:   userID,
		})
		if err != nil {
			if err == sql.ErrNoRows {
				respond.RespondError(c, http.StatusNotFound, "カートアイテムが見つかりません")
			} else {
				respond.RespondError(c, http.StatusInternalServerError, "予期せぬエラーが発生しました")
			}
			return
		}

		c.JSON(http.StatusOK, gin.H{"item": item})
	}
}

func RemoveCartItemHandler(q db.Querier) gin.HandlerFunc {
	return func(c *gin.Context) {

		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			respond.RespondError(c, http.StatusBadRequest, "idが正しくありません")
			return
		}

		uid, ok := c.Get("userID")
		if !ok {
			respond.RespondError(c, http.StatusUnauthorized, "認証が必要です")
			return
		}
		var userID int64
		switch v := uid.(type) {
		case int64:
			userID = v
		case int:
			userID = int64(v)
		case float64:
			userID = int64(v)
		default:
			respond.RespondError(c, http.StatusUnauthorized, "認証が必要です")
			return
		}

		item, err := q.GetCartItemByID(c.Request.Context(), id)
		if err != nil {
			if err == sql.ErrNoRows {
				respond.RespondError(c, http.StatusNotFound, "カートアイテムが見つかりません")
			} else {
				respond.RespondError(c, http.StatusInternalServerError, "予期せぬエラーが発生しました")
			}
			return
		}

		cart, err := q.GetCartByUser(c.Request.Context(), userID)
		if err != nil {
			if err == sql.ErrNoRows {
				respond.RespondError(c, http.StatusNotFound, "カートが見つかりません")
			} else {
				respond.RespondError(c, http.StatusInternalServerError, "予期せぬエラーが発生しました")
			}
			return
		}

		if item.CartID != cart.ID {
			respond.RespondError(c, http.StatusNotFound, "カートアイテムが見つかりません")
			return
		}

		if err := q.RemoveCartItemByUser(c.Request.Context(), db.RemoveCartItemByUserParams{
			ID:     id,
			UserID: userID,
		}); err != nil {
			respond.RespondError(c, http.StatusInternalServerError, "予期せぬエラーが発生しました")
			return
		}

		c.Status(http.StatusNoContent)

	}
}
