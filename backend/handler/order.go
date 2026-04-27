package handler

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"sol_coffeesys/backend/db"
	"sol_coffeesys/backend/pkg/apperror"
	"strconv"

	"github.com/gin-gonic/gin"
)

func createOrderLogic(ctx context.Context, qtx db.Querier, userID int64) (*db.CreateOrderRow, error) {
	// カートを取得
	_, err := qtx.GetOrCreateCartForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// カート内の商品取得
	items, err := qtx.ListCartItemsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// カートが空の場合はエラー
	if len(items) == 0 {
		return nil, apperror.NewValidationError("cart", nil, "", "")
	}

	// 合計金額計算
	var total int64
	for _, item := range items {
		total += item.Price
	}

	// 各商品の検証 - 在庫確認
	for _, item := range items {
		// 商品情報を取得
		product, err := qtx.GetProductForUpdate(ctx, item.ProductID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, apperror.NewNotFoundError("product", item.ProductID, apperror.NotFoundMessageProduct)
			}
			return nil, err
		}

		if product.StockQuantity < item.Quantity {
			return nil, apperror.NewConflictError("qty", fmt.Sprint(item.ProductID), "")
		}
	}

	// 注文レコード作成
	order, err := qtx.CreateOrder(ctx, db.CreateOrderParams{
		UserID: userID,
		Total:  total,
		Status: "pending",
	})
	if err != nil {
		return nil, err
	}

	// 各商品ループで注文明細作成と在庫更新
	for _, item := range items {
		_, err := qtx.CreateOrderItem(ctx, db.CreateOrderItemParams{
			OrderID:             order.ID,
			ProductID:           item.ProductID,
			Quantity:            item.Quantity,
			UnitPrice:           int64(item.ProductPrice),
			ProductNameSnapshot: item.ProductName,
		})
		if err != nil {
			return nil, err
		}

		_, err = qtx.UpdateProductStock(ctx, db.UpdateProductStockParams{
			ID:            item.ProductID,
			StockQuantity: -item.Quantity,
		})
		if err != nil {
			return nil, err
		}
	}

	err = qtx.ClearCartByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &order, nil
}

func CreateOrderHandler(conn *sql.DB, queries *db.Queries) gin.HandlerFunc {
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
			_ = c.Error(apperror.NewUnauthorizedError("userID_parse_failed", apperror.UnauthorizedMessageAuth))
			return
		}

		tx, err := conn.BeginTx(c.Request.Context(), nil)
		if err != nil {
			_ = c.Error(apperror.NewInternalError("BeginTx", err, apperror.InternalServerMessageCommon))
			return
		}

		qtx := queries.WithTx(tx)
		order, err := createOrderLogic(c.Request.Context(), qtx, userID)
		if err != nil {
			_ = tx.Rollback()

			var ve *apperror.ValidationError
			var ce *apperror.ConflictError
			var ne *apperror.NotFoundError
			var be *apperror.BusinessLogicError

			if errors.As(err, &ve) || errors.As(err, &ne) || errors.As(err, &ce) || errors.As(err, &be) {
				_ = c.Error(err)
				return
			}

			_ = c.Error(apperror.NewInternalError("CreateOrder", err, apperror.InternalServerMessageCommon))
			return
		}

		if err := tx.Commit(); err != nil {
			_ = c.Error(apperror.NewInternalError("Commit", err, apperror.InternalServerMessageCommon))
			return
		}
		c.JSON(http.StatusCreated, gin.H{"order": order})
	}
}

