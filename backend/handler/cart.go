package handler

import (
	"net/http"
	"sol_coffeesys/backend/db"
	"sol_coffeesys/backend/pkg/respond"
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
