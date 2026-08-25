CREATE INDEX idx_users_email
ON users(email);

CREATE INDEX idx_categories_name
ON categories(name);

CREATE INDEX idx_products_sku
ON products(sku);

CREATE INDEX idx_carts_user_id
ON carts(user_id);

CREATE INDEX idx_cart_items_cart_id
ON cart_items(cart_id);