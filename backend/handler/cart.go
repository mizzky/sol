package handler

import (
	"database/sql"
	"net/http"
	"sol_coffeesys/backend/db"
	"sol_coffeesys/backend/pkg/apperror"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func GetCartHandler(q db.Querier) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, exists := c.Get("userID")
		if !exists {
			_ = c.Error(apperror.NewUnauthorizedError("", apperror.UnauthorizedMessageAuth))
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
			_ = c.Error(apperror.NewUnauthorizedError("", apperror.UnauthorizedMessageAuth))
			return
		}

		rows, err := q.ListCartItemsByUser(c.Request.Context(), userID)
		if err != nil {
			_ = c.Error(apperror.NewInternalError("ListCartItemsByUser", err, apperror.InternalServerMessageCommon))
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
			_ = c.Error(apperror.NewUnauthorizedError("", apperror.UnauthorizedMessageAuth))
			return
		}

		var req addToCartRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			_ = c.Error(apperror.NewValidationError("request", nil, "", ""))
			return
		}
		if req.Quantity <= 0 {
			_ = c.Error(apperror.NewValidationError("qty", req.Quantity, "", ""))
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
			_ = c.Error(apperror.NewUnauthorizedError("", apperror.UnauthorizedMessageAuth))
			return
		}

		product, err := q.GetProduct(c.Request.Context(), req.ProductID)
		if err != nil {
			if err == sql.ErrNoRows {
				_ = c.Error(apperror.NewNotFoundError("product", req.ProductID, ""))
			} else {
				_ = c.Error(apperror.NewInternalError("GetProduct", err, apperror.InternalServerMessageCommon))
			}
			return
		}

		cart, err := q.GetOrCreateCartForUser(c.Request.Context(), userID)
		if err != nil {
			_ = c.Error(apperror.NewInternalError("GetOrCreateCartForUser", err, apperror.InternalServerMessageCommon))
			return
		}

		item, err := q.AddCartItem(c.Request.Context(), db.AddCartItemParams{
			CartID:    cart.ID,
			ProductID: req.ProductID,
			Quantity:  req.Quantity,
			Price:     int64(product.Price),
		})
		if err != nil {
			_ = c.Error(apperror.NewInternalError("AddCartItem", err, apperror.InternalServerMessageCommon))
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
			_ = c.Error(apperror.NewValidationError("id", id, "", ""))
			return
		}

		uid, ok := c.Get("userID")
		if !ok {
			_ = c.Error(apperror.NewUnauthorizedError("", apperror.UnauthorizedMessageAuth))
			return
		}
		var req updateCartItemRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			_ = c.Error(apperror.NewValidationError("request", nil, "", ""))
			return
		}

		if req.Quantity <= 0 {
			_ = c.Error(apperror.NewValidationError("qty", req.Quantity, "", ""))
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
			_ = c.Error(apperror.NewUnauthorizedError("", apperror.UnauthorizedMessageAuth))
			return
		}

		item, err := q.UpdateCartItemQtyByUser(c.Request.Context(), db.UpdateCartItemQtyByUserParams{
			ID:       id,
			Quantity: req.Quantity,
			UserID:   userID,
		})
		if err != nil {
			if err == sql.ErrNoRows {
				_ = c.Error(apperror.NewNotFoundError("cart_item", id, ""))
			} else {
				_ = c.Error(apperror.NewInternalError("UpdateCartItemQtyByUser", err, apperror.InternalServerMessageCommon))
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
			_ = c.Error(apperror.NewValidationError("id", id, "", ""))
			return
		}

		uid, ok := c.Get("userID")
		if !ok {
			_ = c.Error(apperror.NewUnauthorizedError("", apperror.UnauthorizedMessageAuth))
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
			_ = c.Error(apperror.NewUnauthorizedError("", apperror.UnauthorizedMessageAuth))
			return
		}

		item, err := q.GetCartItemByID(c.Request.Context(), id)
		if err != nil {
			if err == sql.ErrNoRows {
				_ = c.Error(apperror.NewNotFoundError("cart_item", id, ""))
			} else {
				_ = c.Error(apperror.NewInternalError("GetCartItemByID", err, apperror.InternalServerMessageCommon))
			}
			return
		}

		cart, err := q.GetCartByUser(c.Request.Context(), userID)
		if err != nil {
			if err == sql.ErrNoRows {
				_ = c.Error(apperror.NewNotFoundError("cart", userID, ""))
			} else {
				_ = c.Error(apperror.NewInternalError("GetCartByUser", err, apperror.InternalServerMessageCommon))
			}
			return
		}

		if item.CartID != cart.ID {
			_ = c.Error(apperror.NewNotFoundError("cart_item", id, ""))
			return
		}

		if err := q.RemoveCartItemByUser(c.Request.Context(), db.RemoveCartItemByUserParams{
			ID:     id,
			UserID: userID,
		}); err != nil {
			_ = c.Error(apperror.NewInternalError("RemoveCartItemByUser", err, apperror.InternalServerMessageCommon))
			return
		}

		c.Status(http.StatusNoContent)

	}
}

func ClearCartHandler(q db.Querier) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, ok := c.Get("userID")
		if !ok {
			_ = c.Error(apperror.NewUnauthorizedError("", apperror.UnauthorizedMessageAuth))
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
			_ = c.Error(apperror.NewUnauthorizedError("", apperror.UnauthorizedMessageAuth))
			return
		}

		if err := q.ClearCartByUser(c.Request.Context(), userID); err != nil {
			_ = c.Error(apperror.NewInternalError("ClearCartByUser", err, apperror.InternalServerMessageCommon))
			return
		}

		c.Status(http.StatusNoContent)
	}
}
