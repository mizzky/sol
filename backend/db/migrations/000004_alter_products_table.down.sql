ALTER TABLE products
DROP CONSTRAINT IF EXISTS fk_products_category;

DROP INDEX IF EXISTS idx_products_sku;
DROP INDEX IF EXISTS idx_products_category_id;

ALTER TABLE products
DROP COLUMN IF EXISTS updated_at,
DROP COLUMN IF EXISTS created_at,
DROP COLUMN IF EXISTS stock_quantity,
DROP COLUMN IF EXISTS image_url,
DROP COLUMN IF EXISTS description,
DROP COLUMN IF EXISTS sku,
DROP COLUMN IF EXISTS category_id;