\set ON_ERROR_STOP on

\ir ../seed/00_guard.sql
\ir ../seed/00_profile.sql

DO $verify$
DECLARE
    actual_products BIGINT;
    min_product_id BIGINT;
    max_product_id BIGINT;
    invalid_products BIGINT;
    duplicate_skus BIGINT;
    orphan_categories BIGINT;

    available_products BIGINT;
    unavailable_products BIGINT;

    zero_stock_products BIGINT;
    low_stock_products BIGINT;
    normal_stock_products BIGINT;
    high_stock_products BIGINT;

    invalid_category_distribution BIGINT;

    expected_products BIGINT;
    expected_available_products BIGINT;
    expected_unavailable_products BIGINT;
    expected_zero_stock BIGINT;
    expected_low_stock BIGINT;
    expected_normal_stock BIGINT;
    expected_high_stock BIGINT;


    products_sequence_value BIGINT;
    products_sequence_called BOOLEAN;
BEGIN
    SELECT products_count
    INTO expected_products
    FROM pg_temp.perf_profile;

    expected_available_products := expected_products * 19 / 20;
    expected_unavailable_products := expected_products / 20;
    expected_zero_stock  := expected_products / 10;
    expected_low_stock := expected_products * 2 / 10;
    expected_normal_stock := expected_products * 6 / 10;
    expected_high_stock := expected_products / 10;


    SELECT count(*), min(id), max(id)
    INTO actual_products, min_product_id, max_product_id
    FROM public.products;

    IF actual_products <> expected_products
       OR min_product_id IS DISTINCT FROM 1
       OR max_product_id IS DISTINCT FROM expected_products THEN
        RAISE EXCEPTION
            'products range mismatch: expected count/min/max %/1/%, actual %/%/%',
            expected_products,
            expected_products,
            actual_products,
            min_product_id,
            max_product_id;
    END IF;

    SELECT count(*)
    INTO invalid_products
    FROM public.products
    WHERE name IS DISTINCT FROM
              format('product-%s', lpad(id::text, 7, '0'))
       OR price IS DISTINCT FROM
              100 + ((id - 1) % 1000) * 10
       OR is_available IS DISTINCT FROM
              (id % 20 <> 0)
       OR category_id IS DISTINCT FROM
              CASE
                  WHEN (id - 1) % 100 < 40 THEN 1
                  WHEN (id - 1) % 100 < 60 THEN 2
                  ELSE 3 + ((((id - 1) % 100) - 60) / 5)
              END
       OR sku IS DISTINCT FROM
              format('PERF-%s', lpad(id::text, 7, '0'))
       OR description IS DISTINCT FROM
              format(
                  'Performance product %s',
                  lpad(id::text, 7, '0')
              )
       OR image_url IS DISTINCT FROM
              format(
                  'https://example.test/products/%s.jpg',
                  lpad(id::text, 7, '0')
              )
       OR stock_quantity IS DISTINCT FROM
              CASE
                  WHEN ((id - 1) / 20) % 10 = 0
                      THEN 0
                  WHEN ((id - 1) / 20) % 10 IN (1, 2)
                      THEN 1 + ((id - 1) % 5)
                  WHEN ((id - 1) / 20) % 10 BETWEEN 3 AND 8
                      THEN 10 + ((id - 1) % 91)
                  WHEN ((id - 1) / 20) % 10 = 9
                      THEN 500 + ((id - 1) % 501)
              END
       OR created_at IS DISTINCT FROM
              TIMESTAMPTZ '2025-01-01 00:00:00+00'
       OR updated_at IS DISTINCT FROM
              TIMESTAMPTZ '2025-01-01 00:00:00+00';

    IF invalid_products <> 0 THEN
        RAISE EXCEPTION
            'invalid products found: %',
            invalid_products;
    END IF;

    SELECT
        count(*) FILTER (WHERE is_available),
        count(*) FILTER (WHERE NOT is_available)
    INTO
        available_products,
        unavailable_products
    FROM public.products;

    IF available_products <> expected_available_products
       OR unavailable_products <> expected_unavailable_products THEN
        RAISE EXCEPTION
            'availability mismatch: expected true/false %/%, actual %/%',
            expected_available_products,
            expected_unavailable_products,
            available_products,
            unavailable_products;
    END IF;

    SELECT
        count(*) FILTER (WHERE stock_quantity = 0),
        count(*) FILTER (WHERE stock_quantity BETWEEN 1 AND 5),
        count(*) FILTER (WHERE stock_quantity BETWEEN 10 AND 100),
        count(*) FILTER (WHERE stock_quantity BETWEEN 500 AND 1000)
    INTO
        zero_stock_products,
        low_stock_products,
        normal_stock_products,
        high_stock_products
    FROM public.products;

    IF zero_stock_products <> expected_zero_stock
       OR low_stock_products <> expected_low_stock
       OR normal_stock_products <> expected_normal_stock
       OR high_stock_products <> expected_high_stock THEN
        RAISE EXCEPTION
            'stock distribution mismatch: expected %/%/%/%, actual %/%/%/%',
            expected_zero_stock,
            expected_low_stock,
            expected_normal_stock,
            expected_high_stock,
            zero_stock_products,
            low_stock_products,
            normal_stock_products,
            high_stock_products;
    END IF;

    SELECT count(*)
    INTO invalid_category_distribution
    FROM (
        SELECT category_id, count(*) AS actual_count
        FROM public.products
        GROUP BY category_id
    ) AS actual
    FULL JOIN (
        VALUES
            (1::BIGINT, expected_products * 40 / 100),
            (2::BIGINT, expected_products * 20 / 100),
            (3::BIGINT, expected_products * 5 / 100),
            (4::BIGINT, expected_products * 5 / 100),
            (5::BIGINT, expected_products * 5 / 100),
            (6::BIGINT, expected_products * 5 / 100),
            (7::BIGINT, expected_products * 5 / 100),
            (8::BIGINT, expected_products * 5 / 100),
            (9::BIGINT, expected_products * 5 / 100),
            (10::BIGINT, expected_products * 5 / 100)
    ) AS expected(category_id, expected_count)
        USING (category_id)
    WHERE actual.actual_count
          IS DISTINCT FROM expected.expected_count;

    IF invalid_category_distribution <> 0 THEN
        RAISE EXCEPTION
            'category distribution mismatch: mismatched categories %',
            invalid_category_distribution;
    END IF;

    SELECT count(*) - count(DISTINCT sku)
    INTO duplicate_skus
    FROM public.products;

    IF duplicate_skus <> 0 THEN
        RAISE EXCEPTION
            'duplicate product SKUs found: %',
            duplicate_skus;
    END IF;

    SELECT count(*)
    INTO orphan_categories
    FROM public.products AS products
    LEFT JOIN public.categories AS categories
        ON categories.id = products.category_id
    WHERE categories.id IS NULL;

    IF orphan_categories <> 0 THEN
        RAISE EXCEPTION
            'orphan product categories found: %',
            orphan_categories;
    END IF;

    SELECT last_value, is_called
    INTO products_sequence_value, products_sequence_called
    FROM public.products_id_seq;

    IF products_sequence_value <> expected_products
       OR NOT products_sequence_called THEN
        RAISE EXCEPTION
            'products sequence mismatch: value %, called %',
            products_sequence_value,
            products_sequence_called;
    END IF;
END
$verify$;

\echo 'PASS: products'
