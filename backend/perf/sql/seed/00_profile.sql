\set ON_ERROR_STOP on

-- profile未指定の場合、既存のsmall用コマンドとの互換性を保つ
\if :{?profile}
\else
\set profile small
\endif

-- 選択されたprofileを一時表の1行として保持
CREATE TEMP TABLE perf_profile AS
SELECT
    requested.profile,
    profiles.users_count,
    profiles.products_count,
    profiles.carts_count,
    profiles.orders_count
FROM (
    SELECT :'profile'::TEXT AS profile
) AS requested
LEFT JOIN (
    VALUES
        ('small'::TEXT, 1000::BIGINT, 10000::BIGINT, 500::BIGINT, 10000::BIGINT),
        ('medium'::TEXT, 10000::BIGINT, 100000::BIGINT, 5000::BIGINT, 100000::BIGINT),
        ('large'::TEXT, 100000::BIGINT, 1000000::BIGINT, 50000::BIGINT, 1000000::BIGINT)
) AS profiles(
    profile,
    users_count,
    products_count,
    carts_count,
    orders_count
)
ON profiles.profile = requested.profile;

DO $profile$

DECLARE
    requested_profile TEXT;
    selected_users_count BIGINT;
BEGIN
    SELECT 
        profile,
        users_count
    INTO
        requested_profile,
        selected_users_count
    FROM pg_temp.perf_profile;

    IF selected_users_count IS NULL THEN
        RAISE EXCEPTION
            'invalid performance profile: %, expected small, medium or large',
            requested_profile;
    END IF;
END;
$profile$;