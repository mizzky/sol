\set ON_ERROR_STOP on

\ir ../seed/00_profile.sql

DO $verify$

DECLARE
    actual_profile TEXT;
    actual_users BIGINT;
    actual_products BIGINT;
    actual_carts BIGINT;
    actual_orders BIGINT;

    expected_users BIGINT;
    expected_products BIGINT;
    expected_carts BIGINT;
    expected_orders BIGINT;

BEGIN
    SELECT
        profile,
        users_count,
        products_count,
        carts_count,
        orders_count
    INTO
        actual_profile,
        actual_users,
        actual_products,
        actual_carts,
        actual_orders
    FROM pg_temp.perf_profile;

    CASE actual_profile
        WHEN 'small' THEN
            expected_users := 1000;
            expected_products := 10000;
            expected_carts := 500;
            expected_orders := 10000;
        WHEN 'medium' THEN
            expected_users := 10000;
            expected_products := 100000;
            expected_carts := 5000;
            expected_orders := 100000;
        WHEN 'large' THEN
            expected_users := 100000;
            expected_products := 1000000;
            expected_carts := 50000;
            expected_orders := 1000000;
        ELSE
            RAISE EXCEPTION
                'unexpected profile resolve: %',
                actual_profile;
    END CASE;

    IF actual_users <> expected_users
        OR actual_products <> expected_products
        OR actual_carts <> expected_carts
        OR actual_orders <> expected_orders THEN
        RAISE EXCEPTION
            'profile config mismatch for %: expected users/products/carts/orders %/%/%/%, actual %/%/%/%',
            actual_profile,
            expected_users,
            expected_products,
            expected_carts,
            expected_orders,
            actual_users,
            actual_products,
            actual_carts,
            actual_orders;
    END IF;
END;
$verify$;

\echo 'PASS: profile'

-- profile切り替えの検証
--
-- profile=small のとき:
--   users=1,000、products=10,000、carts=500、orders=10,000 を選ぶ。


-- profile=medium のとき:
--   users=10,000、products=100,000、carts=5,000、orders=100,000 を選ぶ。
--
-- profile=large のとき:
--   users=100,000、products=1,000,000、carts=50,000、orders=1,000,000 を選ぶ。
--
-- profileを未指定で実行したとき:
--   既存のsmall用コマンドとの互換性のため、smallを既定として選ぶ。
--
-- profile=foo のような不正値を指定したとき:
--   usersなどの業務テーブルを変更する前に、profile名を含むエラーで停止する。
--
-- categoriesは全profileで10件とする。
-- cart_itemsとorder_itemsの件数は、各profileのcarts/orders件数と既存の分布規則から導出する。
--
-- seedとverifyはpsqlの別接続で実行されるため、
-- 両方が同じprofile解決処理を読み込める必要がある。