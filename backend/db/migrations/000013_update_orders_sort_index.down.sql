DROP INDEX IF EXISTS idx_orders_user_id_created_at;

CREATE INDEX idx_orders_user_id_created_at
ON orders(user_id, created_at DESC);
