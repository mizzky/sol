package handler

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"sol_coffeesys/backend/db"
	"sol_coffeesys/backend/pkg/respond"

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
		return nil, errors.New("カートが空です")
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
				return nil, errors.New("商品が見つかりません")
			}
			return nil, err
		}

		if product.StockQuantity < item.Quantity {
			return nil, errors.New("在庫不足です")
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

		tx, err := conn.BeginTx(c.Request.Context(), nil)
		if err != nil {
			respond.RespondError(c, http.StatusInternalServerError, "予期せぬエラーが発生しました")
			return
		}

		qtx := queries.WithTx(tx)
		order, err := createOrderLogic(c.Request.Context(), qtx, userID)
		if err != nil {
			_ = tx.Rollback()

			switch {
			case err.Error() == "カートが空です":
				respond.RespondError(c, http.StatusBadRequest, "カートが空です")
			case err.Error() == "商品が見つかりません":
				respond.RespondError(c, http.StatusNotFound, "商品が見つかりません")
			case err.Error() == "在庫不足です":
				respond.RespondError(c, http.StatusConflict, "在庫不足です")
			default:
				respond.RespondError(c, http.StatusInternalServerError, "予期せぬエラーが発生しました")
			}
			return
		}

		if err := tx.Commit(); err != nil {
			respond.RespondError(c, http.StatusInternalServerError, "予期せぬエラーが発生しました")
			return
		}
		c.JSON(http.StatusCreated, gin.H{"order": order})
	}
}

func cancelOrderLogic(ctx context.Context, qtx db.Querier, orderID int64, userID int64) (*db.UpdateOrderStatusRow, error) {
	return nil, nil
}

func CancelOrderHandler(conn *sql.DB, queries *db.Queries) gin.HandlerFunc {
	return func(c *gin.Context) {
	}
}
