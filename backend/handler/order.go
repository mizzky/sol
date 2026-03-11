package handler

import (
	"context"
	"database/sql"
	"errors"
	"sol_coffeesys/backend/db"
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
			OrderID:   order.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			UnitPrice: int64(item.ProductPrice),
		})
		if err != nil {
			return nil, err
		}

		_, err = qtx.UpdateProductStock(ctx, db.UpdateProductStockParams{
			ID:            item.ProductID,
			StockQuantity: item.ProductStock - item.Quantity,
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