func cancelOrderLogic(ctx context.Context, qtx db.Querier, orderID int64, userID int64) (*db.UpdateOrderStatusRow, error) {
	ord, err := qtx.GetOrderByIDForUpdate(ctx, orderID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NewNotFoundError("order", orderID, "")
		}
		return nil, err
	}
	// 所有権チェック
	if ord.UserID != userID {
		return nil, apperror.NewNotFoundError("order", orderID, "")
	}

	if ord.Status != "pending" {
		return nil, apperror.NewBusinessLogicError("この注文はキャンセルできません")
	}

	items, err := qtx.ListOrderItemsByOrderID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	for _, it := range items {
		_, err := qtx.UpdateProductStock(ctx, db.UpdateProductStockParams{
			ID:            it.ProductID,
			StockQuantity: it.Quantity,
		})
		if err != nil {
			return nil, err
		}
	}

	updated, err := qtx.UpdateOrderStatus(ctx, db.UpdateOrderStatusParams{
		ID:     orderID,
		Status: "cancelled",
	})
	if err != nil {
		return nil, err
	}

	return &updated, nil
}

func CancelOrderHandler(conn *sql.DB, queries *db.Queries) gin.HandlerFunc {
	return func(c *gin.Context) {
		orderIDParam := c.Param("id")
		if orderIDParam == "" {
			_ = c.Error(apperror.NewValidationError("order", nil, "", apperror.ValidationMessageEssentialOrder))
			return
		}
		orderID, err := strconv.ParseInt(orderIDParam, 10, 64)
		if err != nil {
			_ = c.Error(apperror.NewValidationError("order", nil, "", apperror.ValidationMessageOrder))
			return
		}

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

		tx, err := conn.BeginTx(c.Request.Context(), nil)
		if err != nil {
			_ = c.Error(apperror.NewInternalError("BeginTx", err, apperror.InternalServerMessageCommon))
			return
		}

		qtx := queries.WithTx(tx)
		updated, err := cancelOrderLogic(c.Request.Context(), qtx, orderID, userID)
		if err != nil {
			_ = tx.Rollback()

			var ve *apperror.ValidationError
			var ce *apperror.ConflictError
			var ne *apperror.NotFoundError
			var be *apperror.BusinessLogicError

			if errors.As(err, &ve) || errors.As(err, &ne) || errors.As(err, &ce) || errors.As(err, &be) {
				_ = c.Error(err)
				return
			}
			_ = c.Error(apperror.NewInternalError("CancelOrder", err, apperror.InternalServerMessageCommon))
			return
		}

		if err := tx.Commit(); err != nil {
			_ = c.Error(apperror.NewInternalError("Commit", err, apperror.InternalServerMessageCommon))
			return
		}
		c.JSON(http.StatusOK, gin.H{"order": updated})
	}
}

type OrderWithItems struct {
	Order db.ListOrdersByUserRow `json:"order"`
	Items []db.OrderItem         `json:"items"`
}

func getOrderLogic(ctx context.Context, qtx db.Querier, userID int64) ([]OrderWithItems, error) {
	orders, err := qtx.ListOrdersByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	res := make([]OrderWithItems, 0, len(orders))
	for _, order := range orders {
		items, err := qtx.ListOrderItemsByOrderID(ctx, order.ID)
		if err != nil {
			return nil, err
		}
		res = append(res, OrderWithItems{
			Order: order,
			Items: items,
		})
	}
	return res, nil
}

var validOrderStatuses = map[string]struct{}{
	"pending":   {},
	"cancelled": {},
}

func isValidOrderStatus(status string) bool {
	if status == "" {
		return true
	}
	_, ok := validOrderStatuses[status]
	return ok
}

func GetOrdersHandler(queries db.Querier) gin.HandlerFunc {
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

		status := c.Query("status")
		if !isValidOrderStatus(status) {
			_ = c.Error(apperror.NewValidationError("status", status, "", ""))
			return
		}

		orders, err := getOrderLogic(c.Request.Context(), queries, userID)
		if err != nil {
			_ = c.Error(apperror.NewInternalError("GetOrders", err, apperror.InternalServerMessageCommon))
			return
		}

		var filtered []OrderWithItems
		if status == "" {
			filtered = orders
		} else {
			for _, order := range orders {
				if order.Order.Status == status {
					filtered = append(filtered, order)
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{"orders": filtered})
	}
}
